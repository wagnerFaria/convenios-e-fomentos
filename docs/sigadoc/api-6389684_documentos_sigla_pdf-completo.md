# /documentos/{sigla}/pdf-completo

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/documentos/{sigla}/pdf-completo:
    get:
      summary: /documentos/{sigla}/pdf-completo
      deprecated: false
      description: Obter o PDF completo de um documento para visualização
      tags:
        - SIGADOC/SIGAEX/consulta
        - consulta
      parameters:
        - name: sigla
          in: path
          description: Sigla do mobil
          required: true
          example: '{{SIGLA_DOCUMENTO}}'
          schema:
            type: string
        - name: x-road-client
          in: header
          description: ''
          required: false
          example: '{{X_ROAD_CLIENT}}'
          schema:
            type: string
        - name: Authorization
          in: header
          description: ''
          required: false
          example: Bearer {{TOKEN_AUTH}}
          schema:
            type: string
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  jwt:
                    type: string
                    description: JWT com as informações para download do PDF
                x-apidog-orders:
                  - jwt
          headers: {}
          x-apidog-name: Successful response
        '500':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties: {}
          headers: {}
          x-apidog-name: Error ocurred
      security: []
      x-apidog-folder: SIGADOC/SIGAEX/consulta
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389684-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
