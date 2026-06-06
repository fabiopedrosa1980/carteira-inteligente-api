## Why

Investidores e analistas precisam de uma forma estruturada de gerenciar e acompanhar um portfólio de ações, incluindo dados fundamentais e de mercado de cada ativo. Esta API fornece a base para uma carteira inteligente de ações, permitindo criar, consultar, atualizar e remover posições com dados essenciais de cada ativo.

## What Changes

- Nova API REST para gerenciamento de portfólio de ações com operações CRUD completas
- Modelo de ação com campos: ticker (símbolo), nome, setor, nota (rating), preço atual e variação do dia
- Estrutura de projeto seguindo arquitetura hexagonal (ports & adapters) com camadas bem definidas
- Persistência em banco de dados em memória via GORM com SQLite in-memory
- Servidor HTTP com Gin e organização enterprise-ready (internal/domain, application, infrastructure, adapters)

## Capabilities

### New Capabilities

- `stock-management`: Gerenciamento de ações individuais no portfólio — criação, leitura, atualização e remoção de ativos com ticker, nome, setor, nota, preço atual e variação diária
- `portfolio-query`: Consulta e listagem do portfólio com suporte a filtros por setor e ordenação por nota ou variação

### Modified Capabilities

## Impact

- Projeto novo: nenhum código existente é afetado
- Dependências introduzidas: `gin-gonic/gin`, `gorm.io/gorm`, `gorm.io/driver/sqlite` (in-memory)
- API HTTP exposta na porta 8080 (configurável)
- Estrutura de pacotes: `internal/domain`, `internal/application`, `internal/infrastructure`, `internal/adapters/http`
