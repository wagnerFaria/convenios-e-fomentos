# Cargo

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths: {}
components:
  schemas:
    Cargo:
      type: object
      description: Dados do Cargo da Pessoa
      properties:
        idCargo:
          type: string
          description: Id do Cargo
        idCargoIni:
          type: string
          description: Id inicial do Cargo
        idExterna:
          type: string
          description: id do Cargo Externo
        sigla:
          type: string
          description: Sigla do Cargo
        nome:
          type: string
          description: Nome do Cargo
      x-apidog-orders:
        - idCargo
        - idCargoIni
        - idExterna
        - sigla
        - nome
      x-apidog-folder: ''
  securitySchemes: {}
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
