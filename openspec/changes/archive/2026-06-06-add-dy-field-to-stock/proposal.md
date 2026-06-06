## Why

The frontend at `http://localhost:4200/` displays and submits a `dy` (Dividend Yield) field for stocks, but the API does not yet persist or return this value, causing data loss on every create and update operation.

## What Changes

- Add `dy` (`float64`) field to the `Stock` domain model and GORM table via migration
- Expose `dy` in `CreateStockRequest`, `UpdateStockRequest`, and `StockResponse` DTOs
- Propagate `dy` through the handler's create and update actions and the service's `UpdateStock` logic
- Allow `sort=dy` in the `ListStocks` query parameter alongside the existing `nota` and `variacao` options

## Capabilities

### New Capabilities
- none

### Modified Capabilities
- `stock-management`: `Stock` entity gains a `dy` field (Dividend Yield, `float64`, optional, ≥ 0); all CRUD endpoints accept and return it; `ListStocks` supports `sort=dy`

## Impact

- `internal/domain/stock.go` — add `DY float64` field; validate `DY >= 0`
- `internal/adapters/http/dto/stock_dto.go` — add `DY` to all request/response structs
- `internal/adapters/http/handler/stock_handler.go` — map `DY` in `CreateStock` and `UpdateStock`; accept `sort=dy`
- `internal/application/stock_service.go` — copy `DY` in `UpdateStock`
- `internal/infrastructure/persistence/gorm_stock_repository.go` — GORM auto-migrate picks up the new column
- No breaking changes; existing records will default `dy` to `0.0`
