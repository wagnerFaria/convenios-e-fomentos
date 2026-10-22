# Listar elementos por filtro

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/painel/lista:
    get:
      summary: Listar elementos por filtro
      deprecated: false
      description: ''
      tags:
        - SIGADOC/SIGA/painel
        - painel
      parameters:
        - name: idMarcadores
          in: query
          description: Identificadores iniciais dos Marcadores separados por vírgula
          required: true
          schema:
            type: string
        - name: filtroPessoaLotacao
          in: query
          description: Informa se deseja filtrar documentos por pessoa ou lotação
          required: false
          schema:
            type: string
        - name: tipoMarca
          in: query
          description: Informa se deseja filtrar por um tipo de marca
          required: false
          schema:
            type: string
        - name: itensPorPagina
          in: query
          description: Número de itens por página
          required: true
          schema:
            type: string
        - name: pagina
          in: query
          description: Número da página
          required: true
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
                      $ref: '#/components/schemas/PainelListaItem'
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389661-run
components:
  schemas:
    PainelListaItem:
      description: Lista quantidade de documentos na mesa
      type: object
      properties:
        marcaId:
          type: string
        marcaTipo:
          type: string
        marcaTexto:
          type: string
        marcaIcone:
          type: string
        dataIni:
          type: string
          format: date-time
        dataFim:
          type: string
          format: date-time
        moduloId:
          type: string
        refId:
          type: string
        movId:
          type: string
        tipo:
          type: string
        codigo:
          type: string
        sigla:
          type: string
        descricao:
          type: string
        origem:
          type: string
        ultimaAnotacao:
          type: string
        marcas:
          type: array
          items:
            $ref: '#/components/schemas/Marca'
        acoes:
          type: array
          items:
            $ref: '#/components/schemas/Acao'
      x-apidog-orders:
        - marcaId
        - marcaTipo
        - marcaTexto
        - marcaIcone
        - dataIni
        - dataFim
        - moduloId
        - refId
        - movId
        - tipo
        - codigo
        - sigla
        - descricao
        - origem
        - ultimaAnotacao
        - marcas
        - acoes
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
    Acao:
      description: Ação que pode ser realizada a um documento
      type: object
      properties:
        nome:
          type: string
        icone:
          type: string
        ativa:
          type: boolean
      x-apidog-orders:
        - nome
        - icone
        - ativa
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
    Marca:
      description: Etiquetas relacionadas a um documento
      type: object
      properties:
        pessoa:
          type: string
        lotacao:
          type: string
        nome:
          type: string
        icone:
          type: string
        titulo:
          type: string
        marcaId:
          type: string
        inicio:
          type: string
          format: date-time
        termino:
          type: string
          format: date-time
        daPessoa:
          type: boolean
        deOutraPessoa:
          type: boolean
        daLotacao:
          type: boolean
      x-apidog-orders:
        - pessoa
        - lotacao
        - nome
        - icone
        - titulo
        - marcaId
        - inicio
        - termino
        - daPessoa
        - deOutraPessoa
        - daLotacao
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
