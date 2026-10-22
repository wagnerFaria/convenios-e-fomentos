# Pesquisa de pessoa por sigla

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/pessoas/{sigla}:
    get:
      summary: Pesquisa de pessoa por sigla
      deprecated: false
      description: Retorna a pessoa correspondente a sigla pesquisada
      tags:
        - SIGADOC/SIGA/cadastro
        - cadastro
      parameters:
        - name: sigla
          in: path
          description: Sigla da entidade
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
                  pessoa:
                    $ref: '#/components/schemas/Pessoa'
                x-apidog-orders:
                  - pessoa
                type: object
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
          x-apidog-name: Acesso não permitido
        '404':
          description: ''
          content:
            application/json:
              schema: *ref_0
          headers: {}
          x-apidog-name: Nenhuma Pessoa foi encontrada para a sigla informada.
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389643-run
components:
  schemas:
    Pessoa:
      type: object
      description: Dados de uma determinada Pessoa
      properties:
        idPessoa:
          type: string
          description: Id da pessoa
        idPessoaIni:
          type: string
          description: Id inicial da pessoa
        sigla:
          type: string
          description: Sigla da pessoa
        nome:
          type: string
          description: Nome da pessoa
        email:
          type: string
          description: Email da pessoa
        siglaLotacao:
          type: string
          description: Sigla da lotação da pessoa
        isExternaPessoa:
          type: boolean
          description: Devolve true se for pessoa externa e false se não for
        dataInicioPessoa:
          type: string
          format: date-time
          description: Data início da validade desta versão da pessoa (data da alteração)
        lotacao: &ref_1
          $ref: '#/components/schemas/Lotacao'
        cargo:
          $ref: '#/components/schemas/Cargo'
        funcaoConfianca:
          $ref: '#/components/schemas/FuncaoConfianca'
      x-apidog-orders:
        - idPessoa
        - idPessoaIni
        - sigla
        - nome
        - email
        - siglaLotacao
        - isExternaPessoa
        - dataInicioPessoa
        - lotacao
        - cargo
        - funcaoConfianca
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
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
    Cargo:
      type: object
      description: Dados do Cargo da Pessoa
      properties:
        idCargo:
          type: string
          description: Id do Cargo
        idCargoIni:
          type: string
          description: Id inicial do Cargo
        idExterna:
          type: string
          description: id do Cargo Externo
        sigla:
          type: string
          description: Sigla do Cargo
        nome:
          type: string
          description: Nome do Cargo
      x-apidog-orders:
        - idCargo
        - idCargoIni
        - idExterna
        - sigla
        - nome
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
    Lotacao:
      type: object
      description: Dados de uma determinada Lotação
      properties:
        idLotacao:
          type: string
          description: Id da lotação
        idLotacaoIni:
          type: string
          description: Id inicial da lotação
        sigla:
          type: string
          description: Sigla completa da Lotação (com o órgão)
        siglaLotacao:
          type: string
          description: Sigla da lotação (sem o órgão)
        nome:
          type: string
          description: Nome da lotação
        localidade:
          $ref: '#/components/schemas/Localidade'
        orgao:
          $ref: '#/components/schemas/Orgao'
        superiores:
          type: array
          items: *ref_1
        inferiores:
          type: array
          items: *ref_1
      x-apidog-orders:
        - idLotacao
        - idLotacaoIni
        - sigla
        - siglaLotacao
        - nome
        - localidade
        - orgao
        - superiores
        - inferiores
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
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
    Localidade:
      type: object
      description: Dados da Localidade
      properties:
        idLocalidade:
          type: string
          description: Id da localidade
        nome:
          type: string
          description: Nome da localidade
        uf:
          $ref: '#/components/schemas/Uf'
      x-apidog-orders:
        - idLocalidade
        - nome
        - uf
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
    Uf:
      type: object
      description: Dados das UFs
      properties:
        idUf:
          type: string
          description: Id da UF
        nomeUf:
          type: string
          description: Nome da UF
      x-apidog-orders:
        - idUf
        - nomeUf
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
