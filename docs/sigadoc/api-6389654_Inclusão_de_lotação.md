# Inclusão de lotação

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/lotacoes:
    post:
      summary: Inclusão de lotação
      deprecated: false
      description: >-
        Inclui uma lotação no cadastro de lotações.  Acesso restrito à usuários
        autorizados.
      tags:
        - SIGADOC/SIGA/cadastro
        - cadastro
      parameters:
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
      requestBody:
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              properties:
                siglaOrgao:
                  description: Sigla do órgão
                  example: ''
                  type: string
                nome:
                  description: Nome
                  example: ''
                  type: string
                sigla:
                  description: Sigla da lotação a criar
                  example: ''
                  type: string
                idLotacaoPaiIni:
                  description: Identificador da Lotação Pai Inicial
                  example: ''
                  type: string
                idLocalidade:
                  description: Identificador da Localidade
                  example: ''
                  type: string
                isAcessoExterno:
                  description: Lotação é de acesso externo? (true/false)
                  example: ''
                  type: string
              required:
                - sigla
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  idLotacao:
                    type: string
                  siglaLotacao:
                    type: string
                x-apidog-orders:
                  - idLotacao
                  - siglaLotacao
                x-apidog-ignore-properties: []
          headers: {}
          x-apidog-name: Criado com sucesso
        '400':
          description: ''
          content:
            application/json:
              schema: &ref_0
                $ref: '#/components/schemas/Error'
          headers: {}
          x-apidog-name: Órgão, unidade pai ou localidade não existente.
        '403':
          description: ''
          content:
            application/json:
              schema: *ref_0
          headers: {}
          x-apidog-name: Inclusão de lotação não permitida
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389654-run
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
