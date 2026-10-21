package enduser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
)

func TestRenderTemplateEscapaHTML(t *testing.T) {
	tpl := db.EmailTemplate{Subject: "Olá {{.User.Username}}", BodyHtml: `<p>{{.User.Username}}</p><a href="{{.URL}}">x</a>`}
	subject, body, err := renderTemplate(tpl, TemplateData{
		User: TemplateUser{Username: `<script>alert(1)</script>`, Email: "a@b.com"},
		URL:  "https://exemplo.com/a?x=1&y=2",
	})
	require.NoError(t, err)
	assert.Equal(t, "Olá <script>alert(1)</script>", subject, "assunto é texto puro")
	assert.NotContains(t, body, "<script>")
	assert.Contains(t, body, "&lt;script&gt;")
	assert.Contains(t, body, `href="https://exemplo.com/a?x=1&amp;y=2"`)

	// javascript: no href é neutralizado pelo html/template.
	_, body, err = renderTemplate(db.EmailTemplate{Subject: "s", BodyHtml: `<a href="{{.URL}}">x</a>`}, TemplateData{URL: "javascript:alert(1)"})
	require.NoError(t, err)
	assert.NotContains(t, body, "javascript:alert")
}

func TestRenderTemplateErros(t *testing.T) {
	for name, tpl := range map[string]db.EmailTemplate{
		"assunto com sintaxe inválida": {Subject: "{{", BodyHtml: "x"},
		"assunto com variável":         {Subject: "{{.Nada}}", BodyHtml: "x"},
		"corpo com sintaxe inválida":   {Subject: "s", BodyHtml: "{{.User.Username"},
		"corpo com variável":           {Subject: "s", BodyHtml: "{{.Nada}}"},
	} {
		_, _, err := renderTemplate(tpl, TemplateData{})
		assert.Error(t, err, name)
	}
}

func TestValidURL(t *testing.T) {
	for in, want := range map[string]bool{"": true, "https://a.com/x": true, "http://localhost:3000": true, "javascript:alert(1)": false, "/relativa": false, "ftp://a.com": false, "https://": false} {
		assert.Equal(t, want, validURL(in), in)
	}
}
