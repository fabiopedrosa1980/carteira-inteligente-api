## Context

Arquitetura hexagonal (domain → application → adapters/infrastructure). `goals` e `transactions` já são per-user: o `AuthRequired` valida o token Google e seta `c.Set("userID", sub)`; os handlers leem `c.GetString("userID")`; repositórios filtram por `user_id`. Esta feature replica esse padrão para uma config **singleton por usuário**.

## Goals / Non-Goals

**Goals:** persistir alvos por classe + limite de concentração por usuário; GET com defaults; PUT upsert; protegido por auth; contrato igual ao `ApiAllocation` do frontend.

**Non-Goals:** versionar histórico de config; validar que os alvos somam 100 (o frontend já limita); alterar frontend (já compatível).

## Decisions

- **Entidade singleton por usuário** (`internal/domain/allocation.go`):
  ```go
  type AllocationConfig struct {
      UserID             string  `gorm:"primaryKey" json:"-"`
      AcoesTarget        float64 `json:"-"`
      FIIsTarget         float64 `json:"-"`
      ETFsTarget         float64 `json:"-"`
      ConcentrationLimit float64 `json:"-"`
      UpdatedAt          time.Time `json:"-"`
  }
  type AllocationRepository interface {
      Get(userID string) (*AllocationConfig, error) // nil quando não existe
      Upsert(cfg *AllocationConfig) error
  }
  ```
  `UserID` como PK garante 1 linha/usuário; o JSON do domínio é irrelevante (a serialização vai pelo DTO).
- **Defaults no service** (`allocation_service.go`): `Get(userID)` busca no repo; se `nil`, devolve `AllocationConfig{Acoes:50, FIIs:40, ETFs:10, ConcentrationLimit:20}` (sem persistir). `Save(userID, cfg)` seta `UserID`/`UpdatedAt` e chama `Upsert`.
- **Repo GORM** (`gorm_allocation_repository.go`): `Get` = `First(&cfg, "user_id = ?", userID)` tratando `ErrRecordNotFound` → `(nil, nil)`. `Upsert` = `clause.OnConflict{Columns: user_id, UpdateAll: true}` no `Create`, ou `Save` (PK presente).
- **DTO** (`allocation_dto.go`): `AllocationRequest`/`AllocationResponse` com `Targets struct{ Acoes, FIIs, ETFs float64 }` e `ConcentrationLimit float64`, tags JSON `targets`/`concentrationLimit`. Funções `FromDomain`/`ToDomain`.
- **Handler** (`allocation_handler.go`): `GetAllocation` lê `userID`, chama `service.Get`, responde DTO. `PutAllocation` faz `ShouldBindJSON`, mapeia para domínio, `service.Save`, responde DTO salvo.
- **Router**: grupo `/allocation` com `middleware.AuthRequired()`, `GET ""` e `PUT ""`. `SetupRouter` ganha o parâmetro `allocationHandler`.
- **Migração**: incluir `&domain.AllocationConfig{}` no `AutoMigrate` em `database.go`.
- **Wiring** (`main.go`): `allocationRepo → allocationService → allocationHandler`, passado ao `SetupRouter`.

## Risks / Trade-offs

- [`SetupRouter` muda de assinatura] → atualizar a chamada em `main.go` (e quaisquer testes que montem o router).
- [Upsert com PK string] → usar `OnConflict` (Postgres) garante atomicidade; `Save` também funciona pois a PK é conhecida. Escolher `OnConflict` para evitar corrida read-then-write.
- [Defaults divergirem do frontend] → manter Ações 50 / FIIs 40 / ETFs 10 / limite 20, iguais ao `AllocationService` do front.

## Migration Plan

`AutoMigrate` cria a tabela nova; sem dado legado. Rollback = remover rota/migração; nenhuma outra feature depende disso.

## Open Questions

- Persistir o limite e os alvos como colunas (escolhido) vs um JSON único — colunas são mais simples de consultar e bastam aqui.
