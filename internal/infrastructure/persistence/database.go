package persistence

import (
	"fmt"
	"log"
	"os"

	"carteira-inteligente-api/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ErrMissingDatabaseURL indica que a app foi iniciada em produção sem a DSN do
// PostgreSQL. Falhar é proposital: o fallback SQLite em memória é efêmero e
// perderia todos os dados a cada restart no Render.
var ErrMissingDatabaseURL = fmt.Errorf("DATABASE_URL é obrigatória em produção (APP_ENV=production)")

// NewDB resolve o driver a partir do ambiente: se DATABASE_URL estiver definida,
// conecta ao PostgreSQL (persistência durável). Caso contrário, em produção a
// app falha (ErrMissingDatabaseURL); fora de produção usa SQLite em memória como
// fallback de conveniência para dev local.
func NewDB() (*gorm.DB, error) {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		log.Println("[db] conectando ao PostgreSQL via DATABASE_URL")
		return NewPostgresDB(dsn)
	}
	if os.Getenv("APP_ENV") == "production" {
		return nil, ErrMissingDatabaseURL
	}
	log.Println("[db] DATABASE_URL ausente; usando SQLite em memória (dev/teste)")
	return NewDBWithDSN("file::memory:?cache=shared")
}

// NewPostgresDB abre a conexão com PostgreSQL e aplica a migração de schema.
func NewPostgresDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

// NewDBWithDSN abre uma conexão SQLite com a DSN informada e aplica a migração
// de schema. Mantida para o dev local e a suíte de testes.
func NewDBWithDSN(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

// migrate executa o AutoMigrate e as instruções de manutenção de schema. As
// instruções de manutenção são best-effort/idempotentes e compatíveis tanto com
// SQLite quanto com PostgreSQL.
func migrate(db *gorm.DB) error {
	// Drop the legacy unique index so AutoMigrate recreates it with the new
	// composition (stock_id, ex_date, pay_date, type). Without this, the old
	// (stock_id, month, year, type) index would silently drop distinct
	// dividends of the same type paid in the same month.
	db.Exec("DROP INDEX IF EXISTS idx_dividend_unique")
	if err := db.AutoMigrate(&domain.Stock{}, &domain.Dividend{}, &domain.Transaction{}, &domain.Goal{}, &domain.AllocationConfig{}, &domain.Asset{}); err != nil {
		return err
	}
	// Remove colunas legadas de metas (type/ticker). A coluna `type` era
	// NOT NULL e quebraria inserts de novas metas sem tipo. Best-effort:
	// erros (coluna inexistente em banco novo) são ignorados. `IF EXISTS`
	// garante boot idempotente no PostgreSQL.
	db.Exec("ALTER TABLE goals DROP COLUMN IF EXISTS type")
	db.Exec("ALTER TABLE goals DROP COLUMN IF EXISTS ticker")
	return nil
}
