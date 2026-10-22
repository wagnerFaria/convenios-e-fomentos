# Gera o HTML da entrevista a partir dos campos preenchidos

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/modelos/{id}/processar-entrevista:
    post:
      summary: Gera o HTML da entrevista a partir dos campos preenchidos
      deprecated: false
      description: ''
      tags:
        - SIGADOC/SIGAEX/modelo
        - modelo
      parameters:
        - name: id
          in: path
          description: Identificador da entidade
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
                entrevista:
                  description: >-
                    Campos e conteudos do modelo do documento a ser criado. Para
                    obtê-los, acessar o webservice de modelos. Pode ser descrito
                    em dois formatos, JSON ou Form Data
                    (chave1=valor1&chave2=valor2&chave3=valor3)
                  example: ''
                  type: string
              required:
                - entrevista
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
      x-apidog-folder: SIGADOC/SIGAEX/modelo
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389665-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
