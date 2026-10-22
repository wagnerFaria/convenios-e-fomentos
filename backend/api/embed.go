// Pacote api embute a especificação OpenAPI no binário.
package api

import _ "embed"

// OpenAPI contém o conteúdo de openapi.yaml.
//
//go:embed openapi.yaml
var OpenAPI []byte
