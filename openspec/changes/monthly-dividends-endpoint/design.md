## Context

A API usa arquitetura limpa (domain → application → infrastructure/adapters). O `Stock` já existe e é gerenciado pelo GORM. O frontend gera dividendos localmente com padrões fixos por ticker; esta mudança torna os dividendos dados de primeira classe na API, permitindo persistência e consulta real.

A estrutura atual não possui nenhuma entidade de dividendo — tudo será adicionado sem breaking changes nas rotas existentes.

## Goals / Non-Goals

**Goals:**
- Nova entidade `Dividend` com relacionamento `stock_id → stocks.id`
- GORM auto-migrate cria a tabela `dividends`
- `POST /api/v1/stocks/:id/dividends` — cadastro de dividendo
- `GET /api/v1/stocks/:id/dividends` — listagem com filtro opcional por `?year=`
- `GET /api/v1/dividends/monthly` — resumo mensal com `?year=` (padrão: ano corrente)
- Resumo calcula `avg_yield` usando `preco_atual` do `Stock` associado

**Non-Goals:**
- Importação automática de dividendos de fontes externas
- Histórico de preços para cálculo de yield por data ex
- Paginação dos endpoints de dividendos (volumes pequenos esperados)
- Autenticação ou controle de acesso

## Decisions

**D1 — Dividend como entidade separada com FK para Stock**
Relacionamento `belongs-to` via `StockID uint`. GORM gerencia a FK. Alternativa (campo JSON no Stock) rejeitada: não permite queries eficientes por mês/ano.

**D2 — Resumo mensal calculado na camada de aplicação, não em SQL**
A consulta agrupa todos os dividendos do ano, depois itera os 12 meses na memória. O volume de dividendos é pequeno (≤ 10 ações × 12 meses × N anos), então a abordagem in-memory é simples e suficiente. Alternativa (GROUP BY em SQL) mais complexa sem ganho real nessa escala.

**D3 — avg_yield = mean(amount / stock.preco_atual * 100) para ações com preco_atual > 0**
Usa o preço atual do Stock (snapshot mais recente). Ações com `preco_atual = 0` são excluídas do cálculo de yield para evitar divisão por zero, mas ainda contam em `stock_count` e `avg_total`.

**D4 — month_name em português hardcoded**
Array fixo `["Janeiro", ..., "Dezembro"]` na camada de aplicação, igual ao padrão do frontend. Sem dependência de i18n.

**D5 — Rota de resumo mensal em `/api/v1/dividends/monthly`**
Separada das rotas de ação para deixar explícito que é um recurso agregado. Rotas de dividendos por ação ficam em `/api/v1/stocks/:id/dividends`.

## Risks / Trade-offs

- [Risk] `preco_atual` é mutável; yield calculado muda conforme preço da ação → Aceito; yield é uma estimativa com o preço atual, não histórico.
- [Risk] Sem validação de duplicata (mesmo mês/ano/ação pode ser inserido duas vezes) → Aceito para MVP; frontend pode controlar duplicatas.

## Migration Plan

1. Mergar e fazer deploy.
2. GORM `AutoMigrate` cria tabela `dividends` com FK `stock_id`.
3. Dados existentes em `stocks` não são afetados.
4. **Rollback**: reverter deploy; coluna `dividends` fica orfã mas não quebra a aplicação anterior (ela não conhecia a tabela).
