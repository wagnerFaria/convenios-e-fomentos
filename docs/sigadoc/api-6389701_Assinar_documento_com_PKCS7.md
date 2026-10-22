# Assinar documento com PKCS7

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/documentos/{sigla}/assinar-com-pkcs7:
    post:
      summary: Assinar documento com PKCS7
      deprecated: false
      description: Necessário informar  a sigla e o arquivo .ps7 em base64.
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
      requestBody:
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              properties:
                pkcs7:
                  description: Arquivo PKCS7 em formato base64
                  type: string
                  example: ''
              required:
                - pkcs7
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
                x-apidog-ignore-properties: []
          headers: {}
          x-apidog-name: Successful response
        '400':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  error:
                    type: string
                    description: |
                      error:
                       *  A sigla do documento é obrigatória.
                       *  Arquivo PKCS7, no formato Base64, é obrigatório.
                    enum:
                      - A sigla do documento é obrigatória.
                      - Arquivo PKCS7, no formato Base64, é obrigatório.
                    default: A sigla do documento é obrigatória.
                x-apidog-orders:
                  - error
                x-apidog-ignore-properties: []
          headers: {}
          x-apidog-name: '400'
        '404':
          description: ''
          content:
            application/json:
              schema: &ref_0
                $ref: '#/components/schemas/Error'
          headers: {}
          x-apidog-name: Documento informado não encontrado.
        '500':
          description: ''
          content:
            application/json:
              schema: *ref_0
          headers: {}
          x-apidog-name: Error ocurred
      security: []
      x-apidog-folder: SIGADOC/SIGAEX/acao
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389701-run
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
