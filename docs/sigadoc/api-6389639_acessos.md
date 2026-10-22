# /acessos

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/acessos:
    get:
      summary: /acessos
      deprecated: false
      description: Listar os últimos acessos
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
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  list:
                    type: array
                    items:
                      $ref: '#/components/schemas/AcessoItem'
                x-apidog-orders:
                  - list
                x-apidog-ignore-properties: []
          headers: {}
          x-apidog-name: Successful response
        '500':
          description: ''
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
          headers: {}
          x-apidog-name: Error ocurred
      security: []
      x-apidog-folder: SIGADOC/SIGA/utils
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389639-run
components:
  schemas:
    AcessoItem:
      description: Lista acessos disponíveis
      type: object
      properties:
        idAcesso:
          type: string
        nome:
          type: string
      x-apidog-orders:
        - idAcesso
        - nome
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
    Error:
      type: object
      properties:
        errormsg:
          type: string
      x-apidog-orders:
        - errormsg
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
