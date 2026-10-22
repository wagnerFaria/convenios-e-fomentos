# Faz o download de um artefato qualquer, baseado nas informações contidas em um JWT

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/download/{jwt}/{filename}:
    get:
      summary: >-
        Faz o download de um artefato qualquer, baseado nas informações contidas
        em um JWT
      deprecated: false
      description: ''
      tags:
        - SIGADOC/SIGAEX/download
        - download
      parameters:
        - name: jwt
          in: path
          description: JWT que referencia um artefato a ser baixado
          required: true
          example: '{{JWT_DOWNLOAD}}'
          schema:
            type: string
        - name: filename
          in: path
          description: Nome do arquivo de um artefato a ser baixado
          required: true
          example: '{{UUID_DOWNLOAD}}'
          schema:
            type: string
        - name: disposition
          in: query
          description: Informar attachment se desejar receber como um download
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
      x-apidog-folder: SIGADOC/SIGAEX/download
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389678-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
