## ADDED Requirements

### Requirement: Cadastrar dividendo de uma ação
O sistema SHALL permitir registrar um pagamento de dividendo associado a uma ação existente, contendo: valor (`amount`, positivo), mês (1–12), ano (≥ 2000), tipo (`dividendo`, `jcp` ou `rendimento`), data ex (`ex_date`, formato YYYY-MM-DD) e data de pagamento (`pay_date`, formato YYYY-MM-DD).

#### Scenario: Cadastro bem-sucedido
- **WHEN** o cliente envia POST `/api/v1/stocks/:id/dividends` com payload JSON válido e ID de ação existente
- **THEN** o sistema SHALL persistir o dividendo e retornar 201 Created com o objeto criado incluindo ID gerado

#### Scenario: Ação não encontrada
- **WHEN** o cliente envia POST `/api/v1/stocks/:id/dividends` com ID de ação inexistente
- **THEN** o sistema SHALL retornar 404 Not Found

#### Scenario: Payload inválido — amount negativo ou zero
- **WHEN** o cliente envia POST `/api/v1/stocks/:id/dividends` com `amount` ≤ 0
- **THEN** o sistema SHALL retornar 400 Bad Request

#### Scenario: Payload inválido — mês fora do range
- **WHEN** o cliente envia POST `/api/v1/stocks/:id/dividends` com `month` menor que 1 ou maior que 12
- **THEN** o sistema SHALL retornar 400 Bad Request

#### Scenario: Payload inválido — tipo inválido
- **WHEN** o cliente envia POST `/api/v1/stocks/:id/dividends` com `type` diferente de `dividendo`, `jcp` ou `rendimento`
- **THEN** o sistema SHALL retornar 400 Bad Request

### Requirement: Listar dividendos de uma ação
O sistema SHALL permitir recuperar todos os dividendos registrados para uma ação específica, opcionalmente filtrados por ano.

#### Scenario: Listagem sem filtro
- **WHEN** o cliente envia GET `/api/v1/stocks/:id/dividends` com ID de ação existente
- **THEN** o sistema SHALL retornar 200 OK com array de dividendos da ação em ordem cronológica (ano, mês)

#### Scenario: Listagem filtrada por ano
- **WHEN** o cliente envia GET `/api/v1/stocks/:id/dividends?year=2024`
- **THEN** o sistema SHALL retornar apenas os dividendos do ano informado

#### Scenario: Ação sem dividendos
- **WHEN** o cliente envia GET `/api/v1/stocks/:id/dividends` e a ação não possui dividendos cadastrados
- **THEN** o sistema SHALL retornar 200 OK com array vazio `[]`

#### Scenario: Ação não encontrada
- **WHEN** o cliente envia GET `/api/v1/stocks/:id/dividends` com ID inexistente
- **THEN** o sistema SHALL retornar 404 Not Found
