# Obtem o status de uma operação assíncrona

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/status/{chave}:
    get:
      summary: Obtem o status de uma operação assíncrona
      deprecated: false
      description: ''
      tags:
        - SIGADOC/SIGAEX/download
        - download
      parameters:
        - name: chave
          in: path
          description: Chave que referencia um processo assíncrono
          required: true
          example: '{{UUID_DOWNLOAD}}'
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
                  mensagem:
                    type: string
                    description: Mensagem
                  indice:
                    type: number
                    description: Quantos passos já foram executados
                  contador:
                    type: number
                    description: Número total de passos a serem executados
                  bytes:
                    type: number
                    description: Tamanho total do payload até o momento
                  errormsg:
                    type: string
                    description: Mensagem de erro se houver
                  stacktrace:
                    type: string
                    description: Detalhamento do erro
                x-apidog-orders:
                  - mensagem
                  - indice
                  - contador
                  - bytes
                  - errormsg
                  - stacktrace
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
      x-apidog-folder: SIGADOC/SIGAEX/download
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389680-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
