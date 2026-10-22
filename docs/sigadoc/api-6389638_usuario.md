# /usuario

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/usuario:
    get:
      summary: /usuario
      deprecated: false
      description: Enviar sugestões
      tags:
        - SIGADOC/SIGA/utils
        - utils
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
      responses:
        '200':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties:
                  usuario:
                    $ref: '#/components/schemas/Usuario'
                x-apidog-orders:
                  - usuario
                x-apidog-ignore-properties: []
          headers: {}
          x-apidog-name: Successful response
      security: []
      x-apidog-folder: SIGADOC/SIGA/utils
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389638-run
components:
  schemas:
    Usuario:
      description: Lista modelos disponíveis.
      type: object
      properties:
        identidadeId:
          type: string
        cadastranteId:
          type: string
        cadastranteSigla:
          type: string
        cadastranteNome:
          type: string
        cadastranteCpf:
          type: string
        lotaCadastranteId:
          type: string
        lotaCadastranteSigla:
          type: string
        lotaCadastranteNome:
          type: string
        titularId:
          type: string
        titularSigla:
          type: string
        titularNome:
          type: string
        titularCpf:
          type: string
        lotaTitularId:
          type: string
        lotaTitularSigla:
          type: string
        lotaTitularNome:
          type: string
        substituicaoId:
          type: string
        substituicoesPermitidas:
          type: array
          items:
            $ref: '#/components/schemas/SubstituicaoItem'
      x-apidog-orders:
        - identidadeId
        - cadastranteId
        - cadastranteSigla
        - cadastranteNome
        - cadastranteCpf
        - lotaCadastranteId
        - lotaCadastranteSigla
        - lotaCadastranteNome
        - titularId
        - titularSigla
        - titularNome
        - titularCpf
        - lotaTitularId
        - lotaTitularSigla
        - lotaTitularNome
        - substituicaoId
        - substituicoesPermitidas
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
    SubstituicaoItem:
      description: Lista modelos disponíveis.
      type: object
      properties:
        substituicaoId:
          type: string
        substituicaoDataFim:
          type: string
          format: date-time
        substituicaoDataInicio:
          type: string
          format: date-time
        registroIdInicial:
          type: string
        registroDataFim:
          type: string
          format: date-time
        registroDataInicio:
          type: string
          format: date-time
        substitutoId:
          type: string
        substitutoSigla:
          type: string
        substitutoNome:
          type: string
        lotaSubstitutoId:
          type: string
        lotaSubstitutoSigla:
          type: string
        lotaSubstitutoNome:
          type: string
        titularId:
          type: string
        titularSigla:
          type: string
        titularNome:
          type: string
        lotaTitularId:
          type: string
        lotaTitularSigla:
          type: string
        lotaTitularNome:
          type: string
      x-apidog-orders:
        - substituicaoId
        - substituicaoDataFim
        - substituicaoDataInicio
        - registroIdInicial
        - registroDataFim
        - registroDataInicio
        - substitutoId
        - substitutoSigla
        - substitutoNome
        - lotaSubstitutoId
        - lotaSubstitutoSigla
        - lotaSubstitutoNome
        - titularId
        - titularSigla
        - titularNome
        - lotaTitularId
        - lotaTitularSigla
        - lotaTitularNome
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
