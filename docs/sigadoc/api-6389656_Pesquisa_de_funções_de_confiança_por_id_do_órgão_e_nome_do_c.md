# Pesquisa de funções de confiança por id do órgão e nome do cargo

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/funcoes:
    get:
      summary: Pesquisa de funções de confiança por id do órgão e nome do cargo
      deprecated: false
      description: >-
        Retorna uma lista de funções de confiança correspondentes ao id do órgão
        e nome pesquisados. Se não for informado o nome, retorna todos do órgão.
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
        - name: nomeFuncaoConfianca
          in: query
          description: Nome da função de confiança cargo ou parte do nome
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
                type: object
                properties:
                  list:
                    type: array
                    items:
                      $ref: '#/components/schemas/FuncaoConfianca'
                x-apidog-orders:
                  - list
                x-apidog-ignore-properties: []
          headers: {}
          x-apidog-name: Successful response
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389656-run
components:
  schemas:
    FuncaoConfianca:
      type: object
      description: Dados da Função de Confiança da Pessoa
      properties:
        idFuncaoConfianca:
          type: string
          description: Id da função de confiança
        idFuncaoConfiancaIni:
          type: string
          description: Id inicial da função de confiança
        idExterna:
          type: string
          description: id da função de confiança externa
        sigla:
          type: string
          description: Sigla da função de confiança
        nome:
          type: string
          description: Nome da função de confiança
        idpai:
          type: string
          description: id da função de confiança pai
      x-apidog-orders:
        - idFuncaoConfianca
        - idFuncaoConfiancaIni
        - idExterna
        - sigla
        - nome
        - idpai
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
