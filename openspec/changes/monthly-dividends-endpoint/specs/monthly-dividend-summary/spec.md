## ADDED Requirements

### Requirement: Consultar resumo mensal de dividendos
O sistema SHALL fornecer um resumo anual de dividendos agrupado por mês (Janeiro a Dezembro), contendo para cada mês: lista de tickers que pagam, quantidade de ações pagadoras, total médio de dividendo por ação e yield médio (calculado como `amount / preco_atual * 100` para cada ação pagadora). O ano de referência é informado via query parameter `year` (padrão: ano corrente).

#### Scenario: Resumo com dividendos cadastrados
- **WHEN** o cliente envia GET `/api/v1/dividends/monthly?year=2024`
- **THEN** o sistema SHALL retornar 200 OK com array de 12 objetos, um por mês, cada um contendo: `month` (1–12), `month_name` (nome em português), `tickers` (array de strings), `stock_count` (inteiro), `avg_total` (float, média dos valores pagos no mês) e `avg_yield` (float, yield médio percentual)

#### Scenario: Resumo com ano padrão
- **WHEN** o cliente envia GET `/api/v1/dividends/monthly` sem o parâmetro `year`
- **THEN** o sistema SHALL usar o ano corrente como referência e retornar o array de 12 meses

#### Scenario: Mês sem pagamentos
- **WHEN** nenhuma ação possui dividendo cadastrado em determinado mês do ano solicitado
- **THEN** o sistema SHALL incluir o mês no resultado com `tickers: []`, `stock_count: 0`, `avg_total: 0` e `avg_yield: 0`

#### Scenario: Ano inválido
- **WHEN** o cliente envia GET `/api/v1/dividends/monthly?year=abc`
- **THEN** o sistema SHALL retornar 400 Bad Request com mensagem de erro

#### Scenario: Sempre retorna 12 meses
- **WHEN** o cliente envia GET `/api/v1/dividends/monthly` independente de quantos dados existem
- **THEN** o sistema SHALL sempre retornar exatamente 12 entradas no array, uma para cada mês
