# Pesquisa de órgãos

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/orgaos:
    get:
      summary: Pesquisa de órgãos
      deprecated: false
      description: Retorna uma lista de órgãos.
      tags:
        - SIGADOC/SIGA/cadastro
        - cadastro
      parameters:
        - name: idOrgao
          in: query
          description: Identificador do Órgão (id)
          required: false
          schema:
            type: string
        - name: texto
          in: query
          description: Texto a ser pesquisado
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
                  list:
                    type: array
                    items:
                      $ref: '#/components/schemas/Orgao'
                x-apidog-orders:
                  - list
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
      x-apidog-folder: SIGADOC/SIGA/cadastro
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389649-run
components:
  schemas:
    Orgao:
      type: object
      description: Dados do Órgão da pessoa para perfil de acesso
      properties:
        idOrgao:
          type: string
          description: Id do Órgão da Pessoa
        sigla:
          type: string
          description: Sigla do Órgão da Pessoa
        nome:
          type: string
          description: Nome do Órgão da Pessoa
      x-apidog-orders:
        - idOrgao
        - sigla
        - nome
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
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
