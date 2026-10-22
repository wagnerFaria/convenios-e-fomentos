# /documentos/{id}/verificar-assinatura

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/documentos/{id}/verificar-assinatura:
    get:
      summary: /documentos/{id}/verificar-assinatura
      deprecated: false
      description: Verificar assinatura do documento
      tags:
        - SIGADOC/SIGAEX/consulta
        - consulta
      parameters:
        - name: id
          in: path
          description: Id do documento
          required: true
          example: ''
          schema:
            type: string
        - name: assinaturaDigital
          in: query
          description: Flag para verificação de assinatura digital ou senha
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
                  idDoc:
                    description: Id do documento
                    type: string
                  idMov:
                    description: Id da movimentação da assinatura válida
                    type: string
                  status:
                    type: string
                    default: ERRO!
                x-apidog-orders:
                  - idDoc
                  - idMov
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
      x-apidog-folder: SIGADOC/SIGAEX/consulta
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389694-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
