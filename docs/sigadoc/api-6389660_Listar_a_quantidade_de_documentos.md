# Listar a quantidade de documentos

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/painel/quadro:
    get:
      summary: Listar a quantidade de documentos
      deprecated: false
      description: ''
      tags:
        - SIGADOC/SIGA/painel
        - painel
      parameters:
        - name: estilo
          in: query
          description: Informa se deseja agrupar
          required: false
          schema:
            type: string
        - name: tipoMarca
          in: query
          description: Informa se deseja filtrar por um tipo de marca
          required: false
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
                  list:
                    type: array
                    items:
                      $ref: '#/components/schemas/PainelQuadroItem'
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
      x-apidog-folder: SIGADOC/SIGA/painel
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389660-run
components:
  schemas:
    PainelQuadroItem:
      description: Lista quantidade de documentos na mesa
      type: object
      properties:
        finalidadeId:
          type: string
        tipoId:
          type: string
        tipoNome:
          type: string
        grupoId:
          type: string
        grupoNome:
          type: string
        marcadorEnum:
          type: string
        marcadorId:
          type: string
        marcadorNome:
          type: string
        marcadorIcone:
          type: string
        marcadorCor:
          type: string
        qtds:
          type: array
          items:
            $ref: '#/components/schemas/PainelQuadroQtdItem'
      x-apidog-orders:
        - finalidadeId
        - tipoId
        - tipoNome
        - grupoId
        - grupoNome
        - marcadorEnum
        - marcadorId
        - marcadorNome
        - marcadorIcone
        - marcadorCor
        - qtds
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
    PainelQuadroQtdItem:
      description: Lista quantidade de documentos na mesa
      type: object
      properties:
        tipo:
          type: string
        filtro:
          type: string
        qtd:
          type: string
      x-apidog-orders:
        - tipo
        - filtro
        - qtd
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
