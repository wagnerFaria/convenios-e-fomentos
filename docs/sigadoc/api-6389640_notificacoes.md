# /notificacoes

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/notificacoes:
    get:
      summary: /notificacoes
      deprecated: false
      description: Listar as notificações para determinado Usuário
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
                  list:
                    type: array
                    items:
                      $ref: '#/components/schemas/Notificacao'
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
      x-apidog-folder: SIGADOC/SIGA/utils
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389640-run
components:
  schemas:
    Notificacao:
      type: object
      description: Dados de Notificacao
      properties:
        idNotificacao:
          type: string
          description: Id do Cargo
        titulo:
          type: string
          description: Título da Notificação
        icone:
          type: string
          description: Class Font Awesome do Icone da Notificação
        resumo:
          type: string
          description: Resumo da Notificação
        conteudo:
          type: string
          description: Conteúdo da Notificação
        dataInicio:
          type: string
          format: date-time
          description: Data de Início da Notificação
        dataTermino:
          type: string
          format: date-time
          description: Data de Início de Término da Notificação
        sempreMostrar:
          type: boolean
          description: >-
            true se notificação não for dispensável. false se notificação for
            dispensável
      x-apidog-orders:
        - idNotificacao
        - titulo
        - icone
        - resumo
        - conteudo
        - dataInicio
        - dataTermino
        - sempreMostrar
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
