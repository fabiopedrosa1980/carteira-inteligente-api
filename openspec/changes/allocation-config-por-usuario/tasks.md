## 1. Domínio

- [x] 1.1 Criar `internal/domain/allocation.go`: struct `AllocationConfig` (`UserID` PK, `AcoesTarget`, `FIIsTarget`, `ETFsTarget`, `ConcentrationLimit`, `UpdatedAt`) e interface `AllocationRepository` (`Get(userID) (*AllocationConfig, error)`, `Upsert(*AllocationConfig) error`).

## 2. Aplicação

- [x] 2.1 Criar `internal/application/allocation_service.go`: `Get(userID)` retorna a config ou os defaults (Ações 50 / FIIs 40 / ETFs 10 / limite 20) quando ausente; `Save(userID, cfg)` seta `UserID`/`UpdatedAt` e faz upsert.

## 3. Infraestrutura

- [x] 3.1 Criar `internal/infrastructure/persistence/gorm_allocation_repository.go`: `Get` (First por `user_id`, tratando record-not-found → nil) e `Upsert` (`clause.OnConflict` por `user_id`, `UpdateAll`).
- [x] 3.2 Incluir `&domain.AllocationConfig{}` no `AutoMigrate` em `database.go`.

## 4. HTTP (adapters)

- [x] 4.1 Criar `internal/adapters/http/dto/allocation_dto.go`: `AllocationRequest`/`AllocationResponse` com `targets{Acoes,FIIs,ETFs}` + `concentrationLimit`; `FromDomain`/`ToDomain`.
- [x] 4.2 Criar `internal/adapters/http/handler/allocation_handler.go`: `GetAllocation` (userID do contexto → service.Get → DTO) e `PutAllocation` (bind → ToDomain → service.Save → DTO).
- [x] 4.3 Registrar grupo `/allocation` com `middleware.AuthRequired()` (`GET ""`, `PUT ""`) em `router.go`; estender a assinatura de `SetupRouter` com `allocationHandler`.

## 5. Wiring + verificação

- [x] 5.1 Em `cmd/api/main.go`: instanciar `allocationRepo → allocationService → allocationHandler` e passar ao `SetupRouter` (ajustar chamada/qualquer teste que monte o router).
- [x] 5.2 `go build ./...` e `go test ./...`; validar manualmente `GET` (defaults) e `PUT`+`GET` (persistência) com token válido.
- [x] 5.3 Commit e push no repo da API (stage de arquivos específicos).
