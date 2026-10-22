# Validação de PIN informado do Usuário Autenticado para Assinatura e Autenticação de Documentos e Segundo Fator de Autenticação.

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/pin/validar:
    post:
      summary: >-
        Validação de PIN informado do Usuário Autenticado para Assinatura e
        Autenticação de Documentos e Segundo Fator de Autenticação.
      deprecated: false
      description: >-
        Para realizar a verificação é necessário Enviar o Token de Autenticação
        do Usuário e no parâmetro pinBasicAuth enviar o [CPF do Usuario
        Autenticado]:[PIN informado] em Base64
      tags:
        - SIGADOC/SIGA/PIN
        - PIN
      parameters:
        - name: pinBasicAuth
          in: query
          description: PIN para validação no formato Basic Auth - Base64(CPF:PIN)
          required: true
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
                properties: {}
                x-apidog-orders: []
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
      x-apidog-folder: SIGADOC/SIGA/PIN
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389632-run
components:
  schemas:
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
