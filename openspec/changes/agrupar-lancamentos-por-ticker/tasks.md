## 1. Domínio

- [x] 1.1 Adicionar função de normalização de ticker (`ToUpper` + `TrimSpace`) em `internal/domain/transaction.go`
- [x] 1.2 Adicionar campo `TransactionCount` em `domain.AcoesPosition` e `domain.AcaoItem`

## 2. Normalização e múltiplos lançamentos

- [x] 2.1 Normalizar o ticker em `TransactionService.Create` antes de persistir
- [x] 2.2 Normalizar o ticker em `TransactionService.Update` antes de persistir
- [x] 2.3 Confirmar que não há restrição de unicidade que rejeite múltiplos lançamentos do mesmo ticker/usuário

## 3. Consolidação na consulta

- [x] 3.1 Em `GetAcoesPositions`, agrupar por `UPPER(TRIM(ticker))`, retornar o ticker normalizado e incluir `COUNT(*)` como contagem de lançamentos
- [x] 3.2 Adicionar método `GetAllPositions(userID)` ao `TransactionRepository` (sem filtro de `asset_type`) e à interface `TransactionUseCase`
- [x] 3.3 Implementar `GetAllPositions` em `GormTransactionRepository` com o mesmo agrupamento normalizado

## 4. Metas com todos os tipos de ativo

- [x] 4.1 Em `GoalHandler.buildGoalResponses`, trocar `GetAcoesPositions` por `GetAllPositions` para somar Ações, FIIs e ETFs no patrimônio
- [x] 4.2 Validar que o `currentValue` e `progressPercent` refletem todos os tipos de ativo

## 5. Feedback nas respostas (Meus Ativos e Metas)

- [x] 5.1 Adicionar campo `transaction_count` em `AcaoItem`/resposta de `GET /transactions/acoes`
- [x] 5.2 Adicionar campo `message` nas respostas de criação/edição de lançamentos (`transaction_dto.go`)
- [x] 5.3 Adicionar campo `message` nas respostas de criação/edição de metas (`goal_dto.go`) e preencher nos handlers

## 6. Testes e verificação

- [x] 6.1 Testes de unidade para normalização de ticker e consolidação de lançamentos por ticker
- [x] 6.2 Teste de consolidação de dados legados com caixa divergente (`PETR4` vs `petr4`)
- [x] 6.3 Teste do patrimônio das Metas somando Ações, FIIs e ETFs
- [x] 6.4 `go build ./...`, `go vet ./...`, `go test ./...`
- [x] 6.5 Commit e push seguindo Conventional Commits
