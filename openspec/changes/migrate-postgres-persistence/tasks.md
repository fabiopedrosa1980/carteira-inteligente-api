## 1. Dependência

- [x] 1.1 Rodar `go get gorm.io/driver/postgres` e confirmar atualização de `go.mod`/`go.sum`
- [x] 1.2 Rodar `go mod tidy` e garantir que o módulo continua compilando (`go build ./...`)

## 2. Camada de persistência

- [x] 2.1 Extrair a lógica de `AutoMigrate` + manutenção de schema para uma função privada `migrate(db *gorm.DB) error` em `internal/infrastructure/persistence/database.go`
- [x] 2.2 Tornar os drops de coluna compatíveis com Postgres: `ALTER TABLE goals DROP COLUMN IF EXISTS type` e `ALTER TABLE goals DROP COLUMN IF EXISTS ticker`
- [x] 2.3 Adicionar `NewPostgresDB(dsn string) (*gorm.DB, error)` usando `postgres.Open(dsn)` + `migrate`
- [x] 2.4 Manter `NewDBWithDSN(dsn)` (SQLite) chamando `migrate`, com assinatura inalterada para os testes
- [x] 2.5 Alterar `NewDB()` para ler `os.Getenv("DATABASE_URL")`: se definida, usar `NewPostgresDB`; caso contrário, fallback SQLite em memória

## 3. Build e testes

- [x] 3.1 `go build ./...` e `go vet ./...` sem erros
- [x] 3.2 `go test ./...` passando (testes seguem em SQLite via `NewDBWithDSN`)
- [ ] 3.3 (Opcional/local) Subir com `DATABASE_URL` apontando para o Postgres e validar criação de schema + persistência após restart

## 4. Infra e configuração

- [x] 4.1 Adicionar a env var `DATABASE_URL` em `render.yaml` com `sync: false` (sem valor commitado)
- [ ] 4.2 Definir `DATABASE_URL` no dashboard do Render no formato `postgresql://USER:PASSWORD@HOST:5432/carteira_digital_bd?sslmode=require`
- [x] 4.3 Confirmar que nenhuma credencial foi commitada no repositório

## 5. Deploy e verificação

- [ ] 5.1 Deploy no Render e conferir nos logs a criação do schema (`AutoMigrate`)
- [ ] 5.2 Validar persistência end-to-end: criar registro via API → restart do serviço → confirmar que o dado persiste
- [ ] 5.3 Verificar o índice único composto de `Dividend` (`stock_id, ex_date, pay_date, type`) no Postgres
- [ ] 5.4 (Recomendado) Rotacionar a senha do banco e atualizar `DATABASE_URL` no Render
