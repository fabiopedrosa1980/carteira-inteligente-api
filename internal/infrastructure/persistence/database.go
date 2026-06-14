package persistence

import (
	"carteira-inteligente-api/internal/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDB() (*gorm.DB, error) {
	return NewDBWithDSN("file::memory:?cache=shared")
}

func NewDBWithDSN(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// Drop the legacy unique index so AutoMigrate recreates it with the new
	// composition (stock_id, ex_date, pay_date, type). Without this, SQLite
	// keeps the old (stock_id, month, year, type) index and silently drops
	// distinct dividends of the same type paid in the same month.
	db.Exec("DROP INDEX IF EXISTS idx_dividend_unique")
	if err := db.AutoMigrate(&domain.Stock{}, &domain.Dividend{}, &domain.Transaction{}, &domain.Goal{}); err != nil {
		return nil, err
	}
	// Remove colunas legadas de metas (type/ticker). A coluna `type` era
	// NOT NULL e quebraria inserts de novas metas sem tipo. Best-effort:
	// erros (coluna inexistente em banco novo) são ignorados.
	db.Exec("ALTER TABLE goals DROP COLUMN type")
	db.Exec("ALTER TABLE goals DROP COLUMN ticker")
	return db, nil
}
