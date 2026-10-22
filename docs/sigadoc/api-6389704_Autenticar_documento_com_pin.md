# Autenticar documento com pin

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/documentos/{sigla}/autenticar-com-pin:
    post:
      summary: Autenticar documento com pin
      deprecated: false
      description: Não é necessário informar novamente a senha.
      tags:
        - SIGADOC/SIGAEX/acao
        - acao
      parameters:
        - name: sigla
          in: path
          description: Sigla do mobil
          required: true
          example: '{{DOCUMENTO_SIGLA_FINALIZADO_SEM_FORMATACAO}}'
          schema:
            type: string
        - name: x-road-client
          in: header
          description: ''
          required: true
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
          application/json:
            schema:
              type: object
              properties: {}
              x-apidog-orders: []
            example:
              pin: 41545454
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  sigla:
                    description: Sigla do documento assinado
                    type: string
                  status:
                    type: string
                    default: OK
                x-apidog-orders:
                  - sigla
                  - status
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
      security:
        - bearer: []
      x-apidog-folder: SIGADOC/SIGAEX/acao
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389704-run
components:
  schemas: {}
  securitySchemes:
    bearer:
      type: http
      scheme: bearer
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
