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

func (r *GormTransactionRepository) List(ticker string) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction
	q := r.db.Order("date desc")
	if ticker != "" {
		q = q.Where("ticker = ?", ticker)
	}
	return transactions, q.Find(&transactions).Error
}

func (r *GormTransactionRepository) GetByID(id uint) (*domain.Transaction, error) {
	var t domain.Transaction
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, domain.ErrTransactionNotFound
	}
	return &t, nil
}

func (r *GormTransactionRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Transaction{}, id).Error
}
