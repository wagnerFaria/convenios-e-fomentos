# Recebe login e senha e gera um token JWT para ser usado nos outros métodos da API

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/autenticar:
    post:
      summary: >-
        Recebe login e senha e gera um token JWT para ser usado nos outros
        métodos da API
      deprecated: false
      description: >


        :::info

        #####

        ##### Método utilizado para realizar a autenticação nas apis do Sigadoc

        ###### Forneça o username e senha do usuário de sistema utilizando a
        autenticação do tipo Basic

        ###### No retorno será disponibilizado um token do tipo Bearer que
        deverá ser utilizado para todos os outros métodos, sendo informado no
        Header Authorization 

        :::
      tags:
        - SIGADOC/SIGA/auth
        - auth
      parameters:
        - name: matricula
          in: query
          description: Informar a matrícula que o usuário de sistema irá representar
          required: true
          example: ''
          schema:
            type: string
        - name: x-road-client
          in: header
          description: ''
          required: true
          example: '{{X_ROAD_CLIENT}}'
          schema:
            type: string
      requestBody:
        content:
          multipart/form-data:
            schema:
              type: object
              properties: {}
            example: {}
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  token:
                    type: string
                    description: Token JWT
                x-apidog-orders:
                  - token
                x-apidog-ignore-properties: []
              example:
                token: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9....
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
      security:
        - basic: []
      x-apidog-folder: SIGADOC/SIGA/auth
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389658-run
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
  securitySchemes:
    basic:
      type: http
      scheme: basic
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
