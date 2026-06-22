## Why

O frontend (Carteira Inteligente) já tem a tela de Alocação & Rebalanceamento e chama `GET/PUT /api/v1/allocation`, mas **esse endpoint não existe** nesta API. Hoje a configuração (alvos por classe + limite de concentração) **não persiste**: o frontend degrada para defaults a cada recarregamento. A persistência deve ser **por usuário** (como `goals` e `transactions`), usando o `userID` do token Google validado pelo `AuthRequired`.

## What Changes

- **Novo recurso `AllocationConfig`** persistido por usuário: alvos por classe (Ações, FIIs, ETFs) + limite de concentração por ativo.
- **Endpoint REST protegido** sob `/api/v1/allocation` (com `AuthRequired`):
  - `GET /api/v1/allocation` → retorna a config do usuário; se não houver, retorna os **defaults** (Ações 50 / FIIs 40 / ETFs 10; limite 20).
  - `PUT /api/v1/allocation` → faz **upsert** da config do usuário e retorna a config salva.
- **Singleton por usuário**: uma linha por `userID` (chave primária), com upsert no PUT.
- **Migração de schema** (`AutoMigrate`) para a nova tabela.

## Capabilities

### New Capabilities
- `allocation-config`: leitura e gravação (upsert) da configuração de alocação por usuário, com defaults quando ausente, protegida por autenticação.

## Impact

- **Domínio**: `internal/domain/allocation.go` (entidade `AllocationConfig` + `AllocationRepository`).
- **Aplicação**: `internal/application/allocation_service.go` (Get com defaults, Save/upsert).
- **Infra**: `internal/infrastructure/persistence/gorm_allocation_repository.go` + `AutoMigrate` em `database.go`.
- **HTTP**: `internal/adapters/http/dto/allocation_dto.go`, `handler/allocation_handler.go`, grupo `/allocation` em `router.go` (com `AuthRequired`), wiring em `cmd/api/main.go`.
- **Contrato (combina com o frontend)**: `{ "targets": { "Acoes": number, "FIIs": number, "ETFs": number }, "concentrationLimit": number }`.
- **Frontend**: nenhuma mudança necessária — o `authInterceptor` já envia o `Bearer <token>` e `BackendApiService.getAllocation/updateAllocation` já consomem esse contrato; passam a persistir de fato quando o endpoint existir.
