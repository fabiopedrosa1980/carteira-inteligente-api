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

func (r *GormTransactionRepository) Update(t *domain.Transaction) error {
	return r.db.Model(&domain.Transaction{}).
		Where("user_id = ? AND id = ?", t.UserID, t.ID).
		Updates(map[string]interface{}{
			"asset_type": t.AssetType,
			"quantity":   t.Quantity,
			"price":      t.Price,
			"date":       t.Date,
		}).Error
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

func (r *GormTransactionRepository) GetAcoesPositions(userID string) ([]*domain.AcoesPosition, error) {
	type row struct {
		Ticker        string
		TotalQuantity float64
		AvgPrice      float64
	}
	var rows []row
	err := r.db.Model(&domain.Transaction{}).
		Select("ticker, SUM(quantity) as total_quantity, SUM(quantity*price)/SUM(quantity) as avg_price").
		Where("user_id = ? AND asset_type = ?", userID, domain.AssetTypeAcoes).
		Group("ticker").
		Having("SUM(quantity) > 0").
		Order("ticker").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]*domain.AcoesPosition, len(rows))
	for i, r := range rows {
		result[i] = &domain.AcoesPosition{
			Ticker:        r.Ticker,
			TotalQuantity: r.TotalQuantity,
			AvgPrice:      r.AvgPrice,
		}
	}
	return result, nil
}
