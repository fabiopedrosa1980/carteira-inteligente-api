# Portfolio Query

## Purpose

Consulta e filtragem de ações do portfólio: listagem completa, filtro por setor e ordenação por nota ou variação.

## Requirements

### Requirement: Listar todas as ações do portfólio
O sistema SHALL permitir recuperar a lista completa de ações cadastradas no portfólio.

#### Scenario: Listagem com ações cadastradas
- **WHEN** o cliente envia GET `/api/v1/stocks` com o portfólio contendo uma ou mais ações
- **THEN** o sistema SHALL retornar 200 OK com array JSON de todas as ações

#### Scenario: Listagem com portfólio vazio
- **WHEN** o cliente envia GET `/api/v1/stocks` sem nenhuma ação cadastrada
- **THEN** o sistema SHALL retornar 200 OK com array JSON vazio `[]`

### Requirement: Filtrar ações por setor
O sistema SHALL permitir filtrar as ações pelo campo setor via query parameter.

#### Scenario: Filtro por setor com resultados
- **WHEN** o cliente envia GET `/api/v1/stocks?setor=Tecnologia` e existem ações com esse setor
- **THEN** o sistema SHALL retornar 200 OK com apenas as ações do setor informado

#### Scenario: Filtro por setor sem resultados
- **WHEN** o cliente envia GET `/api/v1/stocks?setor=Energia` e não existem ações com esse setor
- **THEN** o sistema SHALL retornar 200 OK com array vazio `[]`

### Requirement: Ordenar ações por nota ou variação
O sistema SHALL permitir ordenar a listagem por nota (descendente), variação do dia ou dy (Dividend Yield) via query parameter `sort`.

#### Scenario: Ordenação por nota
- **WHEN** o cliente envia GET `/api/v1/stocks?sort=nota`
- **THEN** o sistema SHALL retornar as ações ordenadas por nota de forma decrescente

#### Scenario: Ordenação por variação
- **WHEN** o cliente envia GET `/api/v1/stocks?sort=variacao`
- **THEN** o sistema SHALL retornar as ações ordenadas por variação do dia de forma decrescente

#### Scenario: Ordenação por dy
- **WHEN** o cliente envia GET `/api/v1/stocks?sort=dy`
- **THEN** o sistema SHALL retornar as ações ordenadas por `dy` (Dividend Yield) de forma decrescente

#### Scenario: Parâmetro sort inválido
- **WHEN** o cliente envia GET `/api/v1/stocks?sort=invalido`
- **THEN** o sistema SHALL retornar 400 Bad Request com mensagem indicando valores aceitos: `nota`, `variacao`, `dy`
