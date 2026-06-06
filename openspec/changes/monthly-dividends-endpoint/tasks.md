## 1. Domain Model

- [x] 1.1 Criar `internal/domain/dividend.go` com struct `Dividend` (ID, StockID, Amount, Month, Year, Type, ExDate, PayDate, CreatedAt) e método `Validate()`
- [x] 1.2 Criar `internal/domain/dividend_repository.go` com interface `DividendRepository` (Create, FindByStockID, FindByStockIDAndYear, FindByYear)

## 2. DTOs

- [x] 2.1 Criar `internal/adapters/http/dto/dividend_dto.go` com `CreateDividendRequest`, `DividendResponse` e `MonthSummaryResponse`
- [x] 2.2 Implementar `DividendFromDomain()` e `DividendListFromDomain()` em `dividend_dto.go`

## 3. Application Layer

- [x] 3.1 Criar `internal/application/dividend_service.go` com interface `DividendUseCase` e struct `DividendService`
- [x] 3.2 Implementar `CreateDividend(stockID uint, d *domain.Dividend) error` — valida que o stock existe antes de criar
- [x] 3.3 Implementar `ListDividendsByStock(stockID uint, year *int) ([]domain.Dividend, error)`
- [x] 3.4 Implementar `GetMonthlySummary(year int) ([]dto.MonthSummaryResponse, error)` — agrega dividendos por mês, calcula avg_total e avg_yield usando preco_atual do stock

## 4. Infrastructure

- [x] 4.1 Criar `internal/infrastructure/persistence/gorm_dividend_repository.go` implementando `DividendRepository`
- [x] 4.2 Adicionar `db.AutoMigrate(&domain.Dividend{})` em `internal/infrastructure/persistence/database.go`

## 5. HTTP Handler e Rotas

- [x] 5.1 Criar `internal/adapters/http/handler/dividend_handler.go` com `DividendHandler` e métodos `CreateDividend`, `ListDividends`, `GetMonthlySummary`
- [x] 5.2 Adicionar rotas em `internal/adapters/http/router/router.go`:
  - `POST /api/v1/stocks/:id/dividends`
  - `GET  /api/v1/stocks/:id/dividends`
  - `GET  /api/v1/dividends/monthly`
- [x] 5.3 Atualizar `cmd/api/main.go` para instanciar `GormDividendRepository`, `DividendService` e `DividendHandler` e passá-los ao router

## 6. Testes

- [x] 6.1 Criar `internal/adapters/http/handler/dividend_handler_test.go` com teste de `POST /api/v1/stocks/:id/dividends` retornando 201
- [x] 6.2 Adicionar teste de `GET /api/v1/stocks/:id/dividends` retornando lista com os dividendos criados
- [x] 6.3 Adicionar teste de `GET /api/v1/dividends/monthly` retornando sempre 12 meses com campos corretos
- [x] 6.4 Adicionar teste de `GET /api/v1/dividends/monthly?year=2024` e verificar avg_total e stock_count nos meses com dados
