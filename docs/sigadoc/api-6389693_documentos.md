# /documentos

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/documentos:
    get:
      summary: /documentos
      deprecated: false
      description: >-
        Exibe uma lista de mobils com suas marcas, de acordo com a lista de
        parâmetros. Equivalente à pesquisa de documentos no SIGA.  Liberado
        somente para pesquisa de documentos da própria lotação.
      tags:
        - SIGADOC/SIGAEX/consulta
        - consulta
      parameters:
        - name: offset
          in: query
          description: >-
            Irá trazer o resultado pulando a quantidade de registros informada
            neste campo, para efeito de paginação
          required: false
          schema:
            type: integer
        - name: qtdmax
          in: query
          description: >-
            Quantidade máxima de registros a trazer. Se não informada, assume 50
            registros
          required: false
          schema:
            type: integer
        - name: ordenacao
          in: query
          description: >-
            Ordenação da lista:<ul> <li>1 - Data descendente</li> <li>2 - Data
            ascendente</li> <li>3 - Data de criação do temporário
            descendente</li> <li>4 - Data de criação do temporário
            ascendente</li>  <li>5 - Data da situação descendente</li> <li>6 -
            Data da situação ascendente</li> <li>7 - Data da finalização do doc.
            descendente</li> <li>8 - Data da finalização do doc. ascendente</li>
            </ul>
          required: false
          schema:
            type: integer
        - name: marcador
          in: query
          description: Nome do marcador (situacao)
          required: false
          schema:
            type: string
        - name: grupomarcador
          in: query
          description: Nome do grupo de marcadores da mesa
          required: false
          schema:
            type: string
        - name: idpessoa
          in: query
          description: Identificador da Pessoa Inicial
          required: false
          schema:
            type: integer
        - name: idlotacao
          in: query
          description: >-
            Identificador da Lotação Inicial. Pesquisa de outras lotações
            liberada somente para usuários autorizados. Se não informado, assume
            a lotação do usuário.
          required: false
          schema:
            type: integer
        - name: dtinicial
          in: query
          description: Data inicial para pesquisa (aaaa-mm-dd)
          required: false
          schema:
            type: string
        - name: dtfinal
          in: query
          description: Data final para pesquisa (aaaa-mm-dd)
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
                      type: object
                      properties: {}
                x-apidog-orders:
                  - list
          headers: {}
          x-apidog-name: Successful response
        '404':
          description: ''
          content:
            application/json:
              schema:
                type: object
                properties: {}
          headers: {}
          x-apidog-name: Nenhum documento foi encontrado com os argumentos informados
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
      x-apidog-folder: SIGADOC/SIGAEX/consulta
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389693-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
