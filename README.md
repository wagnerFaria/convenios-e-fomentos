# Convênios e Fomentos

Sistema de controle de **convênios, termos de fomento e prestação de contas**. O projeto tem uma API em Go com
controle de permissões granular (RBAC) e um painel administrativo em Next.js que a consome.

> **Estado atual:** a fundação está pronta (autenticação, administradores, roles e permissões, usuários finais, tokens de
> API, auditoria e um recurso de exemplo, *Documents*). As regras de negócio de convênios e fomentos ainda não foram
> implementadas; a legislação de referência está em [`docs/`](docs/).

## Sumário

1. [Visão geral](#1-visão-geral)
2. [Pré-requisitos](#2-pré-requisitos)
3. [Início rápido (Docker, recomendado)](#3-início-rápido-docker-recomendado)
4. [Primeiro acesso](#4-primeiro-acesso)
5. [Como o ambiente de desenvolvimento funciona](#5-como-o-ambiente-de-desenvolvimento-funciona)
6. [Rotina de trabalho](#6-rotina-de-trabalho)
7. [Rodando sem Docker (no host)](#7-rodando-sem-docker-no-host)
8. [Variáveis de ambiente](#8-variáveis-de-ambiente)
9. [Autenticação e e-mails em desenvolvimento](#9-autenticação-e-e-mails-em-desenvolvimento)
10. [Testes, lint e formatação](#10-testes-lint-e-formatação)
11. [Solução de problemas](#11-solução-de-problemas)
12. [Estrutura do repositório](#12-estrutura-do-repositório)
13. [Documentação adicional](#13-documentação-adicional)

---

## 1. Visão geral

| Parte | Tecnologia | Porta | Pasta |
|---|---|---|---|
| Frontend | Next.js 16, React 19, TypeScript, Tailwind v4, shadcn/ui, TanStack Query/Form, Bun | 3000 | [`frontend/`](frontend/) |
| API | Go 1.27, PostgreSQL 16, sqlc, goose, JWT | 8080 | [`backend/`](backend/) |
| Banco | PostgreSQL 16 | 5432 | volume Docker |

```
Navegador ──► Next.js (:3000) ──► API Go (:8080) ──► PostgreSQL
              │  cookies httpOnly      │  JWT + RBAC
              └─ proxy /api/backend/*  └─ /admin/*  /api/*  /docs
```

O navegador **nunca** fala direto com a API nem enxerga os tokens: o login é feito por *server actions* do Next, que guardam
o access token e o refresh token em cookies `httpOnly`; as chamadas autenticadas passam pelo proxy `/api/backend/*`, que
injeta o `Authorization: Bearer` e renova o par de tokens quando necessário. Detalhes na
[seção 9](#9-autenticação-e-e-mails-em-desenvolvimento).

## 2. Pré-requisitos

Para o fluxo recomendado (Docker) basta:

- **Docker** com **Docker Compose v2** (`docker compose version`).
- Portas livres no host: **3000**, **8080** e **5432** (ou outra, via `DB_PORT`; veja a [seção 11](#11-solução-de-problemas)).

Opcional, só para rodar fora do Docker ou executar os testes do backend:

- **Go 1.27+** (o `go.mod` fixa a versão; o Go baixa o toolchain correto sozinho).
- **Bun 1.x** (gerenciador e runtime do frontend; `npm`/`pnpm` também funcionam, mas o lockfile é do Bun).
- Ferramentas Go: `sqlc`, `goose` e `golangci-lint` (veja [`backend/README.md`](backend/README.md#2-subindo-o-ambiente-do-zero)).

## 3. Início rápido (Docker, recomendado)

Da raiz do repositório:

```bash
cp .env.example .env     # opcional: sem .env valem os padrões de desenvolvimento
docker compose -f docker-compose.dev.yml up --build
```

Toda a configuração do compose vem de variáveis (portas, credenciais do banco, URLs, segredos, SMTP...). Elas ficam no
`.env` da raiz, que o git ignora; o [`.env.example`](.env.example) lista todas, com comentários.

Na primeira vez isso constrói as imagens e instala as dependências (alguns minutos). Quando os logs mostrarem
`servidor iniciado` (API) e `Ready` (Next), acesse:

| O quê | URL |
|---|---|
| Painel (frontend) | <http://localhost:3000> |
| API | <http://localhost:8080> |
| Documentação interativa da API (OpenAPI) | <http://localhost:8080/docs> |
| Liveness / readiness da API | <http://localhost:8080/health> · <http://localhost:8080/ready> |

Dicas:

```bash
# Em segundo plano, e acompanhar os logs de um serviço
docker compose -f docker-compose.dev.yml up -d --build
docker compose -f docker-compose.dev.yml logs -f web api

# Evite repetir -f em todo comando
export COMPOSE_FILE=docker-compose.dev.yml
docker compose up -d

# Parar (mantém o banco) · parar e apagar banco e caches
docker compose down
docker compose down -v
```

## 4. Primeiro acesso

Um banco novo não tem nenhum administrador. Escolha **uma** das opções:

### Opção A — Tela de configuração inicial

Abra <http://localhost:3000/auth/setup>, preencha nome, e-mail e senha (mínimo de 8 caracteres). O primeiro Super Admin é
criado e você já entra no painel. Essa tela só funciona **enquanto não existir nenhum administrador** (depois, o backend
responde `setup_completed` e você é levado ao login). Em produção o registro por HTTP vem desabilitado; use a CLI (opção C).

### Opção B — Seed com usuários de desenvolvimento

```bash
docker compose exec api go run ./cmd/cli seed
```

O seed é idempotente e cria roles e permissões padrão, departamentos de exemplo e, **só em desenvolvimento**, um usuário por
role: `super@`, `editor@`, `author@`, `revisor@` e `usuario@dev.local`. **As senhas aleatórias aparecem no log apenas na
primeira execução**; copie-as na hora (campo `senha`). Se perdê-las, redefina com `admin:reset-password` (opção C) ou
recomece com `docker compose down -v`.

> Depois do seed já existem administradores, então a tela `/auth/setup` deixa de funcionar. Use a opção A **ou** a B como
> primeiro passo; se quiser as duas coisas, faça a A primeiro e rode o seed em seguida.

### Opção C — CLI

```bash
docker compose exec api go run ./cmd/cli admin:create --email voce@exemplo.com --password 'uma-senha-forte'
docker compose exec api go run ./cmd/cli admin:reset-password --email voce@exemplo.com --password 'nova-senha'
```

Os demais administradores são convidados pelo painel; em desenvolvimento a URL do convite (e os e-mails em geral) aparece
nos logs da API (veja a [seção 9](#9-autenticação-e-e-mails-em-desenvolvimento)).

## 5. Como o ambiente de desenvolvimento funciona

O arquivo [`docker-compose.dev.yml`](docker-compose.dev.yml) define quatro serviços:

| Serviço | Imagem / build | O que faz |
|---|---|---|
| `db` | `postgres:16-alpine` | Banco `app` (usuário/senha `app`), dados no volume `pgdata`. Credenciais por `POSTGRES_*`; porta publicada em `DB_PORT` (padrão 5432). |
| `migrate` | [`backend/Dockerfile.dev`](backend/Dockerfile.dev) | Aplica as migrações (`migrate:up`) e termina. A API só sobe depois que ele conclui. |
| `api` | [`backend/Dockerfile.dev`](backend/Dockerfile.dev) | Go com [air](https://github.com/air-verse/air): recompila e reinicia a cada alteração em `.go`, `.sql` ou `.yaml`. |
| `web` | [`frontend/Dockerfile.dev`](frontend/Dockerfile.dev) | `bun install` + `next dev` (Turbopack) com HMR. |

**Live reload.** O código do host é montado nos contêineres (`./backend:/app`, `./frontend:/app`), então basta salvar o
arquivo: a API é recompilada em segundos e o Next atualiza a página sem recarregar. Não há `COPY` do código nas imagens de
desenvolvimento; elas só trazem as ferramentas.

**Volumes nomeados:** `pgdata` (banco), `go-mod` e `go-build` (caches do Go), `web-node-modules` e `web-next` (dependências e
cache do Next, mantidos dentro do Docker para não misturar com o host).

**Permissões de arquivo.** Os contêineres rodam com o UID/GID `1000` por padrão (o de um primeiro usuário Linux); assim
`backend/tmp`, `frontend/.next` e alterações em `go.sum` ficam com o seu usuário. Se o seu for outro (`id -u`, `id -g`), crie um
`.env` na raiz com `UID=1001` e `GID=1001` (o Compose o lê sozinho; o Bash não deixa sobrescrever `UID` na linha de comando).

**Configuração.** Tudo é parametrizável por variáveis (`.env` na raiz ou shell; o shell tem precedência) e tem um padrão de
desenvolvimento no compose, então nada é obrigatório. Os padrões que importam:

- `PUBLIC_URL` e `CORS_ALLOWED_ORIGINS` (`http://localhost:${WEB_EXTERNAL_PORT}`): os links dos e-mails (convite, redefinição de
  senha) abrem o frontend. O Next redireciona `/admin/reset-password` e `/admin/accept-invite` para as telas `/auth/...`.
- `TRUSTED_PROXIES=172.16.0.0/12`: o backend confia no `X-Forwarded-For` enviado pelo Next (rede padrão do Docker), para que
  o rate limit de login considere o IP do usuário e não o do servidor Next.
- `DATABASE_URL`: montada a partir de `POSTGRES_USER`, `POSTGRES_PASSWORD` e `POSTGRES_DB` (host `db`, na rede do Docker).
  Defina-a explicitamente apenas para usar outro banco. **Atenção:** uma `DATABASE_URL` exportada no seu shell (modo sem Docker,
  apontando para `localhost`) sobrescreve a do `.env` e quebra os contêineres.
- `API_URL=http://api:8080` no frontend: o servidor do Next fala com a API pela rede do Docker.
- Segredos (`JWT_*`, `API_TOKEN_SALT`): valores fixos de desenvolvimento, **nunca** use em produção.

## 6. Rotina de trabalho

Todos os comandos abaixo assumem `export COMPOSE_FILE=docker-compose.dev.yml`.

**Backend (Go)**

```bash
docker compose logs -f api                              # logs (JSON) da API
docker compose exec api go run ./cmd/cli migrate:status # situação das migrações
docker compose run --rm migrate                         # aplicar novas migrações
docker compose exec api go run ./cmd/cli migrate:down   # desfazer a última migração
docker compose exec api go run ./cmd/cli cleanup        # limpar tokens e sessões expirados
docker compose exec api go build ./...                  # checar compilação
docker compose restart api                              # reiniciar (raramente necessário)
```

- Criar uma migração: `goose -dir backend/internal/database/migrations create nome_da_migracao sql` (no host) e depois
  `docker compose run --rm migrate`. A API **não** aplica migrações sozinha.
- Alterou `internal/database/queries/*.sql` ou `sqlc.yaml`? Rode `make sqlc` em `backend/` (precisa do `sqlc` no host) e
  commite o código gerado.
- A especificação [`backend/api/openapi.yaml`](backend/api/openapi.yaml) é embutida na API; ao mudar rotas, atualize-a
  (há um teste que garante que ela cobre as rotas registradas).

**Frontend (Next)**

```bash
docker compose logs -f web
docker compose exec web bun add <pacote>        # adicionar dependência (atualiza package.json e bun.lock no host)
docker compose exec web bun run typecheck
docker compose exec web bun run lint
docker compose exec web bun run format
```

- A cada `up`, o `web` roda `bun install --frozen-lockfile`. Se alguém atualizou o `bun.lock`, basta
  `docker compose restart web`.
- Padrões do projeto (formulários, camada de API, ícones, nuqs) em [`frontend/CLAUDE.md`](frontend/CLAUDE.md) e
  [`frontend/AGENTS.md`](frontend/AGENTS.md).

**Banco de dados**

```bash
docker compose exec db psql -U app -d app            # console SQL
# ou, do host, com a porta publicada:
psql "postgres://app:app@localhost:5432/app"
docker compose down -v                                # recomeçar do zero (apaga TODOS os dados locais)
```

## 7. Rodando sem Docker (no host)

Útil para depurar com IDE. Mantenha só o PostgreSQL no Docker:

```bash
# 1. Banco
docker compose -f docker-compose.dev.yml up -d db     # DB_PORT=5433 no .env se a 5432 estiver ocupada

# 2. API (em backend/)
cd backend
cp .env.example .env                                  # ajuste DATABASE_URL se mudou a porta
make migrate-up
make seed                                             # opcional
make dev                                              # air (live reload); ou: make run

# 3. Frontend (em frontend/, outro terminal)
cd frontend
cp env.example.txt .env.local                         # API_URL=http://localhost:8080
bun install
bun run dev
```

Neste modo, aponte `PUBLIC_URL` do backend para `http://localhost:3000` no `backend/.env` para que os links dos e-mails
abram o frontend. Comandos do `Makefile` do backend: `run`, `dev`, `build`, `test`, `test-unit`, `cover-authz`, `lint`,
`sqlc`, `migrate-up`, `migrate-down`, `migrate-create name=...`, `seed`, `cleanup`, `docker-up`, `docker-down`.

> O [`backend/docker-compose.yml`](backend/docker-compose.yml) é uma stack **só do backend**, sem live reload (constrói a
> imagem de produção). Para o dia a dia, prefira o `docker-compose.dev.yml` da raiz.

## 8. Variáveis de ambiente

**Frontend** (`frontend/.env.local`; modelo em [`frontend/env.example.txt`](frontend/env.example.txt)). No Docker, o compose já define o necessário.

| Variável | Padrão | Descrição |
|---|---|---|
| `API_URL` | `http://localhost:8080` | URL da API usada **pelo servidor** do Next (proxy, server actions, `proxy.ts`). |
| `NEXT_PUBLIC_APP_URL` | — | URL pública do frontend (metadata). |
| `NEXT_PUBLIC_SENTRY_DISABLED` | `true` no compose | Desliga o Sentry em desenvolvimento. |
| `WATCHPACK_POLLING` | `false` | `true` força *polling* de arquivos (macOS/Windows, se o HMR não detectar mudanças). |

**Backend** (`backend/.env`; modelo em [`backend/.env.example`](backend/.env.example), lista completa na
[seção 6 do README do backend](backend/README.md#6-variáveis-de-ambiente)). As principais:

| Variável | Descrição |
|---|---|
| `APP_ENV` | `development` \| `test` \| `production`. Em produção os segredos viram obrigatórios. |
| `DATABASE_URL` | String de conexão do PostgreSQL. |
| `JWT_ADMIN_SECRET`, `JWT_API_SECRET`, `API_TOKEN_SALT` | Segredos (≥ 32 caracteres e distintos em produção). Os valores de exemplo são só para desenvolvimento. |
| `ACCESS_TOKEN_TTL` / `REFRESH_TOKEN_*` | Duração dos tokens (15 min / 7 dias de inatividade / 30 dias no máximo). |
| `CORS_ALLOWED_ORIGINS`, `TRUSTED_PROXIES`, `PUBLIC_URL` | CORS, proxies confiáveis (IP real) e URL base dos links de e-mail. |
| `INITIAL_REGISTRATION_ENABLED` | Habilita `register-first` por HTTP; sem valor, só fora de produção. |
| `SMTP_*` | Envio de e-mail; sem `SMTP_HOST` os e-mails são apenas registrados no log. |

**Compose de desenvolvimento** (`.env` na raiz; modelo em [`.env.example`](.env.example)):

| Grupo | Variáveis |
|---|---|
| Portas no host | `WEB_EXTERNAL_PORT` (3000), `API_EXTERNAL_PORT` (8080), `DB_PORT` (5432) |
| Usuário do host | `UID`, `GID` (1000) |
| PostgreSQL | `POSTGRES_IMAGE`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `DATABASE_URL` (opcional) |
| Backend | `APP_ENV`, `LOG_LEVEL`, `PUBLIC_URL`, `CORS_ALLOWED_ORIGINS`, `TRUSTED_PROXIES`, `JWT_ADMIN_SECRET`, `JWT_API_SECRET`, `API_TOKEN_SALT`, `ACCESS_TOKEN_TTL`, `REFRESH_TOKEN_MAX_TTL`, `REFRESH_TOKEN_IDLE_TTL`, `RATE_LIMIT_AUTH_PER_MINUTE`, `RATE_LIMIT_GLOBAL_PER_MINUTE`, `SMTP_*` |
| Frontend | `API_URL`, `NEXT_PUBLIC_APP_URL`, `NEXT_PUBLIC_SENTRY_DISABLED`, `NEXT_TELEMETRY_DISABLED`, `WATCHPACK_POLLING` |

## 9. Autenticação e e-mails em desenvolvimento

**Sessão.** Login, logout, setup, "esqueci a senha", redefinição e aceite de convite são *server actions*
([`frontend/src/features/auth/actions.ts`](frontend/src/features/auth/actions.ts)). O access token (15 min) e o refresh token
ficam em cookies `httpOnly` + `SameSite=Lax` (`Secure` em produção):

- [`frontend/src/proxy.ts`](frontend/src/proxy.ts) protege `/dashboard/*`: sem cookie redireciona ao login; com o access token
  expirado, renova o par de tokens antes de renderizar.
- [`frontend/src/app/api/backend/[...path]/route.ts`](frontend/src/app/api/backend/[...path]/route.ts) é o proxy autenticado
  usado pelo navegador (`apiClient` em [`src/lib/api-client.ts`](frontend/src/lib/api-client.ts)). Só repassa `/admin/*`,
  bloqueia as rotas que devolvem tokens e renova automaticamente em caso de 401.
- A renovação é **serializada por refresh token**: o backend rotaciona o token a cada uso e revoga a sessão inteira se um token
  já usado for reapresentado, então requisições simultâneas compartilham a mesma renovação
  ([`src/lib/auth/refresh.ts`](frontend/src/lib/auth/refresh.ts)). Isso vale para um único processo do Next; várias instâncias
  exigiriam um lock distribuído.
- Permissões na interface: `useCan()` e `<Can>` leem as permissões do `/admin/auth/me`. Servem só para esconder elementos; quem
  autoriza de fato é sempre o backend.

**E-mails.** Sem `SMTP_HOST`, a API não envia nada: registra o e-mail no log, com o link completo. Para testar a redefinição de
senha ou o convite:

```bash
docker compose logs api | grep "e-mail (modo desenvolvimento)"
```

Copie a URL (`http://localhost:3000/admin/reset-password?token=...`) e abra no navegador. Os tokens são de uso único.

## 10. Testes, lint e formatação

**Backend** (rodam no host; os testes de integração sobem um PostgreSQL próprio via *testcontainers*, então exigem Docker
acessível e Go instalado):

```bash
cd backend
make test          # tudo, com -race
make test-unit     # só testes unitários (-short, sem Docker)
make cover-authz   # exige >= 80% de cobertura no pacote authz
make lint          # golangci-lint
```

**Frontend** (no host ou via `docker compose exec web ...`):

```bash
cd frontend
bun run typecheck
bun run lint            # oxlint (avisos não quebram; erros sim)
bun run format:check    # oxfmt; use `bun run format` para corrigir
bun run build           # build de produção (valida rotas e tipos)
```

O frontend ainda não tem suíte de testes automatizados; a verificação de ponta a ponta é feita manualmente no navegador.

## 11. Solução de problemas

| Sintoma | Causa / solução |
|---|---|
| `port is already allocated` / `address already in use` na **5432** | Há um PostgreSQL local. Defina `DB_PORT=5433` no `.env` (a API e o `web` falam com o `db` pela rede interna, então nada mais muda). |
| `3000` ou `8080` ocupadas | Pare o processo que as usa (`ss -ltnp \| grep -E ':(3000\|8080)'`). Ou mude `WEB_EXTERNAL_PORT` / `API_EXTERNAL_PORT` no `.env` (`PUBLIC_URL` e CORS acompanham a porta do frontend). |
| Alterações em arquivos não recarregam (macOS/Windows) | `WATCHPACK_POLLING=true` para o frontend; para a API, defina `poll = true` em `[build]` do `backend/.air.toml`. |
| `permission denied` em `tmp/`, `.next/` ou `go.sum` | Arquivos criados como root por uma execução anterior. `sudo chown -R $USER backend/tmp frontend/.next` e confira `UID`/`GID`. |
| Tela de setup devolve "registro inicial já concluído" | Já existe administrador (talvez criado pelo seed). Entre pelo login, ou recomece com `docker compose down -v`. |
| Esqueci a senha dos usuários do seed | `docker compose exec api go run ./cmd/cli admin:reset-password --email super@dev.local --password '...'`. |
| API sobe e cai com erro de conexão ao banco | O banco ainda iniciava; `docker compose logs db api`. O `migrate` só libera a API com o `db` saudável. |
| `429 rate_limited` no login | Muitas tentativas (limite por IP+e-mail). Aguarde um minuto. |
| Erros de tipos estranhos referindo `.next/types` | Cache gerado antigo: `docker compose rm -sf web && docker volume rm convenios-dev_web-next`, ou `rm -rf frontend/.next` no modo host. |
| Mudou dependências e o `web` não enxerga | `docker compose restart web` (roda `bun install` de novo). Em último caso, `docker volume rm convenios-dev_web-node-modules`. |
| Logout inesperado ao usar várias instâncias do Next | A renovação de tokens é serializada por processo; veja a [seção 9](#9-autenticação-e-e-mails-em-desenvolvimento). |

## 12. Estrutura do repositório

```
.
├── docker-compose.dev.yml      # ambiente completo de desenvolvimento (live reload)
├── .dockerignore
├── backend/                    # API em Go
│   ├── cmd/api, cmd/cli        # servidor HTTP e CLI (migrate, seed, admin:create, cleanup...)
│   ├── internal/
│   │   ├── app/                # montagem do grafo de dependências + testes de integração
│   │   ├── auth/ admin/        # autenticação e gestão de administradores
│   │   ├── authz/              # motor de permissões (ações, campos, condições)
│   │   ├── enduser/ apitoken/  # usuários finais e tokens de API
│   │   ├── document/           # recurso de exemplo
│   │   ├── audit/ httpapi/ ... # auditoria, roteador, middlewares, config, mailer
│   │   └── database/           # migrações (goose) e queries (sqlc)
│   ├── api/openapi.yaml        # contrato da API (servido em /docs)
│   ├── Dockerfile, Dockerfile.dev, docker-compose.yml, .air.toml, Makefile
│   └── README.md, SEGURANCA.md
├── frontend/                   # painel Next.js
│   ├── src/app/                # rotas: /auth/*, /dashboard/*, /api/backend/[...path]
│   ├── src/features/auth/      # ações, formulários, sessão, permissões no cliente
│   ├── src/lib/auth, src/lib/api, src/lib/api-client.ts
│   ├── src/proxy.ts            # guarda de rotas e renovação de tokens
│   ├── src/components/         # UI (shadcn/ui), layout, formulários, data-table
│   └── Dockerfile.dev, Dockerfile, Dockerfile.bun, env.example.txt
└── docs/                       # legislação e material de referência (convênios e termos de fomento)
```

## 13. Documentação adicional

- API e controle de acesso: [`backend/README.md`](backend/README.md) (arquitetura, receitas para nova ação/condição/policy/recurso,
  decisões de projeto) e [`backend/SEGURANCA.md`](backend/SEGURANCA.md) (revisão de segurança e pendências).
- Contrato da API: <http://localhost:8080/docs> com o ambiente no ar, ou [`backend/api/openapi.yaml`](backend/api/openapi.yaml).
- Frontend: [`frontend/CLAUDE.md`](frontend/CLAUDE.md), [`frontend/AGENTS.md`](frontend/AGENTS.md) e [`frontend/docs/`](frontend/docs/)
  (formulários e temas). Os Dockerfiles de **produção** do frontend e `frontend/docs/deployment.md` ainda mencionam o Clerk e
  precisam ser revisados antes do deploy.
- Legislação e referências do domínio: [`docs/`](docs/).
- O frontend parte do template [Kiranism/next-shadcn-dashboard-starter](https://github.com/Kiranism/next-shadcn-dashboard-starter)
  (licença MIT).
