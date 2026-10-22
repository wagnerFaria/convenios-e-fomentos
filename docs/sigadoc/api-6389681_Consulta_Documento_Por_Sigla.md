# Consulta Documento Por Sigla

## OpenAPI Specification

```yaml
openapi: 3.0.1
info:
  title: ''
  description: ''
  version: 1.0.0
paths:
  /sigaex/documentos/{sigla}:
    get:
      summary: Consulta Documento Por Sigla
      deprecated: false
      description: >-
        <h3>Exibir um documento</h3> <p>É devolvido um JSON com uma estrutura
        contendo diversas informações do documento:</p> <ul>
        <li><b>subscritorString: </b>Nome do subscritor</li>
        <li><b>dtFinalizacao: </b>Data em que o documento foi finalizado</li>
        <li><b>descrDocumento: </b>Descrição do documento</li>
        <li><b>conteudoBlobHtmlString: </b>Conteúdo do documento em HTML</li>
        <li><b>mobs: </b>Vias ou volumes do documento, chamadas de mobil no
        sistema. Contém sempre uma via ou volume <b>geral</b> que conterá
        movimentações ou ações referentes ao documento geral e pelo menos uma
        <b>via ou volume</b> contendo movimentações ou ações específicas desta
        via ou volume.</li> <ul> <li><b>isGeral: </b>Se <i>true</i> significa
        que é a via ou volume geral</li> <li><b>expedientesFilhosNaoCancelados:
        </b>Documentos da espécie <i>Expediente</i> não cancelados juntados a
        este <li><b>processosFilhosNaoCancelados: </b>Documentos da espécie
        <i>Processo</i> não cancelados juntados a este documento</li>
        <li><b>movs: </b>Movimentações (histórico) da via ou volume. Vem
        ordenado da mais recente para a mais antiga.</li> <ul> <li><b>cancelada:
        </b>Se true, a movimentação foi cancelada</li> <li><b>descricao:
        </b>Pode conter informações referentes à movimentação realizada</li>
        <li><b>parte:</b>Contém informações dos envolvidos na movimentação</li>
        <ul> <li><b>cadastrante: </b>Quem realizou a movimentação</li>
        <li><b>lotaCadastrante: </b>Lotação da pessoa que realizou a
        movimentação</li> <li><b>resp: </b>No caso de trâmite, pessoa para a
        qual foi tramitado</li> <li><b>lotaResp: </b>No caso de trâmite, lotação
        da pessoa para a qual foi tramitado</li> </ul>
        <li><b>exTipoMovimentacaoSigla: </b>Nome do tipo de movimentação
        realizada (Trâmitação, anotação, etc)</li> <li><b>dtIniMov: </b>Data em
        que a movimentação foi realizada</li> </ul> <li><b>marcadoresEmHtml:
        </b>Marcas (situações) da via ou volume.</li> <li><b>acoes:
        </b>Correspondente aos botões que aparecem na tela do documento. São as
        ações que podem ser realizadas para a via ou volume neste momento. As
        ações não devem ser utilizadas nos webservices, podem somente ser
        pesquisadas.</li> </ul> </ul>
      tags:
        - SIGADOC/SIGAEX/consulta
        - consulta
      parameters:
        - name: sigla
          in: path
          description: Sigla do mobil
          required: true
          example: MTIPRO202300035
          schema:
            type: string
        - name: completo
          in: query
          description: Informar true se deseja receber todos os arquivos juntados e anexos
          required: false
          example: 'true'
          schema:
            type: string
        - name: auditar
          in: query
          description: Informar true se deseja receber dados para auditoria
          required: false
          example: 'false'
          schema:
            type: string
        - name: desabilitarRecebimentoAutomatico
          in: query
          description: Informar true se deseja desabilitar o recebimento automático
          required: false
          example: 'false'
          schema:
            type: string
        - name: exibe
          in: query
          description: >-
            Informar true se deseja exibir lista de movimentações reduzidas no
            histórico
          required: false
          example: 'false'
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
                  mobs:
                    type: array
                    items:
                      type: object
                      properties:
                        sigla:
                          type: string
                        isGeral:
                          type: boolean
                        apensos:
                          type: array
                          items:
                            type: string
                        expedientesFilhosNaoCancelados:
                          type: array
                          items:
                            type: string
                        processosFilhosNaoCancelados:
                          type: array
                          items:
                            type: string
                        anexosNaoAssinados:
                          type: array
                          items:
                            type: string
                        despachosNaoAssinados:
                          type: array
                          items:
                            type: string
                        expedientesFilhosNaoJuntados:
                          type: array
                          items:
                            type: string
                        expedientesJuntadosNaoAssinados:
                          type: array
                          items:
                            type: string
                        pendenciasDeAnexacao:
                          type: array
                          items:
                            type: string
                        pendenciasDeColaboracao:
                          type: array
                          items:
                            type: string
                        movs:
                          type: array
                          items:
                            type: object
                            properties:
                              classe:
                                type: string
                              originadaAqui:
                                type: boolean
                              desabilitada:
                                type: boolean
                              cancelada:
                                type: boolean
                              descricao:
                                type: string
                              exTipoMovimentacao:
                                type: string
                              complemento:
                                type: string
                              parte:
                                type: object
                                properties:
                                  cadastrante:
                                    type: object
                                    properties:
                                      sigla:
                                        type: string
                                      siglaAmpliada:
                                        type: string
                                      descricao:
                                        type: string
                                      descricaoAmpliada:
                                        type: string
                                      iniciais:
                                        type: string
                                      siglaOrgao:
                                        type: string
                                    required:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                    x-apidog-orders:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                  lotaCadastrante:
                                    type: object
                                    properties:
                                      sigla:
                                        type: string
                                      siglaAmpliada:
                                        type: string
                                      descricao:
                                        type: string
                                      descricaoAmpliada:
                                        type: string
                                      iniciais:
                                        type: string
                                      siglaOrgao:
                                        type: string
                                    required:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                    x-apidog-orders:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                  lotaResp:
                                    type: object
                                    properties:
                                      sigla:
                                        type: string
                                      siglaAmpliada:
                                        type: string
                                      descricao:
                                        type: string
                                      descricaoAmpliada:
                                        type: string
                                      iniciais:
                                        type: string
                                      siglaOrgao:
                                        type: string
                                    required:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                    x-apidog-orders:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                  lotaSubscritor:
                                    type: object
                                    properties:
                                      sigla:
                                        type: string
                                      siglaAmpliada:
                                        type: string
                                      descricao:
                                        type: string
                                      descricaoAmpliada:
                                        type: string
                                      iniciais:
                                        type: string
                                      siglaOrgao:
                                        type: string
                                    required:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                    x-apidog-orders:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                  lotaTitular:
                                    type: object
                                    properties:
                                      sigla:
                                        type: string
                                      siglaAmpliada:
                                        type: string
                                      descricao:
                                        type: string
                                      descricaoAmpliada:
                                        type: string
                                      iniciais:
                                        type: string
                                      siglaOrgao:
                                        type: string
                                    required:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                    x-apidog-orders:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                  resp:
                                    type: object
                                    properties:
                                      sigla:
                                        type: string
                                      siglaAmpliada:
                                        type: string
                                      descricao:
                                        type: string
                                      descricaoAmpliada:
                                        type: string
                                      iniciais:
                                        type: string
                                      siglaOrgao:
                                        type: string
                                    required:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                    x-apidog-orders:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                  subscritor:
                                    type: object
                                    properties:
                                      sigla:
                                        type: string
                                      siglaAmpliada:
                                        type: string
                                      descricao:
                                        type: string
                                      descricaoAmpliada:
                                        type: string
                                      iniciais:
                                        type: string
                                      siglaOrgao:
                                        type: string
                                    required:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                    x-apidog-orders:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                  titular:
                                    type: object
                                    properties:
                                      sigla:
                                        type: string
                                      siglaAmpliada:
                                        type: string
                                      descricao:
                                        type: string
                                      descricaoAmpliada:
                                        type: string
                                      iniciais:
                                        type: string
                                      siglaOrgao:
                                        type: string
                                    required:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                    x-apidog-orders:
                                      - sigla
                                      - siglaAmpliada
                                      - descricao
                                      - descricaoAmpliada
                                      - iniciais
                                      - siglaOrgao
                                required:
                                  - cadastrante
                                  - lotaCadastrante
                                  - lotaResp
                                  - lotaSubscritor
                                  - lotaTitular
                                  - resp
                                  - subscritor
                                  - titular
                                x-apidog-orders:
                                  - cadastrante
                                  - lotaCadastrante
                                  - lotaResp
                                  - lotaSubscritor
                                  - lotaTitular
                                  - resp
                                  - subscritor
                                  - titular
                              dtIniMov:
                                type: string
                              dtRegMovDDMMYYHHMMSS:
                                type: string
                              descrTipoMovimentacao:
                                type: string
                              idMov:
                                type: integer
                              duracaoSpan:
                                type: integer
                              duracaoSpanExibirCompleto:
                                type: integer
                              duracao:
                                type: string
                              mimeType:
                                type: string
                              lotaCadastranteSigla:
                                type: string
                              exTipoMovimentacaoSigla:
                                type: string
                              tempoRelativo:
                                type: string
                              podeExibirNoSigale:
                                type: boolean
                              tipoMovimentacaoArquivamento:
                                type: boolean
                              acoes:
                                type: array
                                items:
                                  type: object
                                  properties:
                                    nome:
                                      type: string
                                    nameSpace:
                                      type: string
                                    acao:
                                      type: string
                                    pode:
                                      type: boolean
                                    params:
                                      type: object
                                      properties:
                                        popup:
                                          type: string
                                        sigla:
                                          type: string
                                        dtIni:
                                          type: string
                                        isTransf:
                                          type: string
                                      required:
                                        - sigla
                                      x-apidog-orders:
                                        - popup
                                        - sigla
                                        - dtIni
                                        - isTransf
                                    explicacao:
                                      type: string
                                    post:
                                      type: boolean
                                  required:
                                    - nome
                                    - nameSpace
                                    - acao
                                    - pode
                                    - params
                                    - explicacao
                                    - post
                                  x-apidog-orders:
                                    - nome
                                    - nameSpace
                                    - acao
                                    - pode
                                    - params
                                    - explicacao
                                    - post
                            required:
                              - classe
                              - originadaAqui
                              - desabilitada
                              - cancelada
                              - descricao
                              - exTipoMovimentacao
                              - complemento
                              - parte
                              - dtIniMov
                              - dtRegMovDDMMYYHHMMSS
                              - descrTipoMovimentacao
                              - idMov
                              - duracaoSpan
                              - duracaoSpanExibirCompleto
                              - mimeType
                              - lotaCadastranteSigla
                              - exTipoMovimentacaoSigla
                              - tempoRelativo
                              - podeExibirNoSigale
                              - tipoMovimentacaoArquivamento
                              - acoes
                              - duracao
                            x-apidog-orders:
                              - classe
                              - originadaAqui
                              - desabilitada
                              - cancelada
                              - descricao
                              - exTipoMovimentacao
                              - complemento
                              - parte
                              - dtIniMov
                              - dtRegMovDDMMYYHHMMSS
                              - descrTipoMovimentacao
                              - idMov
                              - duracaoSpan
                              - duracaoSpanExibirCompleto
                              - duracao
                              - mimeType
                              - lotaCadastranteSigla
                              - exTipoMovimentacaoSigla
                              - tempoRelativo
                              - podeExibirNoSigale
                              - tipoMovimentacaoArquivamento
                              - acoes
                        duracoes:
                          type: array
                          items:
                            type: object
                            properties:
                              duracao:
                                type: string
                              span:
                                type: integer
                              spanExibirCompleto:
                                type: integer
                            required:
                              - duracao
                              - span
                              - spanExibirCompleto
                            x-apidog-orders:
                              - duracao
                              - span
                              - spanExibirCompleto
                        ocultar:
                          type: boolean
                        id:
                          type: integer
                        podeTramitar:
                          type: boolean
                        podeAnotar:
                          type: boolean
                        marcadoresEmHtml:
                          type: string
                        descricaoCompleta:
                          type: string
                        acoes:
                          type: array
                          items:
                            type: object
                            properties:
                              icone:
                                type: string
                              nome:
                                type: string
                              descr:
                                type: string
                              nameSpace:
                                type: string
                              acao:
                                type: string
                              pode:
                                type: boolean
                              classe:
                                type: string
                              params:
                                type: object
                                properties:
                                  sigla:
                                    type: string
                                  nomeAcao:
                                    type: string
                                  popup:
                                    type: string
                                  arquivo:
                                    type: string
                                  criandoAnexo:
                                    type: string
                                  mobilPaiSel.sigla:
                                    type: string
                                  assinandoAnexosGeral:
                                    type: string
                                  autuando:
                                    type: string
                                  idMobilAutuado:
                                    type: string
                                  idForma:
                                    type: string
                                  criandoSubprocesso:
                                    type: string
                                  exibirCompleto:
                                    type: string
                                required:
                                  - sigla
                                x-apidog-orders:
                                  - sigla
                                  - nomeAcao
                                  - popup
                                  - arquivo
                                  - criandoAnexo
                                  - mobilPaiSel.sigla
                                  - assinandoAnexosGeral
                                  - autuando
                                  - idMobilAutuado
                                  - idForma
                                  - criandoSubprocesso
                                  - exibirCompleto
                              explicacao:
                                type: string
                              post:
                                type: boolean
                              modal:
                                type: string
                              msgConfirmacao:
                                type: string
                            required:
                              - icone
                              - nome
                              - descr
                              - nameSpace
                              - acao
                              - pode
                              - classe
                              - params
                              - explicacao
                              - post
                              - msgConfirmacao
                              - modal
                            x-apidog-orders:
                              - icone
                              - nome
                              - descr
                              - nameSpace
                              - acao
                              - pode
                              - classe
                              - params
                              - explicacao
                              - post
                              - modal
                              - msgConfirmacao
                      required:
                        - sigla
                        - isGeral
                        - apensos
                        - expedientesFilhosNaoCancelados
                        - processosFilhosNaoCancelados
                        - anexosNaoAssinados
                        - despachosNaoAssinados
                        - expedientesFilhosNaoJuntados
                        - expedientesJuntadosNaoAssinados
                        - pendenciasDeAnexacao
                        - pendenciasDeColaboracao
                        - movs
                        - duracoes
                        - ocultar
                        - id
                        - podeTramitar
                        - podeAnotar
                        - descricaoCompleta
                        - acoes
                      x-apidog-orders:
                        - sigla
                        - isGeral
                        - apensos
                        - expedientesFilhosNaoCancelados
                        - processosFilhosNaoCancelados
                        - anexosNaoAssinados
                        - despachosNaoAssinados
                        - expedientesFilhosNaoJuntados
                        - expedientesJuntadosNaoAssinados
                        - pendenciasDeAnexacao
                        - pendenciasDeColaboracao
                        - movs
                        - duracoes
                        - ocultar
                        - id
                        - podeTramitar
                        - podeAnotar
                        - marcadoresEmHtml
                        - descricaoCompleta
                        - acoes
                  classe:
                    type: string
                  nomeCompleto:
                    type: string
                  dtDocDDMMYY:
                    type: string
                  dataPrimeiraAssinatura:
                    type: string
                  subscritorString:
                    type: string
                  subscritorSigla:
                    type: string
                  subscritorNome:
                    type: string
                  subscritorTipo:
                    type: string
                  originalData:
                    type: string
                  classificacaoDescricaoCompleta:
                    type: string
                  classificacaoSigla:
                    type: string
                  classificacaoNome:
                    type: string
                  tags:
                    type: array
                    items:
                      type: string
                  descrDocumento:
                    type: string
                  nmNivelAcesso:
                    type: string
                  tipoDocumento:
                    type: string
                  dtFinalizacao:
                    type: string
                  nmArqMod:
                    type: string
                  conteudoBlobHtmlString:
                    type: string
                  sigla:
                    type: string
                  fisicoOuEletronico:
                    type: string
                  fDigital:
                    type: boolean
                  cossignatarios:
                    type: object
                    properties: {}
                    x-apidog-orders: []
                  listaOrdenadaCossigSub:
                    type: object
                    properties: {}
                    x-apidog-orders: []
                  forma:
                    type: string
                  modelo:
                    type: string
                  idModelo:
                    type: string
                  tipoFormaDocumento:
                    type: string
                  cadastranteString:
                    type: string
                  lotaCadastranteString:
                    type: string
                  vizTramitacao:
                    type: string
                  vizRelacaoDocs:
                    type: string
                  vizColaboracao:
                    type: string
                  podeAssinar:
                    type: boolean
                  podeCapturarPDF:
                    type: boolean
                  exTipoDocumentoDescricao:
                    type: string
                  podeAnexarArquivoAuxiliar:
                    type: boolean
                  dtPrazoDeAssinatura:
                    type: string
                  acoes:
                    type: array
                    items:
                      type: string
                required:
                  - mobs
                  - classe
                  - nomeCompleto
                  - dtDocDDMMYY
                  - dataPrimeiraAssinatura
                  - subscritorString
                  - subscritorSigla
                  - subscritorNome
                  - subscritorTipo
                  - originalData
                  - classificacaoDescricaoCompleta
                  - classificacaoSigla
                  - classificacaoNome
                  - tags
                  - descrDocumento
                  - nmNivelAcesso
                  - tipoDocumento
                  - dtFinalizacao
                  - nmArqMod
                  - conteudoBlobHtmlString
                  - sigla
                  - fisicoOuEletronico
                  - fDigital
                  - cossignatarios
                  - listaOrdenadaCossigSub
                  - forma
                  - modelo
                  - idModelo
                  - tipoFormaDocumento
                  - cadastranteString
                  - lotaCadastranteString
                  - vizTramitacao
                  - vizRelacaoDocs
                  - vizColaboracao
                  - podeAssinar
                  - podeCapturarPDF
                  - exTipoDocumentoDescricao
                  - podeAnexarArquivoAuxiliar
                  - dtPrazoDeAssinatura
                  - acoes
                x-apidog-orders:
                  - mobs
                  - classe
                  - nomeCompleto
                  - dtDocDDMMYY
                  - dataPrimeiraAssinatura
                  - subscritorString
                  - subscritorSigla
                  - subscritorNome
                  - subscritorTipo
                  - originalData
                  - classificacaoDescricaoCompleta
                  - classificacaoSigla
                  - classificacaoNome
                  - tags
                  - descrDocumento
                  - nmNivelAcesso
                  - tipoDocumento
                  - dtFinalizacao
                  - nmArqMod
                  - conteudoBlobHtmlString
                  - sigla
                  - fisicoOuEletronico
                  - fDigital
                  - cossignatarios
                  - listaOrdenadaCossigSub
                  - forma
                  - modelo
                  - idModelo
                  - tipoFormaDocumento
                  - cadastranteString
                  - lotaCadastranteString
                  - vizTramitacao
                  - vizRelacaoDocs
                  - vizColaboracao
                  - podeAssinar
                  - podeCapturarPDF
                  - exTipoDocumentoDescricao
                  - podeAnexarArquivoAuxiliar
                  - dtPrazoDeAssinatura
                  - acoes
              example:
                mobs:
                  - sigla: irure
                    isGeral: false
                    apensos:
                      - veniam in adipisicing dolor
                    expedientesFilhosNaoCancelados:
                      - esse deserunt veniam
                    processosFilhosNaoCancelados:
                      - Ut reprehenderit fugiat
                    anexosNaoAssinados:
                      - tempor
                    despachosNaoAssinados:
                      - laborum
                    expedientesFilhosNaoJuntados:
                      - voluptate Excepteur laboris dolore
                    expedientesJuntadosNaoAssinados:
                      - sit cupidatat sunt anim irure
                    pendenciasDeAnexacao:
                      - ea anim
                    pendenciasDeColaboracao:
                      - laborum officia
                    movs:
                      - classe: cupidatat esse dolor aliqua
                        originadaAqui: false
                        desabilitada: false
                        cancelada: false
                        descricao: ut id mollit dolore Excepteur
                        exTipoMovimentacao: cillum
                        complemento: dolore qui
                        parte:
                          cadastrante:
                            sigla: incididunt magna sint minim irure
                            siglaAmpliada: occaecat ullamco Ut aute
                            descricao: minim Duis proident
                            descricaoAmpliada: aliquip id irure
                            iniciais: quis
                            siglaOrgao: consectetur est ut
                          lotaCadastrante:
                            sigla: id reprehenderit nulla amet
                            siglaAmpliada: Ut eu est
                            descricao: magna tempor velit
                            descricaoAmpliada: dolore eiusmod cillum
                            iniciais: ullamco amet exercitation elit magna
                            siglaOrgao: velit in aliquip pariatur
                          lotaResp:
                            sigla: dolor enim in in amet
                            siglaAmpliada: consectetur esse reprehenderit adipisicing et
                            descricao: anim mollit
                            descricaoAmpliada: veniam non
                            iniciais: consectetur cupidatat sit eiusmod ullamco
                            siglaOrgao: sit
                          lotaSubscritor:
                            sigla: dolor laborum
                            siglaAmpliada: voluptate eu
                            descricao: dolor
                            descricaoAmpliada: culpa Excepteur nostrud deserunt esse
                            iniciais: elit ea laborum deserunt nostrud
                            siglaOrgao: est id
                          lotaTitular:
                            sigla: laboris
                            siglaAmpliada: veniam consectetur ad id
                            descricao: id ex fugiat nostrud in
                            descricaoAmpliada: anim deserunt adipisicing nisi culpa
                            iniciais: laboris occaecat
                            siglaOrgao: culpa commodo tempor et Lorem
                          resp:
                            sigla: qui fugiat id ex culpa
                            siglaAmpliada: cupidatat Excepteur nisi voluptate sint
                            descricao: officia nulla magna nostrud cillum
                            descricaoAmpliada: in velit
                            iniciais: ullamco
                            siglaOrgao: quis
                          subscritor:
                            sigla: Excepteur non laborum occaecat et
                            siglaAmpliada: nostrud voluptate in occaecat ad
                            descricao: adipisicing
                            descricaoAmpliada: magna sint amet
                            iniciais: nulla amet esse
                            siglaOrgao: magna nostrud culpa exercitation
                          titular:
                            sigla: in
                            siglaAmpliada: cupidatat nisi consectetur qui proident
                            descricao: laborum aliquip
                            descricaoAmpliada: do ipsum
                            iniciais: dolor deserunt veniam anim
                            siglaOrgao: laboris laborum
                        dtIniMov: culpa
                        dtRegMovDDMMYYHHMMSS: dolor non
                        descrTipoMovimentacao: esse
                        idMov: -91354110
                        duracaoSpan: -92674399
                        duracaoSpanExibirCompleto: -78142045
                        duracao: dolor sunt et voluptate exercitation
                        mimeType: ullamco exercitation mollit
                        lotaCadastranteSigla: in sed est id ex
                        exTipoMovimentacaoSigla: elit velit deserunt et
                        tempoRelativo: incididunt laborum tempor
                        podeExibirNoSigale: true
                        tipoMovimentacaoArquivamento: true
                        acoes:
                          - nome: pariatur
                            nameSpace: culpa esse adipisicing ullamco
                            acao: ipsum ullamco
                            pode: true
                            params:
                              popup: irure enim ex
                              sigla: nulla deserunt irure ipsum
                              dtIni: ullamco
                              isTransf: mollit
                            explicacao: consectetur
                            post: true
                    duracoes:
                      - duracao: elit nisi proident sit ut
                        span: 57022568
                        spanExibirCompleto: 89374281
                    ocultar: false
                    id: -92728429
                    podeTramitar: true
                    podeAnotar: false
                    marcadoresEmHtml: do irure cillum fugiat velit
                    descricaoCompleta: Lorem ullamco
                    acoes:
                      - icone: nulla occaecat labore
                        nome: Lorem Duis ad magna ea
                        descr: incididunt
                        nameSpace: in adipisicing consequat do
                        acao: in
                        pode: false
                        classe: ullamco laboris commodo officia in
                        params:
                          sigla: Duis laborum
                          nomeAcao: non consequat
                          popup: veniam
                          arquivo: velit nisi quis
                          criandoAnexo: Lorem eu dolor velit
                          mobilPaiSel.sigla: esse sed proident
                          assinandoAnexosGeral: dolor officia laborum pariatur ut
                          autuando: non minim in in commodo
                          idMobilAutuado: consectetur do
                          idForma: aute ea
                          criandoSubprocesso: elit eu aliqua ad
                          exibirCompleto: commodo
                        explicacao: mollit laborum ad
                        post: true
                        modal: ad velit ut et
                        msgConfirmacao: ut officia in deserunt
                classe: exercitation non magna
                nomeCompleto: minim eu ex et ut
                dtDocDDMMYY: dolor occaecat sit
                dataPrimeiraAssinatura: non anim dolore
                subscritorString: ipsum
                subscritorSigla: ex Lorem dolore proident magna
                subscritorNome: non reprehenderit eu
                subscritorTipo: Ut cillum laboris dolor labore
                originalData: nostrud
                classificacaoDescricaoCompleta: elit cillum sunt tempor
                classificacaoSigla: tempor
                classificacaoNome: incididunt in adipisicing aliqua
                tags:
                  - debitis
                descrDocumento: sunt nulla ipsum velit
                nmNivelAcesso: anim mollit
                tipoDocumento: consectetur Duis sunt in reprehenderit
                dtFinalizacao: incididunt in aute commodo
                nmArqMod: id occaecat
                conteudoBlobHtmlString: cupidatat ut ad tempor voluptate
                sigla: consequat cillum irure
                fisicoOuEletronico: labore
                fDigital: false
                cossignatarios: {}
                listaOrdenadaCossigSub: {}
                forma: elit est mollit cupidatat
                modelo: nisi culpa enim aliquip
                idModelo: minim quis sint ut laborum
                tipoFormaDocumento: laborum Ut id
                cadastranteString: fugiat non Lorem
                lotaCadastranteString: in adipisicing in magna laborum
                vizTramitacao: nulla dolore quis
                vizRelacaoDocs: aliquip sunt
                vizColaboracao: aute labore id do in
                podeAssinar: true
                podeCapturarPDF: true
                exTipoDocumentoDescricao: quis tempor adipisicing
                podeAnexarArquivoAuxiliar: false
                dtPrazoDeAssinatura: ipsum Lorem reprehenderit
                acoes:
                  - in ipsum Ut dolore deserunt
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
      security:
        - bearer: []
      x-apidog-folder: SIGADOC/SIGAEX/consulta
      x-apidog-status: released
      x-run-in-apidog: https://app.apidog.com/web/project/463021/apis/api-6389681-run
components:
  schemas: {}
  securitySchemes:
    bearer:
      type: http
      scheme: bearer
servers:
  - url: http://dev-seguro.xvia.mti.mt.gov.br/r1/mt-dev/GOV/MTI/SIGADOC
    description: DEV_SIGADOC
security: []

```
