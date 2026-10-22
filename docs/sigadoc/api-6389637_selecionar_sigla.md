# /selecionar/{sigla}

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/selecionar/{sigla}:
    get:
      summary: /selecionar/{sigla}
      deprecated: false
      description: Selecionar uma entidade pela sigla
      tags:
        - SIGADOC/SIGA/utils
        - utils
      parameters:
        - name: sigla
          in: path
          description: Sigla da entidade
          required: true
          example: ''
          schema:
            type: string
        - name: matricula
          in: query
          description: Sigla completa da matrícula
          required: false
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
                  entidade:
                    $ref: '#/components/schemas/EntidadeItem'
                x-apidog-orders:
                  - entidade
                x-apidog-ignore-properties: []
          headers: {}
          x-apidog-name: Successful response
      security: []
      x-apidog-folder: SIGADOC/SIGA/utils
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389637-run
components:
  schemas:
    EntidadeItem:
      description: Contem dados de uma entidade que foi localizada pela sigla
      type: object
      properties:
        modulo:
          type: string
        tipo:
          type: string
        refId:
          type: string
        movId:
          type: string
        codigo:
          type: string
        sigla:
          type: string
        descricao:
          type: string
      x-apidog-orders:
        - modulo
        - tipo
        - refId
        - movId
        - codigo
        - sigla
        - descricao
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
