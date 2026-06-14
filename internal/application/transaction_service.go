package application

import "carteira-inteligente-api/internal/domain"

type TransactionUseCase interface {
	Create(t *domain.Transaction) error
	Update(t *domain.Transaction) error
	List(userID, ticker string) ([]*domain.Transaction, error)
	GetByID(userID string, id uint) (*domain.Transaction, error)
	Delete(userID string, id uint) error
	GetAcoesPositions(userID string) ([]*domain.AcoesPosition, error)
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

func (s *TransactionService) Update(t *domain.Transaction) error {
	if _, err := s.repo.GetByID(t.UserID, t.ID); err != nil {
		return domain.ErrTransactionNotFound
	}
	return s.repo.Update(t)
}

func (s *TransactionService) List(userID, ticker string) ([]*domain.Transaction, error) {
	return s.repo.List(userID, ticker)
}

func (s *TransactionService) GetByID(userID string, id uint) (*domain.Transaction, error) {
	return s.repo.GetByID(userID, id)
}

func (s *TransactionService) Delete(userID string, id uint) error {
	if _, err := s.repo.GetByID(userID, id); err != nil {
		return domain.ErrTransactionNotFound
	}
	return s.repo.Delete(userID, id)
}

func (s *TransactionService) GetAcoesPositions(userID string) ([]*domain.AcoesPosition, error) {
	return s.repo.GetAcoesPositions(userID)
}
