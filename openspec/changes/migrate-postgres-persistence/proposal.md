## Why

Hoje a API armazena todos os dados em um banco SQLite efêmero, aberto com a DSN
`file::memory:?cache=shared`. Cada reinício ou novo deploy no Render apaga toda a
carteira (ações, dividendos, transações, metas), de modo que nenhum dado é
durável. Migrar a persistência para a instância PostgreSQL gerenciada já
provisionada no Render faz os dados sobreviverem a reinícios e habilita um
deploy de produção real (inclusive multi-instância).

## What Changes

- Adicionar a dependência `gorm.io/driver/postgres` e selecionar o dialector do
  GORM na inicialização com base na variável de ambiente `DATABASE_URL`.
- Quando `DATABASE_URL` estiver definida, conectar ao PostgreSQL; caso contrário,
  usar a DSN SQLite em memória atual como fallback (preserva o dev local e a
  suíte de testes, que chamam `NewDBWithDSN` diretamente com SQLite).
- Tornar as instruções de manutenção de schema executadas após o `AutoMigrate`
  (drop do índice legado e drop das colunas legadas em `goals`) compatíveis com
  PostgreSQL.
- Configurar a conexão PostgreSQL no Render via variável de ambiente (com
  `sslmode=require`) em vez de fixar credenciais no repositório. **BREAKING**
  para operação: o serviço passa a exigir `DATABASE_URL` no ambiente do Render
  para usar o Postgres.
- Documentar os dados de conexão e os passos de provisionamento; referenciar a
  nova variável em `render.yaml` (valor fornecido via dashboard/secret do Render,
  nunca commitado).

## Capabilities

### New Capabilities
- `data-persistence`: Armazenamento durável de todas as entidades de domínio
  (ações, dividendos, transações, metas) em PostgreSQL, sobrevivendo a reinícios
  e redeploys, com seleção de driver dirigida por configuração e migração
  automática de schema na inicialização.

### Modified Capabilities
<!-- Sem mudança de comportamento em nível de spec nas capabilities existentes (portfolio-query / stock-management); o contrato REST permanece inalterado. -->

## Impact

- **Código**: `internal/infrastructure/persistence/database.go` (seleção de
  driver, SQL de migração compatível com o dialeto), `cmd/api/main.go` (sem
  mudança na ligação além do `NewDB`).
- **Dependências**: adiciona `gorm.io/driver/postgres` (Go puro); o driver SQLite
  é mantido para testes e fallback local.
- **Infra**: `render.yaml` ganha a variável `DATABASE_URL`; o `Dockerfile` pode
  dispensar o CGO no caminho Postgres (mantido como está se o fallback SQLite
  ainda precisar compilar). A instância PostgreSQL no Render
  (`carteira_digital_bd`) passa a ser a fonte da verdade.
- **Dados**: começa vazio no Postgres; o AutoMigrate cria o schema no primeiro
  boot. Não há migração automática de dados do SQLite efêmero (não existe fonte
  durável para migrar).
- **Segurança**: as credenciais do banco passam para configuração via ambiente;
  não devem ser commitadas no repositório.
