# Inicia o processo de geração do PDF completo de um documento para visualização

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/documentos/{sigla}/arquivo:
    get:
      summary: >-
        Inicia o processo de geração do PDF completo de um documento para
        visualização
      deprecated: false
      description: >-
        <p>Solicita uma execução assíncrona de geração de PDF completo contendo
        ou não os documentos filhos e as estampas (marca d'água, assinatura,
        etc). <p>Irá devolver um json contendo um JWT que deverá ser passado
        para o webservice GET /download/{jwt}/{filename}, que fará o download do
        PDF gerado após este estar completamente gerado. <p>Para verificar
        andamento da geração do PDF, utilizar o webservice GET /status/{chave}
        onde chave é o campo uuid devolvido no JSON de retorno deste webservice.
      tags:
        - SIGADOC/SIGAEX/download
        - download
      parameters:
        - name: sigla
          in: path
          description: Sigla do mobil
          required: true
          example: '{{SIGLA_DOCUMENTO}}'
          schema:
            type: string
        - name: contenttype
          in: query
          description: >-
            Informar o tipo de arquivo desejado, pode ser application/pdf ou
            text/html
          required: false
          schema:
            type: string
        - name: estampa
          in: query
          description: Informar false se deseja receber o PDF sem estampas
          required: false
          schema:
            type: string
        - name: completo
          in: query
          description: Informar true se deseja receber todos os arquivos juntados e anexos
          required: false
          schema:
            type: string
        - name: volumes
          in: query
          description: >-
            Informar true se deseja receber uma compilação com todos os volumes
            de um processo
          required: false
          schema:
            type: string
        - name: exibirReordenacao
          in: query
          description: Informar true se deseja habilitar a reordenação dos documentos
          required: false
          schema:
            type: string
        - name: tamanhoOriginal
          in: query
          description: >-
            Informar true se deseja receber o PDF no tamanho orginal não
            redimensionado para A4
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
                  uuid:
                    type: string
                    description: >-
                      UUID que pode ser usado como chave para obter status da
                      geração do PDF
                  jwt:
                    type: string
                    description: JWT com as informações para download do PDF
                x-apidog-orders:
                  - uuid
                  - jwt
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
      x-apidog-folder: SIGADOC/SIGAEX/download
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389679-run
components:
  schemas: {}
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
