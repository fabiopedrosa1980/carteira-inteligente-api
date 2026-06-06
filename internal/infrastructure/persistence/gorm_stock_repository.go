package persistence

import (
	"errors"
	"strings"

	"carteira-inteligente-api/internal/domain"

	"gorm.io/gorm"
)

type GormStockRepository struct {
	db *gorm.DB
}

func NewGormStockRepository(db *gorm.DB) domain.StockRepository {
	return &GormStockRepository{db: db}
}

func (r *GormStockRepository) Create(stock *domain.Stock) error {
	result := r.db.Create(stock)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "UNIQUE constraint failed") {
			return domain.ErrDuplicate
		}
		return result.Error
	}
	return nil
}

func (r *GormStockRepository) FindByID(id uint) (*domain.Stock, error) {
	var stock domain.Stock
	result := r.db.First(&stock, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	return &stock, result.Error
}

func (r *GormStockRepository) FindAll(query domain.StockQuery) ([]domain.Stock, error) {
	var stocks []domain.Stock
	db := r.db.Model(&domain.Stock{})

	if query.Sector != "" {
		db = db.Where("sector = ?", query.Sector)
	}

	switch query.Sort {
	case "score":
		db = db.Order("score DESC")
	case "daily_change":
		db = db.Order("daily_change DESC")
	case "dy":
		db = db.Order("dy DESC")
	}

	if err := db.Find(&stocks).Error; err != nil {
		return nil, err
	}
	return stocks, nil
}

func (r *GormStockRepository) Update(stock *domain.Stock) error {
	result := r.db.Save(stock)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *GormStockRepository) Delete(id uint) error {
	result := r.db.Delete(&domain.Stock{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
