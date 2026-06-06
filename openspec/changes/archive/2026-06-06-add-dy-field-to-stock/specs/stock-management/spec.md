## MODIFIED Requirements

### Requirement: Criar ação no portfólio
O sistema SHALL permitir adicionar uma nova ação ao portfólio com os campos: ticker (símbolo único em maiúsculas), nome, setor, nota (0.0–10.0), preço atual (positivo), variação do dia (percentual) e dy (Dividend Yield, percentual, opcional, ≥ 0).

#### Scenario: Criação bem-sucedida de ação
- **WHEN** o cliente envia POST `/api/v1/stocks` com payload JSON válido contendo ticker, nome, setor, nota, preço e variação
- **THEN** o sistema SHALL persistir a ação e retornar 201 Created com o objeto criado incluindo ID gerado e campo `dy`

#### Scenario: Criação com dy informado
- **WHEN** o cliente envia POST `/api/v1/stocks` com payload JSON contendo `dy` com valor ≥ 0
- **THEN** o sistema SHALL persistir o valor de `dy` e retorná-lo no objeto de resposta

#### Scenario: Criação sem dy informado
- **WHEN** o cliente envia POST `/api/v1/stocks` sem o campo `dy`
- **THEN** o sistema SHALL persistir a ação com `dy` igual a `0.0` e retornar 201 Created

#### Scenario: Criação com ticker duplicado
- **WHEN** o cliente envia POST `/api/v1/stocks` com um ticker já existente no portfólio
- **THEN** o sistema SHALL retornar 409 Conflict com mensagem de erro indicando duplicidade

#### Scenario: Criação com campos obrigatórios ausentes
- **WHEN** o cliente envia POST `/api/v1/stocks` sem ticker ou nome
- **THEN** o sistema SHALL retornar 400 Bad Request com mensagem descritiva dos campos inválidos

#### Scenario: Criação com nota fora do range
- **WHEN** o cliente envia POST `/api/v1/stocks` com nota menor que 0 ou maior que 10
- **THEN** o sistema SHALL retornar 400 Bad Request

#### Scenario: Criação com preço negativo
- **WHEN** o cliente envia POST `/api/v1/stocks` com preço atual menor ou igual a zero
- **THEN** o sistema SHALL retornar 400 Bad Request

### Requirement: Atualizar ação no portfólio
O sistema SHALL permitir atualizar parcialmente ou totalmente os dados de uma ação existente, incluindo o campo `dy`.

#### Scenario: Atualização bem-sucedida
- **WHEN** o cliente envia PUT `/api/v1/stocks/:id` com payload JSON e ID de ação existente
- **THEN** o sistema SHALL atualizar os campos fornecidos, incluindo `dy`, e retornar 200 OK com o objeto atualizado

#### Scenario: Atualização com dy
- **WHEN** o cliente envia PUT `/api/v1/stocks/:id` com payload contendo `dy`
- **THEN** o sistema SHALL persistir o novo valor de `dy` e retorná-lo no objeto de resposta

#### Scenario: Atualização de ação inexistente
- **WHEN** o cliente envia PUT `/api/v1/stocks/:id` com ID inexistente
- **THEN** o sistema SHALL retornar 404 Not Found
