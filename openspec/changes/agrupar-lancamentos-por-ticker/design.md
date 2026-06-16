## Context

A consulta de posições de ações (`GetAcoesPositions`) já agrega lançamentos via `GROUP BY ticker` somando quantidade e preço médio ponderado. Porém o ticker é gravado exatamente como recebido na API, sem normalização. Como `GROUP BY ticker` é sensível à caixa e espaços (no PostgreSQL e SQLite), `PETR4`, `petr4` e `"PETR4 "` formam grupos distintos — fazendo o mesmo ativo aparecer duplicado em "Meus Ativos".

Além disso, o cálculo de patrimônio das Metas (`GoalHandler.buildGoalResponses`) usa exclusivamente `GetAcoesPositions`, que filtra `asset_type = 'Acoes'`. FIIs e ETFs ficam de fora do patrimônio, subestimando o progresso das metas.

As respostas de criação/edição de lançamentos e metas retornam apenas o recurso, sem uma mensagem de resultado para as telas exibirem ao usuário.

## Goals / Non-Goals

**Goals:**
- Tratar múltiplos lançamentos do mesmo ticker como uma única posição consolidada, sem duplicação.
- Normalizar ticker (uppercase + trim) na escrita e consolidar dados legados na leitura.
- Incluir FIIs e ETFs no patrimônio das Metas.
- Retornar mensagens de resultado nas respostas de lançamentos e metas.
- Expor a contagem de lançamentos consolidados por posição.

**Non-Goals:**
- Migração/backfill destrutivo dos tickers já gravados (a consolidação na leitura cobre o legado).
- Mudanças na tela frontend (apenas o contrato da API que ela consome).
- Alteração do algoritmo de nota (`computeNotas`) ou do cálculo de dividendos.

## Decisions

### 1. Normalizar ticker na camada de aplicação
Normalizar (`strings.ToUpper(strings.TrimSpace(ticker))`) em `TransactionService.Create` e `Update`, e não apenas no handler. Garante consistência independente do ponto de entrada e mantém o handler fino.
- *Alternativa considerada*: normalizar só no handler HTTP — rejeitada por deixar a regra de domínio fora do serviço e vulnerável a novos chamadores.

### 2. Agrupar por ticker normalizado na query (cobre legado)
Em `GetAcoesPositions`, agrupar por `UPPER(TRIM(ticker))` e retornar o ticker normalizado, em vez de `GROUP BY ticker` cru. Assim dados legados gravados com caixa/espaço divergentes consolidam sem precisar de migração.
- *Alternativa considerada*: migração em massa dos registros existentes — mais arriscada e desnecessária dado que a normalização na leitura resolve a exibição.

### 3. Contagem de lançamentos na posição
Adicionar `TransactionCount` em `domain.AcoesPosition` e `domain.AcaoItem`, populado por `COUNT(*)` no `GROUP BY`. A tela de Meus Ativos usa esse número para sinalizar a consolidação.

### 4. Novo método de posições para todos os tipos de ativo
Adicionar `GetAllPositions(userID)` ao `TransactionRepository`/`TransactionUseCase`, idêntico a `GetAcoesPositions` mas **sem** o filtro `asset_type = 'Acoes'`. O `GoalHandler` passa a usar `GetAllPositions` para compor o patrimônio (Ações + FIIs + ETFs). `GetAcoesPositions` permanece para a tela de Meus Ativos (que é específica de ações).
- *Alternativa considerada*: parametrizar `GetAcoesPositions` por tipo — rejeitada para manter assinatura clara e o `GROUP BY` igual; um método dedicado é mais legível.

### 5. Mensagens de resultado nas respostas
Adicionar um campo `message` (string) nas respostas de criação/edição de lançamentos e metas, sem quebrar os campos existentes. As telas de Meus Ativos e Metas exibem essa mensagem como feedback da ação.

## Risks / Trade-offs

- [`UPPER(TRIM(...))` no `GROUP BY` pode não usar índice] → O volume de transações por usuário é pequeno (carteira pessoal); o custo é desprezível. Caso cresça, criar índice funcional.
- [Diferença de SQL entre SQLite e PostgreSQL] → `UPPER` e `TRIM` são padrão ANSI e suportados por ambos; o `avg_price` continua `SUM(quantity*price)/SUM(quantity)`.
- [Posições de FIIs/ETFs sem preço de mercado disponível no Yahoo] → o cálculo usa o preço médio (`AvgPrice`) como fallback, como já ocorre para ações.
- [Adição de campo `message`/`transaction_count` nas respostas] → aditivo e retrocompatível; clientes que ignoram campos extras não quebram.

## Migration Plan

1. Implementar normalização e novo método de repositório.
2. `go build ./...`, `go vet ./...`, `go test ./...`.
3. Deploy normal (sem migração de dados; a consolidação na leitura cobre o legado).
4. Rollback: reverter o commit; nenhum schema alterado, sem dado destruído.

## Open Questions

- Nenhuma. As decisões acima cobrem os requisitos do proposal e specs.
