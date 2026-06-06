## 1. Domain Model

- [x] 1.1 Add `DY float64` field to `Stock` struct in `internal/domain/stock.go` with GORM tag `json:"dy"` and validate `DY >= 0` in `Validate()`

## 2. DTOs

- [x] 2.1 Add `DY float64` field with `json:"dy"` to `CreateStockRequest` in `internal/adapters/http/dto/stock_dto.go`
- [x] 2.2 Add `DY float64` field with `json:"dy"` to `UpdateStockRequest` in `internal/adapters/http/dto/stock_dto.go`
- [x] 2.3 Add `DY float64` field with `json:"dy"` to `StockResponse` and map it in `FromDomain()` in `internal/adapters/http/dto/stock_dto.go`

## 3. HTTP Handler

- [x] 3.1 Map `req.DY` to `stock.DY` in `CreateStock` in `internal/adapters/http/handler/stock_handler.go`
- [x] 3.2 Map `req.DY` to `updated.DY` in `UpdateStock` in `internal/adapters/http/handler/stock_handler.go`
- [x] 3.3 Add `"dy"` to the accepted `sort` values allowlist in `ListStocks` (error message and validation) in `internal/adapters/http/handler/stock_handler.go`

## 4. Service

- [x] 4.1 Copy `updated.DY` to `existing.DY` in `UpdateStock` in `internal/application/stock_service.go`

## 5. Repository / Persistence

- [x] 5.1 Verify `gorm_stock_repository.go` uses `AutoMigrate(&domain.Stock{})` so the new `dy` column is created on startup — no manual SQL needed

## 6. Sorting

- [x] 6.1 Add `case "dy": query = query.Order("dy desc")` to the sort switch in `internal/infrastructure/persistence/gorm_stock_repository.go`

## 7. Tests

- [x] 7.1 Update `CreateStock` unit test in `stock_handler_test.go` to include `dy` in request and assert it in response
- [x] 7.2 Update `UpdateStock` unit test in `stock_handler_test.go` to include `dy`
- [x] 7.3 Update service test in `stock_service_test.go` to cover `DY` propagation in `UpdateStock`
