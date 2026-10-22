# Troca de PIN para Assinatura e Autenticação de Documentos e Segundo Fator de Autenticação.

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/pin/trocar:
    post:
      summary: >-
        Troca de PIN para Assinatura e Autenticação de Documentos e Segundo
        Fator de Autenticação.
      deprecated: false
      description: ''
      tags:
        - SIGADOC/SIGA/PIN
        - PIN
      parameters:
        - name: pinAtual
          in: query
          description: PIN Atual para Confirmação da Troca de PIN
          required: true
          schema:
            type: string
        - name: pin
          in: query
          description: >-
            PIN para Segundo Fator de Autenticação e Assinatura e Autenticação
            com Senha
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
                properties:
                  mensagem:
                    type: string
                x-apidog-orders:
                  - mensagem
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389634-run
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
