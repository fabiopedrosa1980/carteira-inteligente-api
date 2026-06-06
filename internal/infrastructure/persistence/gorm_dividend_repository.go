package persistence

import (
	"carteira-inteligente-api/internal/domain"

	"gorm.io/gorm"
)

type GormDividendRepository struct {
	db *gorm.DB
}

func NewGormDividendRepository(db *gorm.DB) domain.DividendRepository {
	return &GormDividendRepository{db: db}
}

func (r *GormDividendRepository) Create(d *domain.Dividend) error {
	return r.db.Create(d).Error
}

func (r *GormDividendRepository) FindByStockID(stockID uint) ([]domain.Dividend, error) {
	var dividends []domain.Dividend
	err := r.db.Where("stock_id = ?", stockID).Order("year, month").Find(&dividends).Error
	return dividends, err
}

func (r *GormDividendRepository) FindByStockIDAndYear(stockID uint, year int) ([]domain.Dividend, error) {
	var dividends []domain.Dividend
	err := r.db.Where("stock_id = ? AND year = ?", stockID, year).Order("month").Find(&dividends).Error
	return dividends, err
}

func (r *GormDividendRepository) FindByYear(year int) ([]domain.Dividend, error) {
	var dividends []domain.Dividend
	err := r.db.Where("year = ?", year).Order("month, stock_id").Find(&dividends).Error
	return dividends, err
}
