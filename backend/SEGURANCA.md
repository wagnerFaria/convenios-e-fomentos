# Revisão de segurança e checklist de pendências

Revisão feita ao final da Fase 8, sobre o código e os testes existentes. Complementa o [README](README.md).
Convenção: ✅ verificado (há teste ou verificação automatizada) · 🔧 corrigido nesta revisão · ⚠️ risco aceito/limitação · ⏳ pendente.

## 1. Escopo e método

- Leitura dirigida do código por área (autenticação, autorização, entrada/saída, banco, segredos, logs, infraestrutura).
- `golangci-lint` com `gosec`, `errorlint` e `noctx` (sem avisos) e `govulncheck` nas dependências.
- Todos os achados abaixo que puderam virar teste viraram teste de regressão.
- Verificação de ponta a ponta com a imagem Docker real (migração → API → `/api/ready`, `/api/docs`, `seed`, `cleanup`).
- **Não** foi feito: teste de intrusão externo, fuzzing, análise de carga nem revisão por terceiros.

## 2. Resultado por área

### Autenticação
**Administradores (MT Login / OIDC):**
- ✅ Sem senha local: a API só valida o access token do MT Login (RS256 pelas chaves do JWKS, `iss` e `exp` obrigatórios, tipo `Bearer`,
  client conferido por `azp`/`aud`). Algoritmo `none`/HS256, tokens de ID ou refresh, de outro emissor, de outro client, expirados ou
  assinados por outra chave são rejeitados (testes em `auth/oidc_test.go` e `app/admin_auth_test.go`).
- ✅ Identidade = CPF validado (dígitos verificadores) vindo do token assinado; não há vínculo por e-mail, então não há tomada de conta por
  e-mail não verificado.
- ✅ Primeiro acesso cria o administrador **sem nenhuma role** (nada é liberado automaticamente), de forma atômica (`ON CONFLICT`) e auditada.
- ✅ Bloqueio e desativação valem na hora: cada requisição confere a conta no banco (sem cache).
- ✅ CPF é dado pessoal: mascarado em listagens e na auditoria, completo apenas para o próprio usuário em `/api/admin/auth/me`; não vai para logs.
- ⚠️ **Qualquer pessoa que autentique no client do MT Login ganha uma conta de administrador sem permissões.** Se o realm for aberto a
  cidadãos, a tabela `admins` acumula contas sem perfil; restrinja o client no Keycloak (role/grupo) ou exija um claim para criar o registro.
- ⚠️ Encerrar a sessão no MT Login não derruba, na hora, o access token já emitido (até a expiração dele, curta). Desativar a conta aqui
  tem efeito imediato.
- ⚠️ MFA, bloqueio por tentativas e política de senha são do MT Login (Keycloak), não desta aplicação.

**Usuários finais e tokens de API:**
- ✅ Senhas em Argon2id (parâmetros gravados no hash, comparação em tempo constante); política de 8–128 caracteres.
- ✅ Login com mensagem genérica; a senha é verificada mesmo quando a conta não existe (tempo equalizado).
- ✅ JWT HS256 com `iss`, `aud=api` e `exp` obrigatórios; algoritmo `none` e tokens de outra audiência rejeitados.
- ✅ Refresh token opaco (32 bytes, só o SHA-256 no banco), rotação atômica, reutilização revoga a família, expiração por ociosidade e máxima.
- ✅ Bloqueio, desativação e revogação valem na hora: cada requisição confere a sessão no banco (sem cache).
- ✅ Tokens de confirmação e redefinição: uso único consumido atomicamente, com validade; token de outro tipo/sujeito é recusado.
- ✅ Tokens de API: 256 bits, só o HMAC-SHA256 no banco, exibidos uma vez, `Cache-Control: no-store`, revogação/expiração imediatas.
- 🔧 Parâmetros do Argon2 passaram a ser validados (e, em produção, não podem ficar abaixo do mínimo da OWASP).
- ⚠️ Usuários finais: não há MFA/2FA nem bloqueio de conta por tentativas, só rate limit por IP + identificador
  (`RATE_LIMIT_AUTH_PER_MINUTE`). Ataques distribuídos (credential stuffing) exigem proteção adicional no proxy/WAF.
- ⚠️ JWT simétrico (HS256) dos usuários finais, sem `kid`: trocar o segredo invalida todos os access tokens (até 15 min).

### Autorização
- ✅ Ação não registrada sempre nega; Super Admin também.
- ✅ Filtros de condição são tipados e viram SQL somente parametrizado e apenas com colunas do `Schema`; a injeção em `sort`, `q`, `status`
  e nos filtros de auditoria foi testada.
- ✅ Campos: leitura remove o não permitido (chaves fora do schema nunca saem); escrita recusa (422) campos não permitidos e **todos** os
  campos de sistema, inclusive para o Super Admin.
- ✅ Registro único (leitura/edição/exclusão/publicação) é conferido contra o filtro da permissão (403 fora dele).
- ✅ Mudar `status` exige a ação `publish` mesmo com o campo liberado.
- ✅ Só um Super Admin concede/retira a role Super Admin ou convida com ela; ninguém desativa/exclui a própria conta; o último Super Admin
  ativo é protegido por lock consultivo (testado com requisições concorrentes).
- ✅ Cada rota de `/admin` tem teste de 401 (sem token) e 403 (sem permissão).
- 🔧 Um teste unitário revelou (Fase 4) que `fields: []` ("nenhum campo") virava `null` ("todos os campos") ao normalizar — corrigido, com regressão.
- ⚠️ Quem tem `settings.roles.update` pode editar a própria role e se conceder mais permissões (exceto a role Super Admin, imutável).
  É o comportamento esperado de um RBAC administrável, como no Strapi: conceda essa permissão com critério.
- ⚠️ Public e Authenticated não herdam permissões uma da outra (como no Strapi); revisar ao configurar a API pública.

### Entrada e saída
- ✅ Corpo limitado (`MAX_BODY_BYTES`, 413); JSON rejeita campos desconhecidos; validação com mensagens em português.
- ✅ A API pública usa uma struct de visão (`PublicDocument`): `observacao_interna` e demais campos internos não existem na resposta.
- ✅ Erros 500 nunca vazam detalhes (só o `requestId`); pânicos são capturados.
- ✅ Templates de e-mail usam `html/template` (escape automático; `javascript:` neutralizado) e são validados ao salvar.
- ✅ Cabeçalhos de e-mail recusam quebra de linha (sem injeção de cabeçalho).
- ✅ Redirecionamento da confirmação de e-mail só para URL `http(s)` absoluta, configurável apenas por administrador com permissão.
- 🔧 `X-Request-ID` do cliente agora é validado (`[A-Za-z0-9._-]{1,64}`); antes qualquer texto de até 64 bytes era refletido em cabeçalho e log.
- 🔧 Todas as respostas levam `Cache-Control: no-store` (dados dinâmicos/pessoais não devem ir para caches compartilhados).
- ✅ CSRF não se aplica: a autenticação é por cabeçalho `Authorization`, sem cookies. CORS restrito às origens configuradas.

### Disponibilidade
- 🔧 Limite geral por IP em todas as rotas (`RATE_LIMIT_GLOBAL_PER_MINUTE`, padrão 600), além do limite específico de autenticação.
- ✅ Timeouts do servidor HTTP, corpo limitado, `Argon2` com memória limitada e tamanho máximo de senha (anti-DoS).
- ✅ IP real só via `X-Forwarded-For` quando a conexão vem de um proxy em `TRUSTED_PROXIES`.
- ⚠️ O rate limit é em memória, por instância. Com várias instâncias o limite efetivo é multiplicado; use também o proxy.
- 🔧 Sessões e tokens de uso único só cresciam: novo comando `cleanup` (`make cleanup`) remove os vencidos há mais de N dias.

### Banco de dados
- ✅ SQL via sqlc; SQL montado à mão em apenas três listagens, com valores em `$n` e colunas/ordenação de listas permitidas.
- ✅ Chaves estrangeiras com `ON DELETE` coerente, `CHECK`s de domínio (status, tipos, arrays JSON) e índices únicos parciais.
- ✅ Operações compostas em transação (substituição de permissões, exclusão de role movendo usuários, convite, lote de exclusões).
- ⚠️ Sem criptografia em repouso, backups ou TLS para o PostgreSQL por conta da aplicação: configure `sslmode=verify-full` na `DATABASE_URL` em produção.

### Segredos e logs
- ✅ Em produção a aplicação não sobe com segredos ausentes, curtos, com valor de desenvolvimento ou com os dois segredos JWT iguais.
- ✅ O log de acesso registra só método, caminho (sem query string), status e duração; senhas, tokens e hashes nunca são registrados.
- ✅ Auditoria sem senhas, tokens nem hashes (testado para tokens de API e usuários finais).
- ⚠️ Em desenvolvimento, o `LogMailer` registra o corpo dos e-mails (com tokens) e o `seed` imprime senhas de exemplo. Isso só ocorre
  fora de produção; em produção sem SMTP os e-mails são descartados sem o corpo.

### Dependências
- ✅ `govulncheck`: nenhuma vulnerabilidade alcançável. Atualizados `golang.org/x/crypto`, `moby/go-archive` e `moby/sys/user`
  (esta última pelo `testcontainers`, só em testes).
- ⚠️ Restou o aviso GO-2026-5932 (`x/crypto/openpgp`, sem correção): a aplicação **não importa** esse pacote.
- Reexecute `govulncheck ./...` e `go list -m -u all` periodicamente (ex.: no CI).

### Arquivos dos documentos (MinIO/S3)
- ✅ Bucket privado; nenhum acesso anônimo. Envio e download só por URLs pré-assinadas de curta duração (envio 15 min, download 2 min), emitidas depois da
  checagem de permissão sobre o **documento** (condições por registro valem para os arquivos).
- ✅ O envio usa POST policy: o armazenamento recusa tipo diferente do declarado, tamanho acima do limite (PDF/DOCX 200 MB; imagens 25 MB), troca da
  chave do objeto e uso fora do prazo. A chave do objeto é gerada pelo servidor (`documents/<documento>/<arquivo>`), nunca derivada do nome do cliente, e não sai da API.
- ✅ A confirmação confere existência, tipo, tamanho real e a assinatura do formato (magic bytes); o que não passa é apagado. Downloads saem como
  `attachment`, com o tipo gravado no servidor, e a API pública só serve arquivos confirmados de documentos publicados (senão 404).
- ✅ Excluir o documento ou o arquivo apaga os objetos; envios nunca confirmados são removidos pelo `cleanup` após 24 h.
- ⚠️ Não há antivírus nem verificação do conteúdo além da assinatura do formato. Se os arquivos vierem de terceiros, passe-os por um scanner (ex.: ClamAV)
  antes de liberar o download.
- ⚠️ O console do MinIO só é exposto (`/minio-console`) no HAProxy de desenvolvimento; não o publique em produção e troque `MINIO_ROOT_USER/PASSWORD`.

### Infraestrutura
- ✅ Imagem multi-stage, binário estático em `distroless/static:nonroot` (sem shell, sem gerenciador de pacotes).
- ⚠️ O `/api/docs` carrega a UI (Scalar) de um CDN, sem SRI, e fica público; a CSP da página libera só esse CDN. Em produção, bloqueie
  `/api/docs` no proxy se a documentação não deve ser pública, ou sirva a UI localmente.
- ⚠️ `Strict-Transport-Security` é enviado sempre, mas só tem efeito em HTTPS: termine o TLS no proxy.

## 3. Riscos aceitos e limitações conhecidas

| Tema | Descrição | Mitigação / quando tratar |
|---|---|---|
| Enumeração de contas | O cadastro de usuário final responde 409 `already_taken` para e-mail/usuário existentes; em "esqueci a senha" há diferença de latência (grava token e envia e-mail só quando a conta existe). As respostas HTTP são idênticas. | Rate limit. Para eliminar o canal de tempo, enfileirar o envio de e-mails. |
| E-mail síncrono | O envio SMTP acontece dentro da requisição; falha de envio é registrada no log, mas não chega ao cliente. | Fila/worker de e-mail quando o volume pedir. |
| Confirmação de e-mail ligada depois | Usuários já logados sem confirmar só são barrados no próximo refresh/login. | Aceitável; revogar sessões ao ligar a opção se necessário. |
| Auditoria | Sem retenção/rotação, sem assinatura de integridade; conteúdo (documentos) não é auditado. | Política de retenção e arquivamento conforme a necessidade. |
| Rate limit por instância | Ver "Disponibilidade". | Redis ou proxy quando houver várias instâncias. |
| Sem MFA (usuário final) | Ver "Autenticação". | Planejar TOTP/WebAuthn; para administradores, é do MT Login. |

## 4. Checklist de pendências

**Funcionalidades fora do escopo definido** (decisões já documentadas no README):
- ⏳ OAuth/OIDC para **usuários finais**: estrutura pronta (`end_users.provider`, `enduser.Provider`); nenhum provedor implementado (os administradores já usam o MT Login).
- ⏳ i18n/locales: preparado (`SupportsLocales`, `properties`), sem conteúdo localizado.
- ⏳ Frontend administrativo.

**Melhorias identificadas (nenhuma bloqueia o uso):**
- ⏳ Remover `admins.password_hash` e a tabela `admin_sessions` (sem uso desde o login pelo MT Login) numa migração, depois de validado o novo fluxo.
- ⏳ Bloquear/desbloquear administradores por endpoint (a coluna `blocked` existe e é respeitada; a API cobre ativar/desativar).
- ⏳ Opção para desabilitar ou servir localmente o `/api/docs`.
- ⏳ Fila de e-mails; retenção da auditoria; MFA para administradores.
- ⏳ Pipeline de CI executando `make lint`, `make test`, `make cover-authz` e `govulncheck`.
- ⏳ Métricas/observabilidade (Prometheus, traces); hoje há logs JSON com `request_id` e `/api/ready`.
- ⏳ `make dev` depende do `air`, que é opcional e não vem instalado.

## 5. Checklist para colocar em produção

1. `APP_ENV=production` e `JWT_API_SECRET`, `API_TOKEN_SALT` aleatórios (`openssl rand -hex 32`), distintos entre si.
   `JWKS_URI`, `OIDC_ISSUER`, `OIDC_CLIENT_ID` e `OIDC_CPF_CLAIM` do MT Login de **produção** (a API não sobe sem eles); no frontend, as `MTLOGIN_*`
   de produção, com a URL de retorno e a de pós-logout cadastradas no client.
2. `DATABASE_URL` com `sslmode=verify-full`; usuário do banco com o mínimo de privilégios; backups agendados.
3. TLS no proxy reverso, `TRUSTED_PROXIES` com o CIDR dele e `CORS_ALLOWED_ORIGINS` só com o(s) domínio(s) do front.
4. `PUBLIC_URL` correto e SMTP configurado (`SMTP_*`); sem SMTP os e-mails são descartados.
5. Criar o primeiro Super Admin pela CLI (`cli admin:create --cpf ...`) ou `SEED_SUPER_ADMIN_CPF`.
6. Rodar `cli migrate:up` antes de subir a nova versão; **não** rodar o `seed` em produção, a menos que se queira criar as roles de exemplo
   (os usuários de exemplo só existem com `APP_ENV=development`).
7. Agendar `cli cleanup` (diário) e acompanhar `GET /ready`.
8. Decidir sobre a exposição de `/api/docs` e revisar as permissões das roles (principalmente `settings.roles.*` e `settings.users.*`).
9. Rodar `govulncheck ./...` e a suíte completa no pipeline antes de cada release.
