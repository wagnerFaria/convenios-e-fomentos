# Obtem informações sobre um modelo específico

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/modelos/{id}:
    get:
      summary: Obtem informações sobre um modelo específico
      deprecated: false
      description: ''
      tags:
        - SIGADOC/SIGAEX/modelo
        - modelo
      parameters:
        - name: id
          in: path
          description: Identificador da entidade
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
                type: object
                properties:
                  idModelo:
                    type: string
                  idModeloInicial:
                    type: string
                  nome:
                    type: string
                  descr:
                    type: string
                  especie:
                    type: string
                  nivelDeAcesso:
                    type: string
                  classificacao:
                    type: string
                  classificacaoParaCriacaoDeVias:
                    type: string
                  tipoDeSubscritor:
                    type: string
                  tipoDeDestinatario:
                    type: string
                  tipoDeConteudo:
                    type: string
                  tipoDeDocumento:
                    type: string
                  conteudo:
                    type: string
                x-apidog-orders:
                  - idModelo
                  - idModeloInicial
                  - nome
                  - descr
                  - especie
                  - nivelDeAcesso
                  - classificacao
                  - classificacaoParaCriacaoDeVias
                  - tipoDeSubscritor
                  - tipoDeDestinatario
                  - tipoDeConteudo
                  - tipoDeDocumento
                  - conteudo
          headers: {}
          x-apidog-name: Successful response
        '500':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties: {}
          headers: {}
          x-apidog-name: Error ocurred
      security: []
      x-apidog-folder: SIGADOC/SIGAEX/modelo
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389675-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
