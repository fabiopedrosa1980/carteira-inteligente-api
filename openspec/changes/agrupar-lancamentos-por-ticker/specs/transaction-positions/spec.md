## ADDED Requirements

### Requirement: Cadastrar múltiplos lançamentos por ticker
O sistema SHALL permitir que um usuário registre vários lançamentos (compras) para o mesmo ticker, sem rejeitar lançamentos duplicados de um ativo já existente na carteira.

#### Scenario: Segundo lançamento do mesmo ticker
- **WHEN** o usuário envia POST `/transactions` para um ticker que já possui lançamentos anteriores
- **THEN** o sistema SHALL criar um novo lançamento e retornar 201 Created sem erro de duplicidade

### Requirement: Normalizar ticker dos lançamentos
O sistema SHALL normalizar o ticker (remover espaços nas extremidades e converter para caixa alta) ao criar e ao editar um lançamento, garantindo que variações de digitação do mesmo ativo sejam tratadas como um único ticker.

#### Scenario: Ticker com caixa e espaço inconsistentes
- **WHEN** o usuário envia POST `/transactions` com ticker `" petr4 "`
- **THEN** o sistema SHALL persistir o ticker como `PETR4`

#### Scenario: Edição normaliza o ticker
- **WHEN** o usuário envia PUT `/transactions/:id` cuja requisição altera o ticker para `petr4`
- **THEN** o sistema SHALL persistir o ticker como `PETR4`

### Requirement: Consolidar posições por ticker na consulta
O sistema SHALL, ao consultar as posições de ações (`GET /transactions/acoes`), somar todos os lançamentos de um mesmo ticker e usuário em uma única posição, agrupando por ticker normalizado para que dados legados com caixa/espaço inconsistentes também sejam consolidados.

#### Scenario: Vários lançamentos do mesmo ticker somados
- **WHEN** o usuário possui dois lançamentos de `PETR4` (10 e 20 cotas) e consulta GET `/transactions/acoes`
- **THEN** o sistema SHALL retornar uma única posição de `PETR4` com quantidade total 30 e preço médio ponderado pelos lançamentos

#### Scenario: Dados legados com caixa divergente consolidados
- **WHEN** o usuário possui lançamentos gravados como `PETR4` e `petr4` e consulta GET `/transactions/acoes`
- **THEN** o sistema SHALL retornar uma única posição consolidada do ticker e não duas posições duplicadas

### Requirement: Informar quantidade de lançamentos consolidados
O sistema SHALL incluir, em cada posição retornada por `GET /transactions/acoes`, a quantidade de lançamentos que foram somados para compor aquela posição, permitindo que a tela de Meus Ativos sinalize a consolidação ao usuário.

#### Scenario: Posição com múltiplos lançamentos
- **WHEN** o usuário consulta GET `/transactions/acoes` e a posição de `PETR4` resulta de 3 lançamentos
- **THEN** a posição retornada SHALL conter a contagem de lançamentos igual a 3

### Requirement: Patrimônio das Metas considera todos os tipos de ativo
O sistema SHALL calcular o patrimônio atual das Metas somando os lançamentos consolidados de todos os tipos de ativo do usuário (Ações, FIIs e ETFs), e não apenas de Ações.

#### Scenario: Carteira com Ações, FIIs e ETFs
- **WHEN** o usuário possui lançamentos de Ações, FIIs e ETFs e consulta GET `/goals`
- **THEN** o `currentValue` de cada meta SHALL refletir a soma do valor de mercado das posições de todos os três tipos de ativo

#### Scenario: Carteira somente com FIIs
- **WHEN** o usuário possui lançamentos apenas de FIIs e consulta GET `/goals`
- **THEN** o `currentValue` das metas SHALL ser maior que zero, refletindo o valor das posições de FIIs

### Requirement: Informar resultado da ação em Meus Ativos e Metas
O sistema SHALL retornar uma mensagem de resultado nas respostas de criação e edição de lançamentos e de metas, para que as telas de Meus Ativos e Metas possam informar ao usuário o resultado da ação executada.

#### Scenario: Criação de lançamento bem-sucedida
- **WHEN** o usuário envia POST `/transactions` com dados válidos
- **THEN** a resposta 201 Created SHALL incluir uma mensagem de sucesso descrevendo o lançamento registrado

#### Scenario: Criação ou atualização de meta bem-sucedida
- **WHEN** o usuário cria ou edita uma meta com dados válidos
- **THEN** a resposta SHALL incluir uma mensagem de sucesso confirmando a operação
