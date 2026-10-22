# Pesquisa de localidade por id

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /siga/localidades/{id}:
    get:
      summary: Pesquisa de localidade por id
      deprecated: false
      description: Retorna a localidade correspondente ao id pesquisado
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
                  localidade:
                    $ref: '#/components/schemas/Localidade'
                x-apidog-orders:
                  - localidade
                type: object
                x-apidog-ignore-properties: []
          headers: {}
          x-apidog-name: Successful response
        '404':
          description: ''
          content:
            application/json:
              schema: &ref_0
                $ref: '#/components/schemas/Error'
          headers: {}
          x-apidog-name: Nenhuma localidade foi encontrada para os parâmetros informados.
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389648-run
components:
  schemas:
    Localidade:
      type: object
      description: Dados da Localidade
      properties:
        idLocalidade:
          type: string
          description: Id da localidade
        nome:
          type: string
          description: Nome da localidade
        uf:
          $ref: '#/components/schemas/Uf'
      x-apidog-orders:
        - idLocalidade
        - nome
        - uf
      x-apidog-ignore-properties: []
      x-apidog-folder: ''
    Uf:
      type: object
      description: Dados das UFs
      properties:
        idUf:
          type: string
          description: Id da UF
        nomeUf:
          type: string
          description: Nome da UF
      x-apidog-orders:
        - idUf
        - nomeUf
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
