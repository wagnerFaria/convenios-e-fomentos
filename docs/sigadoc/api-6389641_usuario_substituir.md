# /usuario/substituir

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/usuario/substituir:
    post:
      summary: /usuario/substituir
      deprecated: false
      description: Enviar sugestões
      tags:
        - SIGADOC/SIGA/utils
        - utils
      parameters:
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
      requestBody:
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              properties:
                substituicaoId:
                  description: Id da substituição para ser efetivada
                  example: ''
                  type: string
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
                x-apidog-orders:
                  - status
          headers: {}
          x-apidog-name: Successful response
      security: []
      x-apidog-folder: SIGADOC/SIGA/utils
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389641-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
