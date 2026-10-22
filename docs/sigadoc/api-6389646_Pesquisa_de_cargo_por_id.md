# Pesquisa de cargo por id

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/cargos/{id}:
    get:
      summary: Pesquisa de cargo por id
      deprecated: false
      description: Retorna o cargo correspondente ao id pesquisado
      tags:
        - SIGADOC/SIGA/cadastro
        - cadastro
      parameters:
        - name: id
          in: path
          description: id da entidade
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
                  cargo:
                    $ref: '#/components/schemas/Cargo'
                x-apidog-orders:
                  - cargo
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
          x-apidog-name: >-
            Acesso não permitido ou Usuário não autorizado para consultar dados
            deste órgão.
        '404':
          description: ''
          content:
            application/json:
              schema: *ref_0
          headers: {}
          x-apidog-name: Nenhum cargo foi encontrado para os parâmetros informados.
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389646-run
components:
  schemas:
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
