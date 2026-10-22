# /documentos/{sigla}/publicacao

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/documentos/{sigla}/publicacao:
    get:
      summary: /documentos/{sigla}/publicacao
      deprecated: false
      description: Consulta dados de publicação no diário oficial
      tags:
        - SIGADOC/SIGAEX/consulta
        - consulta
      parameters:
        - name: sigla
          in: path
          description: Sigla do documento
          required: true
          example: ''
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
                  sigla:
                    description: Sigla do documento
                    type: string
                  data:
                    description: Data da Publicação no Diário Oficial
                    type: string
                  pagina:
                    description: Página da Publicação no Diário Oficial
                    type: string
                  numeroDiario:
                    description: Número do Diário Oficial
                    type: string
                  protocolo:
                    description: Protocolo de Envio da Publicação no Diário Oficial
                    type: string
                x-apidog-orders:
                  - sigla
                  - data
                  - pagina
                  - numeroDiario
                  - protocolo
                x-apidog-ignore-properties: []
          headers: {}
          x-apidog-name: Successful response
        '400':
          description: ''
          content:
            application/json:
              schema: &ref_0
                $ref: '#/components/schemas/Error'
          headers: {}
          x-apidog-name: A sigla do documento é obrigatória
        '404':
          description: ''
          content:
            application/json:
              schema: *ref_0
          headers: {}
          x-apidog-name: Nenhum dado de publicação foi encontrado
        '500':
          description: ''
          content:
            application/json:
              schema: *ref_0
          headers: {}
          x-apidog-name: Error ocurred
      security: []
      x-apidog-folder: SIGADOC/SIGAEX/consulta
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6444463-run
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
