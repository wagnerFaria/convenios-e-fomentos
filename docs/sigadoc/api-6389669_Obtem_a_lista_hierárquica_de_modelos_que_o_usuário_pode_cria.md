# Obtem a lista hierárquica de modelos que o usuário pode criar. Os nomes dos modelos devem estar cadastrados no formato Grupo: Nome do Modelo (separado por dois pontos)

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/modelos/lista-hierarquica:
    get:
      summary: >-
        Obtem a lista hierárquica de modelos que o usuário pode criar. Os nomes
        dos modelos devem estar cadastrados no formato Grupo: Nome do Modelo
        (separado por dois pontos)
      deprecated: false
      description: ''
      tags:
        - SIGADOC/SIGAEX/modelo
        - modelo
      parameters:
        - name: isEditandoAnexo
          in: query
          description: Informa true se está editando anexo
          required: false
          schema:
            type: string
        - name: isCriandoSubprocesso
          in: query
          description: Informa true se está criando subprocesso
          required: false
          schema:
            type: string
        - name: siglaMobPai
          in: query
          description: Id do mobil pai
          required: false
          schema:
            type: string
        - name: isAutuando
          in: query
          description: Informa true se está autuando
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
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389669-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
