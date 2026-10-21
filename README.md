# Convênios e Fomentos

Sistema de controle de Convênios e Fomento.

## Backend

API em Go com controle de permissões granular (RBAC) inspirado no Strapi 5: administradores, usuários finais e tokens de API.

- Código e documentação completa: [`backend/`](backend/README.md)
- Revisão de segurança e pendências: [`backend/SEGURANCA.md`](backend/SEGURANCA.md)
- Início rápido:

```bash
cd backend
cp .env.example .env
make docker-up                    # PostgreSQL + migrações + API em http://localhost:8080
docker compose run --rm --entrypoint /cli api seed
# documentação interativa: http://localhost:8080/docs
```

## Frontend

Painel administrativo em Next.js + shadcn/ui, baseado no template
[Kiranism/next-shadcn-dashboard-starter](https://github.com/Kiranism/next-shadcn-dashboard-starter)
(commit `9084ffd`, licença MIT). O código está em [`frontend/`](frontend/) e o repositório
original serve como referência de exemplo para páginas, componentes e padrões.

```bash
cd frontend
bun install   # ou npm/pnpm
cp env.example.txt .env.local
bun run dev
```
