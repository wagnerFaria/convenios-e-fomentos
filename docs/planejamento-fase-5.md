# Planejamento detalhado — Fase 5: telas da equipe

Parte de [planejamento-fluxo-analise-enquadramento.md](planejamento-fluxo-analise-enquadramento.md) (seção 10) e do [BPMN](bpmn/convenios-e-fomento.bpmn). Todo o backend do fluxo já existe (fases 1 a 4): esta fase é **só frontend** e consome a API como ela está. Sem código aqui: decisões, telas, contratos de uso, ordem de trabalho e como validar.

**Estado: implementada** (passos P5.1 a P5.9 commitados; P5.10 é este conjunto de documentos). F1 a F8 aceitas como recomendadas; **falta o roteiro de aceite no navegador, que só você pode percorrer** ([roteiro-aceite-fluxo-equipe.md](roteiro-aceite-fluxo-equipe.md)).

**Resultado da fase:** cada perfil de equipe tem a sua fila e consegue fazer, pela interface, o que o BPMN prevê para ele, sem abrir mão do que a API já recusa:

| Perfil | Na tela |
|---|---|
| **Analista** | Fila (aguardando análise, em análise, em enquadramento, aguardando proponente, aguardando parecer); assume a análise; abre os documentos e aprova item a item; devolve para complementação; arquiva por falta de documentos; indica o departamento (só com tudo aprovado); vê o histórico das recusas |
| **Revisor do Departamento** | Fila "Enquadramentos pendentes" (só do departamento dele); lê o projeto e os documentos; confirma ou recusa com justificativa |
| **Secretário de Estado** | Vê todos os projetos enviados; nega com motivo, em qualquer etapa |
| **Supervisor Administrador** | Acompanha: vê tudo, sem nenhuma ação |
| **Proponente** | Sem tela nova; a etapa "Aguardando parecer" aparece com o rótulo certo |

---

## 1. O que existe hoje no frontend

- `features/projects/` (api → service → queries + componentes) já cobre o fluxo do **Proponente**: lista (`projects-list.tsx`), assistente por etapas (`project-wizard.tsx`), resumo e comprovante (`review-step.tsx`), pedido de complementação visto pelo proponente (`complement-banner.tsx`), documentos com download (`checklist-item.tsx`, `getFileDownload`).
- A equipe **já consegue abrir** `/dashboard/projetos/{id}`: `access` vazio faz o assistente abrir em modo somente leitura (`accessOf` com `base = false`). Faltam as **ações** e a **conferência**.
- O menu "Meus projetos" exige `admin::projects.read`, que agora a equipe também tem: sem mudança, **a equipe veria todos os projetos sob o título "Meus projetos"**.
- Tipos desatualizados: `ProjectStage` sem `awaiting_opinion`; sem `review`, `analyst`, `framing(s)` na visão nem `department`/`framingStatus`/`analyst` na linha; `Department` sem `reviewerCount` e `projectCount`; `projectActions` sem as ações novas; `listProjects` só aceita `q` e `page`.
- Convenções a seguir (`frontend/CLAUDE.md`): React Query com `api/types → service → queries`; filtros de URL com **nuqs**; ícones só de `@/components/icons`; formulários com `useAppForm` e Zod; cabeçalhos pelo `PageContainer`; permissões com `useCan()`/`<Can>`; formatação do `oxfmt`. Não há suíte de testes automatizados no frontend.

---

## 2. Decisões

### 2.1 Já tomadas
O fluxo, os perfis e as regras da API (fases 1 a 4). A tela **espelha** as regras do servidor só como conveniência; **o servidor é a autoridade** e cada erro dele é mostrado em português.

### 2.2 Novas, com recomendação adotada

| # | Decisão | Recomendação | Alternativa |
|---|---|---|---|
| F1 | Onde fica a fila da equipe | **Página nova** `/dashboard/analise` ("Análise de projetos"), com abas conforme a permissão. "Meus projetos" passa a exigir `admin::projects.create` (só quem cadastra). Quem acumula Proponente e Analista vê os dois itens | Uma página só que muda de conteúdo conforme a permissão |
| F2 | Detalhe do projeto | **A mesma rota** `/dashboard/projetos/{id}` para todos; a visão da equipe aparece quando `project.access` é vazio. A fila leva para ela | Rota separada para a equipe |
| F3 | Conferência | **Painel próprio** ("Conferência"), na primeira posição das abas do detalhe da equipe, listando seções, documentos e adicionais com as ações. As abas do cadastro seguem somente leitura, como hoje. Não altero os componentes do proponente | Embutir os botões de aprovar em cada etapa do cadastro |
| F4 | Abrir documento | Reaproveitar o fluxo do `checklist-item.tsx` (pede o link à API e abre em nova aba). Depois de pedir, **recarregar o projeto** para o selo "aberto" aparecer | Baixar dentro da página |
| F5 | Contadores nas abas da fila | **Sim**, uma consulta leve por aba (`pageSize=1` e `meta.total`), só para as abas do perfil | Sem contadores |
| F6 | Visão em quadro (kanban) | **Não** nesta fase (o componente `ui/kanban.tsx` existe, e as abas já dão a fila) | Quadro por etapa |
| F7 | Atualização | Sem tempo real. Filas e detalhe recarregam ao voltar para a janela (padrão do React Query). Conflito (409/403 de quem chegou depois) → aviso e recarga automática | Atualização periódica |
| F8 | Verificação | `typecheck`, lint e formato **sem problema novo** (ver linha de base da Fase 0) + roteiro manual por perfil (seção 8). Não introduzir ferramenta de teste de interface sem você pedir | Adicionar Vitest/Playwright |

---

## 3. Contrato de dados no frontend

### 3.1 `features/projects/api/types.ts`
- `ProjectStage` ganha `'awaiting_opinion'`; `stageLabels` ganha **"Aguardando parecer"**.
- `projectActions` ganha `startAnalysis`, `reviewItem`, `requestComplement`, `archive`, `deny`, `assignDepartment`, `reviewFraming` (os mesmos nomes de ação do backend).
- Tipos novos: `ReviewKind` (`section | document | extra`), `ReviewItem` (kind, target, label, status, blocking, reason?, reviewedBy?, reviewedAt?, files[{id, filename, opened}]), `Review` (items, summary{total, approved, pending, blockers}, allApproved), `AnalystInfo`, `DepartmentRef`, `Framing` (round, department, decision, indicatedBy, indicatedAt, decidedBy?, decidedAt?, justification?).
- `Project` ganha `analyst?`, `review?`, `framing?`, `framings?` (opcionais: só a equipe recebe). `ProjectSummary` ganha `department?`, `framingStatus?`, `analyst?`.
- `FramingStatus = 'pending' | 'confirmed'`; rótulos pt-BR de enquadramento ("Aguardando decisão do departamento", "Enquadrado") e de decisão de rodada ("Pendente", "Confirmado", "Recusado").

### 3.2 `service.ts`
`startAnalysis(id)`, `reviewItem(id, {kind, target, approved})`, `requestComplement(id, {message, items[]})`, `archiveProject(id, note)` (motivo fixo `missing_documents`), `denyProject(id, note)`, `assignDepartment(id, departmentId)`, `confirmFraming(id, round)`, `refuseFraming(id, round, justification)`. `listProjects` aceita `stage[]`, `framing`, `department`, `analyst`. Todas devolvem o projeto completo (a API responde assim), então o cache do detalhe é atualizado com a resposta, sem nova consulta.

### 3.3 `queries.ts`
`projectsQueryOptions` já recebe os parâmetros; a chave passa a incluí-los (a fábrica `projectKeys.list(params)` já faz isso). Consultas dos contadores reaproveitam a mesma função com `pageSize: 1`. `departmentsQueryOptions({ finalistic })` na feature de departamentos.

### 3.4 Departamentos
`Department` ganha `reviewerCount` e `projectCount`; `listDepartments({ finalistic })` envia `finalistic=true`. O Analista **já tem** `departments.read`, então consegue listar para o seletor.

---

## 4. Telas

### 4.1 Fila da equipe — `/dashboard/analise`
- `PageContainer` com `access` pelo critério **"é equipe de projetos"** (seção 5), conferido no servidor, como pede `docs/nav-rbac.md`.
- **Abas por perfil**, estado na URL com nuqs (`aba`, `q`, `pagina`):
  - *Analista:* Aguardando análise (`stage=awaiting_checklist`) · Em análise (`stage=in_analysis&framing=none`) · Em enquadramento (`framing=pending`) · Aguardando proponente (`stage=awaiting_proponent`) · Aguardando parecer (`stage=awaiting_opinion`) · Arquivados (`status=archived`) · Todos.
  - *Revisor:* Pendentes (`framing=pending`) · Confirmados (`framing=confirmed`).
  - *Secretário e Supervisor:* Todos · Em análise · Aguardando parecer · Arquivados.
- Linha: título e código, instituição (CNPJ formatado), instrumento, valor, **etapa** e, quando houver, **enquadramento e departamento**, **analista responsável**, data de atualização e "Abrir". Selos usam texto, nunca só cor.
- Busca por título, instituição, CNPJ ou nº do SIGCon (parâmetro `q` que a API já tem). Paginação como a lista atual.
- Filtro de departamento (lista de departamentos) para Analista, Secretário e Supervisor; o Revisor não precisa (a API já limita).
- Contadores por aba (F5).

### 4.2 Detalhe do projeto para a equipe — `/dashboard/projetos/{id}`
Quando `project.access` é vazio (equipe), o `ProjectWizard` renderiza uma **visão da equipe**:

1. **Cabeçalho:** título, código, instrumento, selos de situação, **etapa**, **enquadramento** (departamento e rodada), **analista responsável**. Os botões do proponente (compartilhar, duplicar, desistir) já se escondem sozinhos por `isOwner`/`can`.
2. **Barra de ações** (só as que cabem no estado e na permissão do usuário):

| Ação | Quem | Quando aparece | Efeito |
|---|---|---|---|
| Iniciar análise | Analista | `awaiting_checklist` | `start-analysis` |
| Devolver para complementação | Analista | `awaiting_checklist` ou `in_analysis`, sem enquadramento pendente | Diálogo: mensagem opcional + escolha dos itens (seções e documentos, **não** adicionais) com observação por item; ao menos 1 |
| Arquivar por falta de documentos | Analista | idem | Diálogo com motivo (3 a 1.000 caracteres) |
| Indicar departamento | Analista | `in_analysis`, sem pendência de enquadramento, **tudo aprovado** | Diálogo com seletor dos departamentos finalísticos; os sem Revisor ativo aparecem **desabilitados** com a explicação |
| Confirmar enquadramento | Revisor | enquadramento pendente | `confirm` com a rodada vista |
| Recusar enquadramento | Revisor | idem | Diálogo com justificativa (3 a 1.000) |
| Negar projeto | Secretário | `submitted` | Diálogo com motivo e aviso de que **arquiva na hora e avisa o proponente, a equipe e os Revisores** |

   A barra **explica** o que está bloqueado em vez de só esconder (por exemplo, "Aprove todos os itens para indicar o departamento" com o número de pendências), seguindo `framing_pending`, `review_incomplete` e afins.
3. **Abas:** *Conferência* (primeira para quem confere), *Cadastro* (as 6 etapas atuais, somente leitura), *Histórico* (rodadas de enquadramento com a justificativa, pedido de complementação e arquivamento).
4. **Painel de conferência (`ReviewPanel`):** barra de progresso (aprovados de total) e quatro grupos: Seções, Documentos, Adicionais e **Impedimentos** (documento obrigatório vazio, com a razão e a lembrança de que é caso de devolver ou arquivar). Cada item mostra rótulo, selo (Pendente/Aprovado, com quem e quando), e:
   - seção: "Ver" (vai para a aba do cadastro correspondente) e **Aprovar / Desfazer**;
   - documento ou adicional com arquivo: para **cada arquivo**, "Abrir" e o selo **Aberto / Não aberto** (do usuário); **Aprovar** fica habilitado só depois de abrir todos, com a dica do que falta (o servidor confirma);
   - documento só com justificativa: mostra o texto e **Aprovar**.
   Sem a permissão `review-item`, ou fora de `in_analysis`, ou com enquadramento pendente, o painel é **somente leitura** (vê o progresso, não mexe).
5. **Painel de enquadramento (`FramingPanel`):** rodada em curso (departamento, quem indicou, quando) e o histórico, da mais antiga para a mais nova, com a justificativa de cada recusa. Para o Revisor, os botões de decidir.

### 4.3 Lado do Proponente
- `awaiting_opinion` entra nos rótulos da lista e do cabeçalho ("Aguardando parecer"). Um aviso curto no detalhe: "O projeto foi enquadrado e aguarda o parecer da equipe." Nenhuma informação interna (departamento, analista, conferência) chega ao proponente: a API já não envia.

### 4.4 Administração
- **Departamentos:** colunas **Revisores ativos** e **Projetos**; selo "Sem revisor" nos finalísticos com zero Revisores (explica por que não podem ser indicados); a exclusão mostra o motivo quando há projetos.
- **Usuários:** aviso ("Revisor do Departamento sem departamento não enxerga nenhum projeto") ao conceder o perfil Revisor a alguém sem departamento (diálogos de novo administrador e de liberação de acesso e tela de detalhe). O código do perfil (`department-reviewer`) vem da lista de perfis que a tela já carrega.

---

## 5. Navegação

O menu aceita **uma** permissão por item, e o Proponente e a equipe têm a mesma ação `admin::projects.read`. Mas a sessão (`GET /auth/me`) traz, para cada permissão, as **condições** dela, e é isso que separa os dois: a leitura do Proponente só tem as condições "é o criador" e "é membro/editor do projeto"; a da equipe tem `admin::project-submitted` ou `admin::project-submitted-same-department` (ou nenhuma condição, no caso de um perfil personalizado que lê tudo). Sem criar permissão nova nem mexer no backend:

- **"É equipe de projetos"** = o Super Admin, ou quem tem `admin::projects.read` com alguma condição **fora** de {`is-creator`, `project-member`, `project-editor`} (ou sem condição nenhuma).
- **"É proponente"** = o Super Admin, ou quem tem `admin::projects.create`.

| Item de menu | Aparece para | Antes |
|---|---|---|
| Meus projetos | "é proponente" | exigia `read`: a equipe passaria a vê-lo com todos os projetos |
| **Análise de projetos** (novo, grupo Projetos) | "é equipe de projetos" | — |

Quem acumula Proponente e Analista vê os dois itens (as permissões se unem, e as condições também). O Supervisor, que só lê, entra pelo mesmo item "Análise de projetos" e não vê nenhuma ação.

Para isso, P5.2 estende `NavItem.access` com um campo `when` (`'project-staff'` | `'project-proponent'`), resolvido em `use-nav.ts` por duas funções puras sobre a sessão (`isProjectStaff`, `isProjectProponent`, em `features/projects/access.ts`); as páginas conferem o mesmo critério no servidor com um equivalente em `lib/auth/access`. Os atalhos de teclado e a busca Cmd+K herdam o filtro de `use-nav.ts`.

---

## 6. Erros e conflitos

Toda chamada passa por `errorMessage` (`@/lib/errors`); para os códigos do fluxo, mensagens próprias:

| Código | Mensagem na tela | Ação |
|---|---|---|
| `file_not_opened` | "Abra todos os arquivos deste documento antes de aprovar." + lista dos que faltam | nenhuma |
| `review_incomplete` | "Ainda há N itens pendentes e M impedimentos." | recarrega |
| `framing_pending` | "Há um enquadramento aguardando a decisão do departamento." | recarrega |
| `framing_already_decided`, `framing_round_stale` | "Outra pessoa já decidiu, ou a indicação mudou. A página foi atualizada." | **recarrega** e fecha o diálogo |
| `department_not_finalistic`, `department_without_reviewer` | texto da API | mantém o diálogo |
| `analysis_not_started` | "Inicie a análise antes de conferir." | recarrega |
| 403 em ação já exibida | "Você não tem permissão para esta ação." | recarrega |

Botões ficam desabilitados com `loading` durante a chamada (`LoadingButton`), para não enviar duas vezes.

---

## 7. Passos de trabalho (um commit por passo; `typecheck`, lint e formato sem problema novo antes de cada um)

1. **P5.1 Contrato.** Tipos, rótulos, ações e funções de serviço (seção 3). Rótulo de `awaiting_opinion` nas listas existentes. *Verificável só por `typecheck`.*
2. **P5.2 Navegação.** `NavItem.access.when` com os critérios "é equipe de projetos" e "é proponente" (seção 5); "Meus projetos" só para quem cadastra; item novo "Análise de projetos"; equivalente no servidor em `lib/auth/access`; documentação em `docs/nav-rbac.md`.
3. **P5.3 Fila.** Página `/dashboard/analise` com abas por perfil, nuqs, busca, filtro de departamento e contadores.
4. **P5.4 Detalhe da equipe, esqueleto.** O assistente reconhece a equipe e renderiza o cabeçalho, as abas e o histórico (somente leitura); aviso do Proponente em `awaiting_opinion`.
5. **P5.5 Conferência.** `ReviewPanel`: progresso, grupos, abrir arquivo, aprovar e desfazer, impedimentos.
6. **P5.6 Ações do Analista.** Iniciar análise, devolver, arquivar e indicar departamento (diálogos e mensagens da seção 6).
7. **P5.7 Enquadramento.** `FramingPanel` e as ações de confirmar e recusar do Revisor.
8. **P5.8 Negação.** Diálogo do Secretário.
9. **P5.9 Administração.** Colunas e selos em departamentos; avisos nos diálogos de usuários.
10. **P5.10 Documentação.** `frontend/CLAUDE.md`/`AGENTS.md` (nova página), `docs/nav-rbac.md`, `backend/README.md` quando preciso, e o roteiro de aceite (seção 8) em `docs/`.

P5.1 e P5.2 vêm primeiro e podem entrar sozinhas. P5.5 a P5.8 dependem do esqueleto (P5.4) e são independentes entre si.

---

## 8. Como validar

**Automático:** `typecheck` sem erros; `lint` e `format:check` sem problema novo em relação à linha de base (`docs/relatorio-fase-0.md`: 1 erro e 12 avisos de lint, 5 arquivos fora do formato, todos anteriores). Rodam no contêiner `web` (`docker compose -f docker-compose.dev.yml exec web bun run ...`).

**Manual, por perfil.** Importante: os administradores de exemplo do seed têm **CPFs fictícios e não entram pelo MT Login**. Para ver a interface de verdade, é preciso um login real:
- **Super Admin** (CPF em `SEED_SUPER_ADMIN_CPF`) enxerga e executa tudo, o que serve para percorrer todas as telas, mas **não** prova as restrições de cada perfil;
- Para provar cada perfil, usar CPFs reais de pessoas que possam entrar: cadastrar em *Administração > Usuários* e atribuir `Analista`, `Revisor do Departamento` (com departamento), `Secretário de Estado` e `Supervisor Administrador`, um por pessoa ou trocando os perfis de quem testa entre os passos.

**Roteiro de aceite** (o mesmo da Fase 6, executável por quem testa):
1. Proponente envia um projeto → aparece na aba *Aguardando análise* do Analista.
2. Analista inicia a análise → abre cada documento, tenta aprovar sem abrir (vê o bloqueio), abre e aprova; devolve um item; o Proponente corrige e reenvia; o Analista reaprova só o devolvido.
3. Com tudo aprovado, o botão **Indicar departamento** habilita; escolher um departamento sem Revisor (desabilitado) e depois um válido.
4. Revisor do departamento certo vê o projeto em *Pendentes*; o de outro departamento não vê.
5. Revisor recusa com justificativa → projeto volta ao Analista, aparece o histórico, o Supervisor recebe o e-mail (no log da API em desenvolvimento).
6. Analista indica o outro departamento → o Revisor dele confirma → o projeto vai para *Aguardando parecer* e o Proponente vê o rótulo.
7. Secretário nega um projeto em cada etapa → todos recebem o aviso de arquivamento.
8. Dois navegadores: dois Revisores do mesmo departamento decidindo juntos → um vence, o outro vê o aviso e a página recarrega.

---

## 9. Riscos e cuidados

| Risco | Cuidado |
|---|---|
| Não há teste automatizado de interface; um erro de tela só aparece na validação manual | Passos pequenos, um commit cada; tipos estritos; roteiro de aceite; você me apontar o que vir diferente |
| Reaproveitar as etapas do proponente para a equipe pode expor um botão de edição | Elas já recebem `readOnly` a partir de `accessOf` (equipe → `base = false`); P5.4 confere cada etapa e o `DocumentsStep` |
| A tela mostrar uma ação que a API recusa | A tela só esconde por permissão/estado; o servidor decide e a mensagem do erro explica |
| `useCan` não avalia condições (departamento, status); o critério de menu "é equipe" lê os **nomes** das condições | Aceito: serve só para esconder menu; o 403 do servidor é tratado (seção 6). Se o nome de uma condição mudar no backend, o teste de perfis da Fase 1 e este critério precisam mudar juntos (anotado no código) |
| Quem acumula Proponente e Analista | Dois itens de menu; no detalhe, `project.access` define a visão (titular vê a do proponente) |
| Dados internos vazarem para o Proponente | A API não os envia; os tipos são opcionais e o painel só renderiza quando existem |
| Lista longa de arquivos/itens | Agrupar por tipo; a API já devolve ordenado; rolagem na própria área |
| Hot reload do `web` em desenvolvimento reiniciar durante o teste | Normal; recarregar a página |

---

## 10. Perguntas

1. **F1:** a fila da equipe fica em página nova (`/dashboard/analise`, "Análise de projetos") e "Meus projetos" passa a aparecer só para quem pode cadastrar projeto? Recomendo que sim.
2. **F8:** aceita validar a interface por roteiro manual, sem criar suíte de testes de interface agora? Se preferir ter testes automáticos de tela, é uma decisão à parte (envolve adicionar ferramentas ao projeto).
3. Para o roteiro manual: você consegue atribuir os perfis a CPFs reais de quem vai testar (inclusive o seu, trocando entre os passos)? Os usuários de exemplo do seed não conseguem entrar.

---

## 11. Definição de pronto

- [ ] Todas as telas da seção 4 funcionando para os perfis da tabela inicial.
- [ ] `typecheck` sem erros; lint e formato sem problema novo.
- [ ] Menu e páginas conferem a permissão (cliente e servidor); "Meus projetos" só para quem cadastra.
- [ ] Todos os erros do fluxo mostram mensagem em português (seção 6).
- [ ] Roteiro de aceite percorrido por você.
- [ ] Documentação atualizada.
