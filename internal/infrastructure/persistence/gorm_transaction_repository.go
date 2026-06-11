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
