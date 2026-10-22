# Pesquisa de lotacoes por sigla

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/lotacoes/{sigla}:
    get:
      summary: Pesquisa de lotacoes por sigla
      deprecated: false
      description: Retorna a lotacao correspondente a sigla pesquisda
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
                  lotacao: &ref_1
                    $ref: '#/components/schemas/Lotacao'
                x-apidog-orders:
                  - lotacao
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
          x-apidog-name: Nenhuma lotação foi encontrada para a sigla informada.
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389644-run
components:
  schemas:
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
