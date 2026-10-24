# Planejamento — Documentos obrigatórios no cadastro do projeto

> Status: **fases 2 e 3 implementadas** (marcação e continuidade, com as regras atuais) · fases 1, 4 e 5 pendentes · Criado em 24/10/2026
> Decisões adotadas: a justificativa de não envio cumpre o obrigatório; o bloqueio vale só para o botão "Continuar" (as abas seguem livres e o envio é travado no servidor); a obrigatoriedade só muda por migração versionada.
> Escopo: aba **Documentos** do cadastro de projeto (etapa 5 do assistente). Marcar o que é obrigatório e o que não é, liberar a continuidade quando os obrigatórios estiverem cumpridos e definir **como as regras de obrigatoriedade serão identificadas e mantidas**.
> Complementa a seção 7 de [planejamento-cadastro-projeto.md](planejamento-cadastro-projeto.md). Fora do escopo: documentos posteriores à análise técnica favorável (conta bancária) e a conferência do analista.

---

## 1. Como está hoje

### 1.1 Backend (`backend/internal/project`)
- O catálogo `document_requirements` já é **guiado por dados**: cada exigência tem `instruments` (convênio/fomento), `conditions` (E lógico sobre fatos do projeto), `scope` (projeto ou por emenda) e `accepts_justification`. A carga inicial está em `requirements.go` (`defaultRequirements`, 45 itens).
- O checklist é calculado em `checklist()` ([documents.go:154](../backend/internal/project/documents.go)): entram as exigências cuja condição casa com as respostas do projeto. **Todo item que entra é, na prática, obrigatório**, mas isso não está dito em lugar nenhum: não existe o conceito de "opcional".
- Um item está cumprido se tiver **arquivo confirmado** (`done`) **ou justificativa de não envio** (`justified`, Anexo X/XVII). Senão é `missing`.
- `documentIssue()` ([documents.go:234](../backend/internal/project/documents.go)) cria a pendência **bloqueante** `documents_missing` quando `Summary.Missing > 0`. O envio ([submit.go:112](../backend/internal/project/submit.go)) recusa o projeto com qualquer pendência bloqueante. Ou seja, **o bloqueio do envio já funciona**; o que falta é a marcação e a continuidade na tela.
- `EnsureDefaults` só insere o que **ainda não existe** (`ON CONFLICT DO NOTHING`). Alterar uma exigência já carregada exige migração.

### 1.2 Frontend (`frontend/src/features/projects`)
- [documents-step.tsx](../frontend/src/features/projects/components/documents-step.tsx) mostra "X de Y documentos entregues" e a barra de progresso, agrupados por seção. [checklist-item.tsx](../frontend/src/features/projects/components/checklist-item.tsx) mostra só o selo da fonte (C2, F4…) e a situação (Pendente / Anexado / Justificado).
- **Nada diz se o documento é obrigatório.** Um item "Pendente" parece igual a qualquer outro.
- **A aba Documentos não tem botão de continuar.** É a única etapa sem `StepActions`; o proponente só avança clicando na aba seguinte. O rodapé do assistente ainda diz "Os documentos e o envio do projeto vêm nas próximas fases" (texto desatualizado, [project-wizard.tsx](../frontend/src/features/projects/components/project-wizard.tsx)).

---

## 2. Onde as regras de obrigatoriedade precisam ser revistas

Lendo os checklists da SEDEC ([prefeituras/convênio](www-sedec-mt-gov-br-conv-C3-AAnios.md) e [OSC/fomento](www-sedec-mt-gov-br-termo-de-fomento.md)) contra o catálogo atual, aparecem qualificadores que o catálogo ainda não representa. O critério de leitura está na seção 4.

| Item na norma | Qualificador | Como está no catálogo | Proposta |
|---|---|---|---|
| F2.1 Ofício da emenda | "quando aplicável" | Condicional (`hasAmendments`) | Já correto |
| F13 Uso de espaço público | "quando aplicável" | Condicional (opção do projeto) | Já correto |
| C20 Redução de prazo | "quando aplicável" | Condicional (opção do projeto) | Já correto |
| C7.2 / F8 Exclusividade | "em caso de" | Condicional (opção do projeto) | Já correto |
| C6 Contrapartida | dispensada em calamidade | Condicional (`calamity`) | Já correto |
| **C21 a 21.12 Projeto básico da obra** | 21.5 "se houver"; 21.6 "quando exigido"; 21.7 "ou dispensa"; 21.8 "no caso de restauração" | **Um único item** `projeto_basico_obra` para os 13 documentos | Desdobrar (seção 5.4) |
| **C22 Termo de Referência** | "quando envolver aquisição de bens ou serviços" | `termo_referencia` vale **sempre** (C5/F5), inclusive em obra pura | Perguntar à SEDEC (seção 7, pergunta 3) |
| C16 Diploma do prefeito | só prefeito | Condicional (`entityTypeCode`) | Já correto |
| C6.1 / C6.2 QDD e LOA | só entidade pública | Condicional (`entityNature`) | Já correto |
| F25 / C Anexo X Justificativa de não envio | "quando aplicável" | Não é item: é o mecanismo de justificativa | Já correto |
| Declarações geradas (5) | obrigatórias, mas cumpridas por **PDF assinado no MT-ID** | Item com painel de declaração | Ver seção 6.3 |

Resumo: das 45 exigências, a grande maioria é **obrigatória sempre que aplicável**, e a aplicabilidade já está bem modelada por condições. O que falta é (a) um nível **opcional** para os poucos itens "se houver" e (b) o desdobramento do projeto básico da obra.

---

## 3. Modelo de regra proposto

Dois mecanismos separados, cada um com um significado:

1. **Aplicabilidade** (já existe): `instruments` + `conditions`. Responde "este documento entra na lista deste projeto?". Se não entra, não aparece.
2. **Obrigatoriedade** (novo): `mandatory boolean NOT NULL DEFAULT true`. Responde "dos que entram, este bloqueia o envio?". Itens `false` aparecem marcados como **Opcional** e nunca bloqueiam.

Preferir sempre a aplicabilidade quando o proponente consegue responder ("o projeto tem climatização?"): vira uma pergunta, o item entra e é obrigatório. Usar `mandatory = false` só quando **só o analista consegue saber** se o documento é exigido (ex.: projeto de incêndio "quando exigido" pelo Corpo de Bombeiros).

| Nível | Aparece? | Bloqueia continuar/enviar? | Selo na tela |
|---|---|---|---|
| Não se aplica | Não | — | — |
| Opcional | Sim | Não | **Opcional** |
| Obrigatório | Sim | Sim, até ter arquivo ou justificativa | **Obrigatório** |

**Justificativa de não envio cumpre o obrigatório?** Recomendo **sim** (é o que a Portaria 247, art. 2º III, e o Anexo X/XVII preveem, e o que o envio já faz). Mas o item cumprido por justificativa fica visualmente distinto ("Justificado") e o resumo conta separado, porque o analista pode devolver (seção 7, pergunta 1).

Itens que **não aceitam** justificativa (`accepts_justification = false`) só cumprem com arquivo. Hoje nenhum está marcado assim; a matriz da seção 4 deve dizer quais deveriam estar.

---

## 4. Como identificar as regras (o processo)

Entregável: **`docs/matriz-documentos-obrigatorios.md`**, uma tabela por instrumento, uma linha por documento, validada pela SEDEC antes de virar carga no banco.

### 4.1 Fontes, em ordem de autoridade
1. Checklists oficiais da SEDEC (os dois arquivos em `docs/`), que são o que o analista usa na conferência.
2. IN Conjunta 001/2015 (art. 4º III e art. 8º) para convênio.
3. IN Conjunta 001/2016 (art. 7º, 28 e 29), alterada pelas INs 004/2024 e 002/2026, para fomento.
4. Lei 13.019/2014 (arts. 33 a 35 e 39) para fomento.
5. Lei 12.082/2023 (evento com artistas) e Portaria Conjunta 71/2026 (efeito da falta de obrigatórios: arquivamento sem análise).
6. Portaria 247/2024 (fluxo e justificativa da ausência).

### 4.2 Regra de classificação (como ler cada linha da norma)
| O texto da norma diz | Classificação |
|---|---|
| Item sem qualificador | Obrigatório sempre que o instrumento coincidir |
| "quando aplicável", "em caso de", "no caso de", "se houver" e o proponente sabe responder | **Condição** (nova pergunta ou fato) e obrigatório quando casar |
| "quando exigido", "se houver" e só o analista sabe | **Opcional** (`mandatory = false`) com a descrição explicando quando enviar |
| "ou dispensa" / "ou justificativa" | Obrigatório, com justificativa aceita |
| Item após a análise técnica favorável | Fora do cadastro |

### 4.3 Colunas da matriz
`código` · `documento` · `item C#/F#` · `norma e artigo` · `texto literal do qualificador` · `nível` (obrigatório / opcional) · `condição de aplicação` · `fato necessário (existe ou é novo)` · `aceita justificativa` · `modelo (Anexo)` · `validado pela SEDEC (data e quem)`.

### 4.4 Passos
1. **Extrair**: gerar a matriz inicial a partir de `defaultRequirements` (já traz código, instrumentos, condições e fonte) e acrescentar as colunas de norma e qualificador.
2. **Classificar** cada linha com a regra da 4.2 e marcar as dúvidas.
3. **Fechar as lacunas** da seção 2 e levar à SEDEC (Coordenadoria de Convênios) a lista de perguntas da seção 7.
4. **Congelar** a matriz com a validação e só então escrever a migração e a carga (seção 5).
5. **Manter**: toda mudança de norma altera primeiro a matriz, depois a carga. Um teste compara a matriz com `defaultRequirements` para não divergirem (seção 8).

---

## 5. Mudanças no backend

### 5.1 Dados
- Migração `00020`: `ALTER TABLE document_requirements ADD COLUMN mandatory boolean NOT NULL DEFAULT true`. Todos os itens atuais continuam obrigatórios, então **nada muda para quem já tem rascunho**.
- `reqDef` ganha `optional bool`; `InsertDefaultDocumentRequirement` e o sqlc passam a gravar `mandatory`.
- Como `EnsureDefaults` não altera linhas existentes, as mudanças de nível, o desdobramento da obra e as novas condições entram **na própria migração** (`UPDATE`/`INSERT` explícitos), e o `defaultRequirements` é atualizado para refletir o mesmo estado (instalações novas).

### 5.2 Checklist e resumo
- `ChecklistItem` ganha `Mandatory bool`.
- `DocumentSummary` ganha `RequiredTotal` e `RequiredMissing` (mantém `Total/Done/Justified/Missing`, para não quebrar o que já os usa).
- `documentIssue` passa a usar `RequiredMissing` e a mensagem muda para "Faltam N de M documentos **obrigatórios**…". Itens opcionais pendentes **não** geram pendência.
- Como o envio já bloqueia por pendência bloqueante, **nenhuma alteração em `submit.go`** é necessária. O retrato imutável do envio guarda `mandatory` por item (o analista vê o que era obrigatório na data).
- Complementação (`missingIn`) segue igual: o analista devolve o item que quiser, obrigatório ou não.

### 5.3 API e documentação
- Atualizar `backend/api/openapi.yaml`: `ChecklistItem.mandatory`, `DocumentSummary.requiredTotal/requiredMissing`, e o texto do `documents` no projeto (o teste de cobertura de rotas não muda, pois não há rota nova).

### 5.4 Desdobrar o projeto básico da obra
`projeto_basico_obra` (C21 a 21.12) vira itens separados no grupo "Obras e serviços de engenharia", porque cada um é um arquivo diferente e com regra diferente:

| Item | Nível | Como decidir |
|---|---|---|
| 21.1 Arquitetônico executivo, 21.2 Elétrico, 21.3 Hidrossanitário, 21.4 Estrutural, 21.9 Memorial, 21.10 Planilha, 21.11 Cronograma físico-financeiro, 21.12 ART/RRT | Obrigatório | Sem qualificador |
| 21.5 Climatização | **Condição** `hasHvac` | "se houver": pergunta nova nas situações do projeto |
| 21.8 Mapa de danos | **Condição** `isRestoration` | "no caso de restauração": pergunta nova (ou valor novo em `objectNature`) |
| 21.6 Prevenção e combate a incêndio | **Opcional** | "quando exigido": quem decide é o Corpo de Bombeiros/analista |
| 21.7 SPDA | Obrigatório, **aceita justificativa** | "ou dispensa": a dispensa é a justificativa |

As perguntas novas entram em `DocumentOptions` junto das três atuais (`facts()` e `factFields` em `requirements.go`, `UpdateProjectDocumentOptions`, colunas em `projects`). Projetos de obra já em rascunho com arquivos no item antigo precisam de **migração dos arquivos** para um dos novos códigos (por exemplo, o item antigo vira `projeto_arquitetonico`; o proponente redistribui o resto). A chave estrangeira `project_document_items.requirement_code` é `ON DELETE RESTRICT`, então o código antigo não pode ser removido enquanto houver itens; ele fica `active = false`.

---

## 6. Mudanças no frontend

### 6.1 Marcação
- [checklist-item.tsx](../frontend/src/features/projects/components/checklist-item.tsx): selo **Obrigatório** (preenchido) ou **Opcional** (contorno) ao lado do selo da fonte. O texto do selo é o que informa; cor sozinha não basta (acessibilidade).
- Item opcional pendente não usa o rótulo "Pendente" e sim "Não enviado" em tom neutro, para não parecer erro.
- Ordem dentro de cada grupo: obrigatórios pendentes primeiro? **Não**: manter a ordem da norma (é a que o proponente confere contra o checklist da SEDEC). Em vez disso, um filtro "Mostrar só o que falta".

### 6.2 Resumo e continuidade
- [documents-step.tsx](../frontend/src/features/projects/components/documents-step.tsx): o progresso passa a contar **obrigatórios**: "X de Y obrigatórios cumpridos" (e "Z opcionais"), mantendo o detalhe "N com justificativa".
- Novo rodapé da aba: **"Continuar para Revisão e envio"**.
  - Habilitado quando `requiredMissing === 0`.
  - Desabilitado com o motivo visível (não só `disabled`): "Faltam N documentos obrigatórios" com a lista de nomes, cada um levando ao item (âncora), ligada ao botão por `aria-describedby`.
  - Ao continuar, navega para a etapa `submission` (mesmo comportamento de "Salvar e continuar" das outras etapas; aqui não há o que salvar, pois cada arquivo já grava ao anexar).
- O selo numérico da aba Documentos ([project-wizard.tsx](../frontend/src/features/projects/components/project-wizard.tsx)) passa a mostrar `requiredMissing`.
- Remover o texto desatualizado "Os documentos e o envio do projeto vêm nas próximas fases".
- Tipos em `api/types.ts`: `ChecklistItem.mandatory`, `DocumentSummary.requiredTotal/requiredMissing`.

### 6.3 Declarações assinadas
As cinco declarações geradas só ficam `done` com o PDF assinado no MT-ID, que hoje é **simulado** ([debito-tecnico-mt-id.md](debito-tecnico-mt-id.md)). Com o cliente real, o proponente sem MTCidadão vinculado não conseguiria cumprir um obrigatório. É preciso decidir o plano B (seção 7, pergunta 4) antes de ligar o bloqueio em produção.

### 6.4 Servidor é a autoridade
O botão é conveniência. A regra real continua no envio (`submit.go`) e a tela só reflete `requiredMissing`, nunca recalcula a obrigatoriedade no navegador.

---

## 7. Perguntas

**Para o órgão/SEDEC** (bloqueiam a matriz):
1. A justificativa de não envio (Anexo X/XVII) cumpre o item obrigatório para **liberar o envio**? Recomendação: sim, com sinalização para o analista.
2. Algum obrigatório **não** aceita justificativa (ex.: Certidão de Habilitação Plena)? A IN 001/2016 art. 8º e a habilitação no SIGCon sugerem que não.
3. **Termo de Referência** (C5/F5 sem qualificador e C22 "quando envolver aquisição de bens ou serviços"): em obra pura sem bens/serviços o TR é dispensado?
4. **Declaração de contrapartida** de entidade que não é município (convênio): é obrigatória? O modelo só existe para município; hoje a declaração gerada é bloqueada para outras entidades ([planejamento-declaracoes.md](planejamento-declaracoes.md), pergunta 6).
5. **Projeto de incêndio (21.6)**: quem decide a exigência? Opcional com orientação é aceitável?
6. **Convênio com entidade privada** usa o checklist de prefeituras (decisão 5). Itens como "Cartão do CNPJ da Prefeitura" e "RG do Prefeito" valem como "do dirigente"? Há itens da lista de prefeituras que não fazem sentido para entidade privada além dos já condicionados?

**Para você** (decisões de produto, com a recomendação):
- **Continuar bloqueia a navegação ou só o botão?** Recomendo só o botão: as abas continuam livres, como nas outras etapas, e o envio segue travado no servidor. Se quiser bloquear também a aba "Revisão e envio", é uma linha no `go()` do assistente.
- **Quem pode mudar a obrigatoriedade?** Recomendo só por migração versionada nesta entrega; uma tela de administração do catálogo (hoje inexistente) fica para depois.

---

## 8. Testes

Backend (`internal/app/project_documents_test.go` e `internal/project`):
- Os testes atuais que esperam `missing = 22/23` passam a afirmar também `requiredMissing`; revisar os números depois do desdobramento da obra.
- Item opcional pendente **não** gera `documents_missing` e **não** bloqueia o envio.
- Obrigatório pendente bloqueia; obrigatório com arquivo confirmado cumpre; com justificativa cumpre e aparece em `justified`; justificativa em item que não a aceita é 422.
- Obra: cada subitem aparece só quando sua condição casa (climatização, restauração); incêndio é opcional; SPDA aceita justificativa.
- Alterar uma resposta que remove um item aplicável não deixa "órfão" bloqueando o envio.
- Convênio × fomento × (público/privado) × (com/sem emenda): a lista de obrigatórios bate com a matriz. Um teste lê a matriz (`docs/matriz-documentos-obrigatorios.md` ou um JSON gerado dela) e compara com `defaultRequirements` e com a migração.
- Retrato do envio guarda `mandatory`.

Frontend (não há suíte automatizada; verificar no navegador com o ambiente de desenvolvimento):
- Selos Obrigatório/Opcional corretos nos dois instrumentos.
- Botão desabilitado com o motivo e a lista; habilita ao anexar o último obrigatório; habilita ao justificar.
- Marcar "obra" e "restauração" faz aparecerem os itens novos e desabilita o botão de novo.
- Leitor de tela anuncia o motivo do botão desabilitado.
- Projeto compartilhado só com leitura: vê os selos, não vê ações.

---

## 9. Entrega em fatias

| Fase | Conteúdo | Pronto quando |
|---|---|---|
| 1 | Matriz (`docs/matriz-documentos-obrigatorios.md`) e perguntas à SEDEC | Matriz completa; perguntas enviadas |
| 2 | Backend: coluna `mandatory`, resumo, `documentIssue`, OpenAPI, testes | `requiredMissing` correto e opcional não bloqueia. **Feito** (migração 00020) |
| 3 | Frontend: selos, resumo por obrigatórios, botão "Continuar" com motivo | Fluxo completo no navegador. **Feito**; falta a verificação no navegador (ver 9.1) |
| 4 | Desdobrar o projeto básico da obra e as perguntas novas | Obra mostra só o que se aplica |
| 5 | (Depois) Plano B das declarações sem MT-ID; tela de admin do catálogo; vigência por versão de norma | Decisão registrada |

As fases 2 e 3 já entregam o que foi pedido (marcação + continuidade) usando as regras atuais, em que tudo é obrigatório. As fases 1 e 4 refinam **quais** são obrigatórios e não bloqueiam as anteriores.

---

## 10. Riscos
- **Mudar a obrigatoriedade de um item muda o que bloqueia o envio** de rascunhos existentes. Torná-lo opcional libera; torná-lo obrigatório novo trava. Por isso itens novos obrigatórios (desdobramento da obra) precisam de aviso e a migração dos arquivos antigos.
- **Rótulo errado de "Opcional" induz o proponente a não enviar** e a Portaria Conjunta 71, art. 1º §2º, arquiva a proposta sem análise se faltar obrigatório. Na dúvida, manter obrigatório e deixar o analista dispensar.
- **Catálogo sem versão por vigência**: o planejamento original prevê versionamento, mas ele não existe. Enquanto não existir, uma mudança de norma afeta também os rascunhos em andamento.

---

## 9.1 O que foi feito nas fases 2 e 3

- **Migração `00020_document_requirement_mandatory.sql`:** coluna `mandatory` (padrão `true`). Nenhuma exigência atual virou opcional; o catálogo continua como estava até a matriz ser validada. `reqDef.optional` marca um item como opcional na carga inicial. O código gerado pelo sqlc (`models.go`, `project_documents.sql.go`) foi editado à mão porque o `sqlc` não está instalado neste ambiente; rodar `make sqlc` onde ele existir não deve alterar nada.
- **`documents.go`:** `ChecklistItem.Mandatory`; `DocumentSummary.RequiredTotal/RequiredMissing`; a contagem agora está em `Checklist.tally()`, usada também no ajuste de declaração desatualizada (`declaration_service.go`), no lugar dos contadores ajustados à mão. `documentIssue` usa `RequiredMissing`.
- **OpenAPI** atualizado (`mandatory`, `requiredTotal`, `requiredMissing`).
- **Frontend:** selo Obrigatório/Opcional em cada item; "Não enviado" (neutro) para opcional sem arquivo; progresso e texto contam os obrigatórios; filtro "Mostrar só o que falta"; rodapé com a lista do que falta (cada nome leva ao item) e o botão "Continuar para Revisão e envio", desabilitado enquanto faltar obrigatório e ligado à lista por `aria-describedby`; o selo da aba Documentos mostra `requiredMissing`; removido o texto "vêm nas próximas fases".
- **Testes novos** (`project_documents_test.go`): todos obrigatórios por padrão; opcional fora da contagem e sem a pendência `documents_missing`.
- **Falta:** conferir no navegador (selos, botão desabilitado com o motivo e liberado depois de anexar/justificar, leitor de tela). O login é pelo MT Login e não foi exercitado nesta sessão.
