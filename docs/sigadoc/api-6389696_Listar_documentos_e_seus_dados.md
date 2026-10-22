# Listar documentos e seus dados

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/bi/listar:
    get:
      summary: Listar documentos e seus dados
      deprecated: false
      description: ''
      tags:
        - SIGADOC/SIGAEX/bi
        - bi
      parameters:
        - name: modelo
          in: query
          description: Nome do modelo
          required: false
          schema:
            type: string
        - name: lotaSubscritor
          in: query
          description: Sigla da lotação do subscritor
          required: false
          schema:
            type: string
        - name: marcador
          in: query
          description: Nome do marcador (situacao)
          required: false
          schema:
            type: string
        - name: dataMin
          in: query
          description: Data mínima documento (formato aaaa-mm-dd, inclusive)
          required: false
          schema:
            type: string
        - name: dataMax
          in: query
          description: Data máxima do documento (exclusive)
          required: false
          schema:
            type: string
        - name: campos
          in: query
          description: Campos (default é todos)
          required: false
          schema:
            type: string
        - name: delimitador
          in: query
          description: Delimitador (default é vírgula)
          required: false
          schema:
            type: string
        - name: quebraDeLinha
          in: query
          description: Quebra de Linha (default é CRLF)
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
                type: string
                format: binary
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
      x-apidog-folder: SIGADOC/SIGAEX/bi
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389696-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
