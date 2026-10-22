# Pesquisa de órgão por sigla

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/orgaos/{sigla}:
    get:
      summary: Pesquisa de órgão por sigla
      deprecated: false
      description: Retorna o órgão correspondente a sigla pesquisada
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
                  orgao:
                    $ref: '#/components/schemas/Orgao'
                x-apidog-orders:
                  - orgao
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
          x-apidog-name: Nenhum Órgão foi encontrado para a sigla informada.
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389645-run
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
