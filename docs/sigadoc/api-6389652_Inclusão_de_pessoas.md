# Inclusão de pessoas

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/pessoas:
    post:
      summary: Inclusão de pessoas
      deprecated: false
      description: >-
        Inclui uma pessoa no cadastro de pessoas e cria a identidade para a
        mesma.  Acesso restrito para usuários autorizados.
      tags:
        - SIGADOC/SIGA/cadastro
        - cadastro
      parameters:
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
                siglaOrgao:
                  description: Sigla do órgão
                  example: ''
                  type: string
                idCargoIni:
                  description: Id inicial do cargo
                  example: ''
                  type: string
                idFuncaoConfiancaIni:
                  description: Id inicial da função de confiança
                  example: ''
                  type: string
                lotacao:
                  description: Sigla completa da lotação integrada ao Siga
                  example: ''
                  type: string
                formCpf:
                  description: Número do CPF, com ou sem traço e pontos
                  example: ''
                  type: string
                nome:
                  description: Nome
                  example: ''
                  type: string
                nomeAbreviado:
                  description: Nome abreviado
                  example: ''
                  type: string
                dataNascimento:
                  description: Data de nascimento no formato dd/mm/aaaa
                  example: ''
                  type: string
                email:
                  description: E-mail
                  example: ''
                  type: string
                rg:
                  description: Número do RG com ou sem traços ou pontos
                  example: ''
                  type: string
                rgOrgaoExpedidor:
                  description: Órgão Expedidor do RG
                  example: ''
                  type: string
                rgUf:
                  description: UF do RG (Estado). Sigla com dois caracteres.
                  example: ''
                  type: string
                rgDataExpedicao:
                  description: Data de expedição do RG no formato dd/mm/aaaa
                  example: ''
                  type: string
                enviarEmailAcesso:
                  description: >-
                    Opção para enviar o e-mail de acesso automaticamente ao
                    usuário
                  example: ''
                  type: string
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  idPessoa:
                    type: string
                  siglaPessoa:
                    type: string
                x-apidog-orders:
                  - idPessoa
                  - siglaPessoa
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
          x-apidog-name: Órgão, lotação, cargo ou função de confiança não existente.
        '403':
          description: ''
          content:
            application/json:
              schema: *ref_0
          headers: {}
          x-apidog-name: Inclusão de pessoa não permitida
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389652-run
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
