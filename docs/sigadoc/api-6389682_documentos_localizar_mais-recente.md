# /documentos/localizar/mais-recente

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/documentos/localizar/mais-recente:
    get:
      summary: /documentos/localizar/mais-recente
      deprecated: false
      description: Exibe dados do móbil mais recente que atende às condições desejadas.
      tags:
        - SIGADOC/SIGAEX/consulta
        - consulta
      parameters:
        - name: modelo
          in: query
          description: Nome do modelo
          required: false
          schema:
            type: string
        - name: ano
          in: query
          description: Ano de produção do documento
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
        '404':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties: {}
          headers: {}
          x-apidog-name: Nenhum documento foi encontrado com os argumentos informados
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389682-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
