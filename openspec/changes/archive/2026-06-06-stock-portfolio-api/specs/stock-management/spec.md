## ADDED Requirements

### Requirement: Criar ação no portfólio
O sistema SHALL permitir adicionar uma nova ação ao portfólio com os campos: ticker (símbolo único em maiúsculas), nome, setor, nota (0.0–10.0), preço atual (positivo) e variação do dia (percentual).

#### Scenario: Criação bem-sucedida de ação
- **WHEN** o cliente envia POST `/api/v1/stocks` com payload JSON válido contendo ticker, nome, setor, nota, preço e variação
- **THEN** o sistema SHALL persistir a ação e retornar 201 Created com o objeto criado incluindo ID gerado

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

### Requirement: Consultar ação por ID
O sistema SHALL permitir recuperar os dados de uma ação específica pelo seu ID.

#### Scenario: Ação encontrada
- **WHEN** o cliente envia GET `/api/v1/stocks/:id` com ID de uma ação existente
- **THEN** o sistema SHALL retornar 200 OK com o objeto completo da ação

#### Scenario: Ação não encontrada
- **WHEN** o cliente envia GET `/api/v1/stocks/:id` com ID inexistente
- **THEN** o sistema SHALL retornar 404 Not Found com mensagem de erro

### Requirement: Atualizar ação no portfólio
O sistema SHALL permitir atualizar parcialmente ou totalmente os dados de uma ação existente.

#### Scenario: Atualização bem-sucedida
- **WHEN** o cliente envia PUT `/api/v1/stocks/:id` com payload JSON e ID de ação existente
- **THEN** o sistema SHALL atualizar os campos fornecidos e retornar 200 OK com o objeto atualizado

#### Scenario: Atualização de ação inexistente
- **WHEN** o cliente envia PUT `/api/v1/stocks/:id` com ID inexistente
- **THEN** o sistema SHALL retornar 404 Not Found

### Requirement: Remover ação do portfólio
O sistema SHALL permitir excluir uma ação do portfólio pelo ID.

#### Scenario: Remoção bem-sucedida
- **WHEN** o cliente envia DELETE `/api/v1/stocks/:id` com ID de ação existente
- **THEN** o sistema SHALL remover a ação e retornar 204 No Content

#### Scenario: Remoção de ação inexistente
- **WHEN** o cliente envia DELETE `/api/v1/stocks/:id` com ID inexistente
- **THEN** o sistema SHALL retornar 404 Not Found
