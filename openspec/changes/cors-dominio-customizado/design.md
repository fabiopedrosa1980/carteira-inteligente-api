## Context

A API Go (Gin) usa `github.com/gin-contrib/cors` configurado em `pkg/middleware/cors.go`, registrado globalmente em `internal/adapters/http/router/router.go` via `r.Use(middleware.CORS())`. A lista `AllowOrigins` é estática (hardcoded) e hoje contém os hosts de localhost e `https://carteira-inteligente-eight.vercel.app`, com `AllowCredentials: true`.

## Goals / Non-Goals

**Goals:**
- Autorizar `https://carteira-inteligente.com` e `https://www.carteira-inteligente.com`.
- Mudança mínima, sem regressão das origens existentes.

**Non-Goals:**
- Refatorar a config de CORS para variável de ambiente (poderia ser feito depois).
- Usar origem curinga ou regex de origens.
- Alterar métodos, headers ou credenciais permitidos.

## Decisions

### Decisão: Adicionar as duas origens à lista estática `AllowOrigins`
- **Escolha**: Inserir os dois novos hosts diretamente no slice `AllowOrigins`.
- **Alternativa considerada**: Mover a lista para variável de ambiente (`CORS_ALLOWED_ORIGINS`) lida no startup. Rejeitada por escopo — aumenta a mudança sem necessidade imediata; pode virar um change futuro.
- **Rationale**: É a alteração mais simples, segura e alinhada ao padrão atual do arquivo.

### Decisão: Não usar curinga
- **Escolha**: Listar origens explícitas.
- **Rationale**: `AllowCredentials: true` é incompatível com `Access-Control-Allow-Origin: *`; além disso, origens explícitas são mais seguras.

## Risks / Trade-offs

- **Lista hardcoded cresce e exige novo deploy a cada origem** → Aceitável agora; considerar env var no futuro.
- **Esquecer o `www` ou o apex** → A spec cobre ambos com cenários distintos; a verificação por `curl -H "Origin: ..."` confirma os dois.
- **Mudança só vale após redeploy na Render** → Incluir passo de deploy/validação nas tasks.

## Migration Plan

1. Editar `pkg/middleware/cors.go` adicionando as duas origens.
2. Build/test local (`go build ./...`, `go test ./...`).
3. Commit e push; deploy na Render.
4. Validar com requisição preflight a partir das novas origens.

**Rollback**: Reverter o commit (remover as duas strings) e redeployar.
