## Context

A persistência atual vive em `internal/infrastructure/persistence/database.go`,
que abre o GORM com SQLite usando a DSN `file::memory:?cache=shared`. O banco é
totalmente em memória: cada restart do processo (incluindo todo deploy no Render
free tier) zera os dados. Toda a aplicação usa repositórios GORM
(`gorm_*_repository.go`), então a troca de banco é localizada na construção da
conexão — os repositórios e serviços são agnósticos ao dialeto.

Pontos relevantes do estado atual:
- `NewDB()` chama `NewDBWithDSN("file::memory:?cache=shared")`.
- `NewDBWithDSN(dsn)` abre `sqlite.Open(dsn)`, executa
  `DROP INDEX IF EXISTS idx_dividend_unique`, roda `AutoMigrate` para
  `Stock`, `Dividend`, `Transaction`, `Goal`, e em seguida executa
  `ALTER TABLE goals DROP COLUMN type` / `DROP COLUMN ticker` (best-effort).
- Os testes (`*_test.go`) chamam `NewDBWithDSN` com DSNs SQLite em memória.
- `Dockerfile` compila com `CGO_ENABLED=1` (exigência do driver `mattn/go-sqlite3`).
- Existe uma instância PostgreSQL provisionada no Render
  (`carteira_digital_bd`, host `dpg-...oregon-postgres.render.com`).

## Goals / Non-Goals

**Goals:**
- Persistir os dados em PostgreSQL quando configurado, sobrevivendo a restarts.
- Selecionar o driver (Postgres vs SQLite) por configuração de ambiente.
- Manter a suíte de testes e o dev local funcionando com SQLite em memória.
- Manter a migração automática de schema funcionando no PostgreSQL.
- Não vazar credenciais para o repositório.

**Non-Goals:**
- Migrar dados existentes do SQLite efêmero (não há fonte durável).
- Introduzir um sistema de migrations versionadas (ex.: golang-migrate);
  continua-se com `AutoMigrate`.
- Alterar o contrato REST, DTOs, serviços ou repositórios.
- Configurar pool de conexões avançado/observabilidade (fora de escopo).

## Decisions

### Decisão 1: Seleção de dialector por `DATABASE_URL`
`NewDB()` passa a ler `os.Getenv("DATABASE_URL")`. Se não vazia, abre Postgres;
caso contrário, mantém o fallback SQLite em memória. A função interna de migração
passa a receber o `*gorm.DB` já aberto, separando "abrir conexão" de
"migrar/manter schema".

- **Por quê**: localiza a mudança em um único ponto, preserva testes (que
  injetam DSN SQLite via `NewDBWithDSN`) e segue o padrão idiomático do Render
  (`DATABASE_URL`).
- **Alternativas consideradas**: (a) variável `DB_DRIVER` explícita — mais
  verboso e redundante com a presença da URL; (b) trocar SQLite por Postgres
  também nos testes — exigiria um Postgres em CI, aumentando atrito sem ganho.

### Decisão 2: Estrutura das funções de conexão
Refatorar para:
- `NewDB()` — resolve o driver a partir do ambiente e delega.
- `NewPostgresDB(dsn)` — `gorm.Open(postgres.Open(dsn))` + migração.
- `NewDBWithDSN(dsn)` — permanece SQLite (assinatura inalterada para os testes).
- `migrate(db)` (privada) — `AutoMigrate` + manutenção de schema, compartilhada
  por ambos os caminhos.

- **Por quê**: evita duplicar a lógica de `AutoMigrate`/manutenção e mantém a
  superfície usada pelos testes intacta.

### Decisão 3: SQL de manutenção compatível com PostgreSQL
- `DROP INDEX IF EXISTS idx_dividend_unique` já é válido em ambos os dialetos.
- Os drops de coluna passam a usar `ALTER TABLE goals DROP COLUMN IF EXISTS type`
  e `... IF EXISTS ticker` (PostgreSQL suporta `IF EXISTS`). No SQLite, esses
  statements continuam best-effort (erros ignorados como hoje), então a forma com
  `IF EXISTS` no Postgres evita poluir logs/abortar conexão.

- **Por quê**: garante boot idempotente no Postgres sem erros em banco já limpo.

### Decisão 4: Dependência e build
- Adicionar `gorm.io/driver/postgres` (Go puro, sem CGO).
- O `Dockerfile` mantém `CGO_ENABLED=1` enquanto o driver SQLite permanecer
  importado (fallback/local). Como o binário ainda linka `mattn/go-sqlite3`,
  remover o CGO quebraria o build; portanto **não** alteramos o CGO nesta etapa.

- **Alternativa**: trocar para um SQLite puro-Go (ex.: `glebarez/sqlite`) e
  desligar CGO — adiado para não ampliar o escopo.

### Decisão 5: DSN e segurança
A DSN é montada no formato URL aceito pelo driver Postgres:
`postgresql://USER:PASSWORD@HOST:5432/DBNAME?sslmode=require`.
Para o Render externo é necessário `sslmode=require`. O valor vai em
`DATABASE_URL` no ambiente do Render (dashboard/secret). No `render.yaml`,
declara-se a env var com `sync: false` (sem valor commitado).

- **Por quê**: Render exige TLS para conexões externas e credenciais não devem
  ir ao versionamento.

## Risks / Trade-offs

- **Credenciais expostas no histórico do chat/issue** → usar somente via
  `DATABASE_URL` no Render; recomendar rotação da senha do banco após a
  configuração inicial.
- **`AutoMigrate` em produção pode divergir do schema desejado** → escopo atual
  é greenfield (banco vazio); mapeamentos GORM já validados em SQLite. Risco
  baixo, mas observar tipos específicos (ex.: índices únicos compostos de
  dividendos) no primeiro boot Postgres.
- **Diferenças de dialeto (tipos/índices) entre SQLite e Postgres** → cobertas
  pelo `AutoMigrate`; validar manualmente o índice único de `Dividend`
  (`stock_id, ex_date, pay_date, type`) após o primeiro deploy.
- **Free tier do Render pode pausar/limitar conexões** → fora de escopo; pool
  padrão do GORM é suficiente para o volume atual.
- **Falha de conexão na inicialização** → o `main` já trata erro de `NewDB` com
  `log.Fatalf`, então conexão inválida impede subir com persistência quebrada.

## Migration Plan

1. Adicionar `gorm.io/driver/postgres` (`go get`), atualizar `go.mod`/`go.sum`.
2. Refatorar `database.go` (decisões 1–3).
3. `go build ./...` e `go test ./...` (testes seguem em SQLite).
4. Provisionar/confirmar a env var `DATABASE_URL` no Render com
   `sslmode=require`; declarar a var em `render.yaml` (`sync: false`).
5. Deploy. Verificar nos logs a criação do schema e validar persistência
   (criar registro → restart → conferir que persiste).
6. **Rollback**: remover/limpar `DATABASE_URL` no Render faz o serviço voltar ao
   SQLite em memória (sem durabilidade), restaurando o comportamento anterior
   sem novo deploy de código.

## Open Questions

- Rotacionar a senha do banco após a configuração? (Recomendado, dado que foi
  compartilhada em texto claro.)
- Usar a connection string interna do Render (host `...-a` sem sufixo de domínio
  público) quando API e banco estiverem na mesma região, evitando tráfego
  externo e dispensando `sslmode=require`? (Otimização opcional.)
