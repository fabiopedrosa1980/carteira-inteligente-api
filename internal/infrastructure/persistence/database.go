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
	if err := db.AutoMigrate(&domain.Stock{}, &domain.Dividend{}); err != nil {
		return nil, err
	}
	return db, nil
}
