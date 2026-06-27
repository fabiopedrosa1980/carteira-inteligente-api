# Response Cache

## Purpose

Reduzir a latência da API cacheando respostas de leitura em diferentes camadas: cache volátil de curta duração para views dependentes do banco, cache de cotações por ticker com janela curta, e cache persistente de longa duração para o catálogo ticker→ativo da B3 — com invalidação escopada por usuário nas mutações.

## Requirements

### Requirement: Cache de leitura para endpoints de banco

A API SHALL cachear em memória, com TTL curto (~60s), as respostas dos endpoints de leitura que dependem apenas do banco de dados: `GET /transactions` (lista), `GET /allocation`, `GET /goals`, `GET /stocks`, `GET /stocks/:id`, `GET /stocks/:id/dividends` e `GET /dividends/monthly`.

#### Scenario: Segunda leitura é servida do cache

- **WHEN** o mesmo endpoint cacheável é chamado duas vezes dentro da janela de TTL, sem mutações no intervalo
- **THEN** a segunda resposta é servida do cache, com o mesmo corpo, sem reconsultar o banco

#### Scenario: Chave de cache isolada por usuário

- **WHEN** dois usuários autenticados distintos chamam `GET /transactions`
- **THEN** cada um recebe os seus próprios dados, sem compartilhar a entrada de cache do outro

### Requirement: Cache de cotação com janela curta

A API SHALL cachear cotações por ticker com TTL de ~30–60s, compartilhado entre `GET /quote/:ticker` e o enriquecimento de cotação dos endpoints `GET /transactions/acoes|fiis|etfs`.

#### Scenario: Cotação reaproveitada dentro da janela

- **WHEN** a cotação de um mesmo ticker é solicitada novamente dentro do TTL
- **THEN** a resposta usa o valor em cache, sem nova chamada ao provedor externo (Yahoo)

#### Scenario: Cotação não é invalidada por lançamento

- **WHEN** o usuário cria um lançamento ou importa a planilha da B3
- **THEN** o cache de cotação por ticker permanece válido até expirar por TTL

### Requirement: Cache persistente do catálogo ticker→ativo

A API SHALL cachear na própria aplicação Go, com TTL longo (~24h), a resolução de ticker para ativo do catálogo da B3: `GET /assets/:ticker` e `GET /assets/search`.

#### Scenario: Catálogo sobrevive a lançamentos e importações

- **WHEN** o usuário adiciona um lançamento ou importa a planilha da B3
- **THEN** as entradas de cache de `GET /assets/:ticker` e `GET /assets/search` NÃO são invalidadas

#### Scenario: Catálogo renova apenas no refresh

- **WHEN** o catálogo é atualizado via `POST /admin/catalog/refresh` ou pela sincronização periódica
- **THEN** o cache do catálogo passa a refletir os dados atualizados

### Requirement: Invalidação do cache volátil em mutações

A API SHALL invalidar o cache volátil em toda mutação que altere as views cacheadas — incluindo criação, importação, atualização e remoção de lançamentos, e atualização de allocation, goals, stocks e dividends.

#### Scenario: Adicionar lançamento invalida as views do usuário

- **WHEN** o usuário envia `POST /transactions` ou `POST /transactions/import`
- **THEN** as entradas de cache volátil daquele usuário (lista de lançamentos, allocation, etc.) são removidas, e a próxima leitura reflete o novo estado

#### Scenario: Atualização ou remoção também invalida

- **WHEN** o usuário envia `PUT /transactions/:id`, `DELETE /transactions/:id`, `DELETE /transactions` ou `PUT /allocation`
- **THEN** o cache volátil daquele usuário é invalidado, evitando servir dados desatualizados

#### Scenario: Flush escopado por usuário

- **WHEN** um usuário realiza uma mutação que invalida o seu cache volátil
- **THEN** as entradas de cache de outros usuários NÃO são afetadas
