# /documentos/{sigla}/dossie

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/documentos/{sigla}/dossie:
    get:
      summary: /documentos/{sigla}/dossie
      deprecated: false
      description: Obtem a lista de documentos que compoem o dossiê
      tags:
        - SIGADOC/SIGAEX/consulta
        - consulta
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
                properties:
                  list:
                    type: array
                    items:
                      type: object
                      properties: {}
                x-apidog-orders:
                  - list
                type: object
          headers: {}
          x-apidog-name: Successful response
        '403':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties: {}
          headers: {}
          x-apidog-name: Acesso ao documento não permitido
        '404':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties: {}
          headers: {}
          x-apidog-name: Protocolo ainda não foi gerado para este documento
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389692-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
