package application

import "carteira-inteligente-api/internal/domain"

type TransactionUseCase interface {
	Create(t *domain.Transaction) error
	List(ticker string) ([]*domain.Transaction, error)
	GetByID(id uint) (*domain.Transaction, error)
	Delete(id uint) error
}

type TransactionService struct {
	repo domain.TransactionRepository
}

func NewTransactionService(repo domain.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Create(t *domain.Transaction) error {
	return s.repo.Create(t)
}

func (s *TransactionService) List(ticker string) ([]*domain.Transaction, error) {
	return s.repo.List(ticker)
}

func (s *TransactionService) GetByID(id uint) (*domain.Transaction, error) {
	return s.repo.GetByID(id)
}

func (s *TransactionService) Delete(id uint) error {
	if _, err := s.repo.GetByID(id); err != nil {
		return domain.ErrTransactionNotFound
	}
	return s.repo.Delete(id)
}
