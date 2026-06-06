## 1. Scaffolding do Projeto

- [x] 1.1 Inicializar módulo Go com `go mod init` e definir nome do módulo
- [x] 1.2 Adicionar dependências: `gin-gonic/gin`, `gorm.io/gorm`, `gorm.io/driver/sqlite`
- [x] 1.3 Criar estrutura de diretórios: `cmd/api`, `internal/domain`, `internal/application`, `internal/infrastructure/persistence`, `internal/adapters/http/handler`, `internal/adapters/http/dto`, `internal/adapters/http/router`

## 2. Camada de Domínio

- [x] 2.1 Criar entidade `Stock` em `internal/domain/stock.go` com campos: ID, Ticker, Nome, Setor, Nota, PrecoAtual, VariacaoHoje, CreatedAt, UpdatedAt
- [x] 2.2 Definir port `StockRepository` (interface) em `internal/domain/stock_repository.go` com métodos: Create, FindByID, FindAll, Update, Delete
- [x] 2.3 Adicionar validações de domínio na entidade: Nota entre 0–10, PrecoAtual > 0, Ticker não vazio

## 3. Camada de Infraestrutura

- [x] 3.1 Criar `GormStockRepository` em `internal/infrastructure/persistence/gorm_stock_repository.go` implementando a interface `StockRepository`
- [x] 3.2 Implementar método `Create` com tratamento de erro para ticker duplicado (unique constraint)
- [x] 3.3 Implementar métodos `FindByID`, `FindAll` (com filtro por setor e ordenação por nota/variacao), `Update` e `Delete`
- [x] 3.4 Criar função de setup do banco de dados em `internal/infrastructure/persistence/database.go` usando SQLite in-memory com `AutoMigrate`

## 4. Camada de Aplicação

- [x] 4.1 Criar `StockService` em `internal/application/stock_service.go` recebendo `StockRepository` via injeção de dependência
- [x] 4.2 Implementar use case `CreateStock` com validação de campos obrigatórios
- [x] 4.3 Implementar use cases `GetStockByID`, `ListStocks` (com parâmetros de filtro e sort), `UpdateStock` e `DeleteStock`
- [x] 4.4 Retornar erros tipados do domínio (ErrNotFound, ErrDuplicate, ErrValidation) para o serviço distinguir e mapear para HTTP status correto

## 5. DTOs e Validação HTTP

- [x] 5.1 Criar `CreateStockRequest` e `UpdateStockRequest` em `internal/adapters/http/dto/stock_dto.go` com tags de binding do Gin
- [x] 5.2 Criar `StockResponse` mapeando a entidade de domínio para o formato JSON da resposta
- [x] 5.3 Adicionar validações de binding: `required` para ticker e nome, `min=0,max=10` para nota, `gt=0` para preço

## 6. Handlers HTTP

- [x] 6.1 Criar `StockHandler` em `internal/adapters/http/handler/stock_handler.go` recebendo `StockService` via injeção
- [x] 6.2 Implementar handler `CreateStock` (POST `/api/v1/stocks`) — bind, validação, 201/400/409
- [x] 6.3 Implementar handler `GetStock` (GET `/api/v1/stocks/:id`) — 200/404
- [x] 6.4 Implementar handler `ListStocks` (GET `/api/v1/stocks`) — query params `setor` e `sort`, 200/400
- [x] 6.5 Implementar handler `UpdateStock` (PUT `/api/v1/stocks/:id`) — bind, 200/400/404
- [x] 6.6 Implementar handler `DeleteStock` (DELETE `/api/v1/stocks/:id`) — 204/404

## 7. Router e Bootstrap

- [x] 7.1 Criar `SetupRouter` em `internal/adapters/http/router/router.go` registrando todas as rotas do grupo `/api/v1/stocks`
- [x] 7.2 Criar `cmd/api/main.go` com wiring de dependências: DB → Repository → Service → Handler → Router
- [x] 7.3 Configurar porta do servidor (padrão 8080) via variável de ambiente `PORT`

## 8. Testes e Verificação

- [x] 8.1 Escrever testes unitários para `StockService` mockando o repositório
- [x] 8.2 Escrever testes de integração dos handlers com banco in-memory usando `httptest`
- [x] 8.3 Verificar todos os cenários das specs: criação, duplicata, campos inválidos, not found, filtro por setor, ordenação por nota e variação
- [x] 8.4 Garantir que o servidor inicia e responde corretamente com `go run ./cmd/api`
