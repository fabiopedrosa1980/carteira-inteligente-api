## Why

O frontend em `http://localhost:4200/` exibe um resumo mensal de dividendos (Janeiro a Dezembro) com as ações pagadoras, quantidade, total médio e yield médio, mas a API não possui nenhum modelo de dividendos nem endpoint para fornecer esses dados — o frontend usa dados gerados localmente sem persistência.

## What Changes

- Criar a entidade `Dividend` no domínio (valor, mês, ano, tipo, exDate, payDate, associada a um Stock)
- Adicionar tabela `dividends` persistida via GORM auto-migrate
- Criar endpoint `GET /api/v1/dividends/monthly` que retorna resumo mensal de Janeiro a Dezembro
- Criar endpoint `POST /api/v1/stocks/:id/dividends` para cadastrar dividendos de uma ação
- Criar endpoint `GET /api/v1/stocks/:id/dividends` para listar dividendos de uma ação

## Capabilities

### New Capabilities
- `dividend-management`: CRUD de dividendos associados a ações (cadastro e listagem por ação)
- `monthly-dividend-summary`: Endpoint de resumo mensal de dividendos para todos os meses do ano, com ações pagadoras, quantidade, total médio e yield médio

### Modified Capabilities
- `stock-management`: Nenhuma alteração nos requisitos de Stock; a relação com dividendos é nova

## Impact

- `internal/domain/` — nova entidade `Dividend` e repositório `DividendRepository`
- `internal/application/` — novo `DividendService` com casos de uso
- `internal/adapters/http/dto/` — novos DTOs para dividendos e resumo mensal
- `internal/adapters/http/handler/` — novo `DividendHandler`
- `internal/adapters/http/router/` — novas rotas
- `internal/infrastructure/persistence/` — `GormDividendRepository` + auto-migrate
- Nenhuma breaking change nas rotas existentes
