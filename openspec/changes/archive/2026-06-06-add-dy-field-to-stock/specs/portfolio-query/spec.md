## MODIFIED Requirements

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
