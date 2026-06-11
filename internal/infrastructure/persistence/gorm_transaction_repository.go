package persistence

import (
	"carteira-inteligente-api/internal/domain"

	"gorm.io/gorm"
)

type GormTransactionRepository struct {
	db *gorm.DB
}

func NewGormTransactionRepository(db *gorm.DB) *GormTransactionRepository {
	return &GormTransactionRepository{db: db}
}

func (r *GormTransactionRepository) Create(t *domain.Transaction) error {
	return r.db.Create(t).Error
}

func (r *GormTransactionRepository) List(userID, ticker string) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	q := r.db.Where("user_id = ?", userID).Order("date desc")
	if ticker != "" {
		q = q.Where("ticker = ?", ticker)
	}
	return transactions, q.Find(&transactions).Error
}

func (r *GormTransactionRepository) GetByID(userID string, id uint) (*domain.Transaction, error) {
	var t domain.Transaction
	if err := r.db.Where("user_id = ? AND id = ?", userID, id).First(&t).Error; err != nil {
		return nil, domain.ErrTransactionNotFound
	}
	return &t, nil
}

func (r *GormTransactionRepository) Delete(userID string, id uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&domain.Transaction{}, id).Error
}

func (r *GormTransactionRepository) GetPortfolio(userID string) ([]*domain.PortfolioItem, error) {
	type row struct {
		Ticker        string
		AssetType     string
		TotalQuantity float64
		AvgPrice      float64
	}
	var rows []row
	err := r.db.Model(&domain.Transaction{}).
		Select("ticker, asset_type, SUM(quantity) as total_quantity, SUM(quantity*price)/SUM(quantity) as avg_price").
		Where("user_id = ?", userID).
		Group("ticker, asset_type").
		Having("SUM(quantity) > 1").
		Order("ticker").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]*domain.PortfolioItem, len(rows))
	for i, r := range rows {
		items[i] = &domain.PortfolioItem{
			Ticker:        r.Ticker,
			AssetType:     domain.AssetType(r.AssetType),
			TotalQuantity: r.TotalQuantity,
			AvgPrice:      r.AvgPrice,
		}
	}
	return items, nil
}
