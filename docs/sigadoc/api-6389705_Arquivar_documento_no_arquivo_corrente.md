# Arquivar documento no arquivo corrente

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/documentos/{sigla}/arquivar-corrente:
    post:
      summary: Arquivar documento no arquivo corrente
      deprecated: false
      description: Utilize este método para arquivar o documento.
      tags:
        - SIGADOC/SIGAEX/acao
        - acao
      parameters:
        - name: sigla
          in: path
          description: Sigla do mobil
          required: true
          example: ''
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
                  sigla:
                    description: Sigla do documento
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
      security: []
      x-apidog-folder: SIGADOC/SIGAEX/acao
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389705-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
