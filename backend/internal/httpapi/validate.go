package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = func() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	// Usa o nome JSON do campo nas mensagens.
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" || name == "" {
			return f.Name
		}
		return name
	})
	_ = v.RegisterValidation("username", func(fl validator.FieldLevel) bool { return usernameRe.MatchString(fl.Field().String()) })
	return v
}()

var usernameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// FieldError descreve uma falha de validação de um campo.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Decode lê o corpo JSON (rejeitando campos desconhecidos) e valida a struct.
func Decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return &APIError{Status: 413, Code: "payload_too_large", Message: "O corpo da requisição excede o limite permitido."}
		}
		return ErrBadRequest("Corpo da requisição inválido: " + err.Error())
	}
	return Validate(dst)
}

// Validate valida uma struct e devolve um erro 422 com mensagens em português.
func Validate(v any) error {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}
	var ves validator.ValidationErrors
	if !errors.As(err, &ves) {
		return err
	}
	fields := make([]FieldError, 0, len(ves))
	for _, fe := range ves {
		fields = append(fields, FieldError{Field: fe.Field(), Message: messageFor(fe)})
	}
	return ErrValidation(map[string]any{"fields": fields})
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "Campo obrigatório."
	case "email":
		return "E-mail inválido."
	case "min":
		return fmt.Sprintf("Deve ter no mínimo %s.", fe.Param())
	case "max":
		return fmt.Sprintf("Deve ter no máximo %s.", fe.Param())
	case "oneof":
		return fmt.Sprintf("Deve ser um dos valores: %s.", fe.Param())
	case "username":
		return "Use apenas letras, números, ponto, hífen e sublinhado (começando por letra ou número)."
	case "uuid":
		return "UUID inválido."
	default:
		return "Valor inválido."
	}
}
