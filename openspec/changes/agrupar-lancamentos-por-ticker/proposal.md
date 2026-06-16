## Why

Hoje o usuário pode cadastrar vários lançamentos (compras) para o mesmo ticker, mas ao consultar "Meus Ativos" o mesmo ativo aparece duplicado em vez de consolidado numa única posição. A causa é que os tickers são gravados sem normalização (caixa e espaços), então `PETR4`, `petr4` e `PETR4 ` viram grupos distintos no agrupamento por ticker. Além disso, ações de cadastro/lançamento não retornam feedback claro para as telas de Meus Ativos e Metas.

## What Changes

- Normalizar o ticker (uppercase + trim) na criação e edição de lançamentos, garantindo que múltiplos lançamentos do mesmo ativo sejam tratados como um único ticker.
- Garantir que a consulta de posições (`GET /transactions/acoes`) agrupe e some corretamente todos os lançamentos de um mesmo ticker e usuário, mesmo para dados legados gravados com caixa/espaço inconsistentes (agrupar por ticker normalizado).
- Incluir a contagem de lançamentos consolidados em cada posição retornada, permitindo às telas de Meus Ativos sinalizar que vários lançamentos foram somados.
- Retornar mensagens de resultado nas respostas de criação/edição de lançamentos e metas, para que as telas de Meus Ativos e Metas informem ao usuário o resultado da ação.
- Considerar **todos os tipos de ativo** (Ações, FIIs e ETFs) no cálculo do patrimônio das Metas, somando todos os lançamentos consolidados por ticker — hoje apenas Ações são contabilizadas.

## Capabilities

### New Capabilities
- `transaction-positions`: cadastro de múltiplos lançamentos por ticker/usuário, normalização de ticker e consolidação (soma) das posições na consulta — incluindo o cálculo do patrimônio das Metas sobre todos os tipos de ativo (Ações, FIIs e ETFs) — com feedback de resultado da ação.

### Modified Capabilities
<!-- Nenhuma capability existente tem requisitos alterados. -->

## Impact

- `internal/domain/transaction.go`: normalização de ticker; campo de contagem de lançamentos em `AcoesPosition`/`AcaoItem`.
- `internal/application/transaction_service.go`: normalização na criação/edição.
- `internal/infrastructure/persistence/gorm_transaction_repository.go`: agrupamento por ticker normalizado em `GetAcoesPositions`; novo método para consolidar posições de todos os tipos de ativo (para o cálculo de Metas).
- `internal/domain/transaction_repository.go` e `internal/application/transaction_service.go`: novo método de posições agregadas por todos os tipos de ativo.
- `internal/adapters/http/handler/transaction_handler.go` e `goal_handler.go`: cálculo de patrimônio das Metas sobre todos os ativos; mensagens de resultado nas respostas.
- `internal/adapters/http/dto/transaction_dto.go` e `goal_dto.go`: campos de mensagem/contagem nas respostas.
- APIs afetadas: `POST/PUT /transactions`, `GET /transactions/acoes`, `POST/PUT /goals`.
