# Pesquisa matrícula por cpf e sigla do orgão

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/matricula:
    get:
      summary: Pesquisa matrícula por cpf e sigla do orgão
      deprecated: false
      description: Retorna a matrícula.
      tags:
        - SIGADOC/SIGA/cadastro
        - cadastro
      parameters:
        - name: cpf
          in: query
          description: CPF do usuário, somente números sem ponto ou traço
          required: false
          schema:
            type: string
        - name: siglaDoOrgao
          in: query
          description: Sigla do órgão
          required: false
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
          required: true
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
                description: matricula
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
                       *  O cpf é obrigatório!
                       *  O cpf é inválido!
                       *  A sigla do Orgão é obrigatória!
                    enum:
                      - O cpf é obrigatório!
                      - O cpf é inválido!
                      - A sigla do Orgão é obrigatória.
                    default: A sigla do Orgão é obrigatória.
                x-apidog-orders:
                  - error
                x--orders:
                  - error
                x--ignore-properties: []
          headers: {}
          x-apidog-name: '400'
        '403':
          description: ''
          content:
            application/json:
              schema: &ref_0
                $ref: '#/components/schemas/Error'
          headers: {}
          x-apidog-name: >-
            Acesso não permitido ou Usuário não autorizado para consultar dados
            deste órgão.
        '404':
          description: ''
          content:
            application/json:
              schema: *ref_0
          headers: {}
          x-apidog-name: >-
            Nenhuma função de confiança foi encontrada para os parâmetros
            informados.
        '500':
          description: ''
          content:
            application/json:
              schema: *ref_0
          headers: {}
          x-apidog-name: Error ocurred
      security: []
      x-apidog-folder: SIGADOC/SIGA/cadastro
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389642-run
components:
  schemas:
    Error:
      type: object
      properties:
        errormsg:
          type: string
      x-apidog-orders:
        - errormsg
      x--orders:
        - errormsg
      x--ignore-properties: []
      x-apidog-folder: ''
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
