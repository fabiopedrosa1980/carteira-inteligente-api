package domain

type TransactionRepository interface {
	Create(t *Transaction) error
	Update(t *Transaction) error
	List(userID, ticker string) ([]*Transaction, error)
	GetByID(userID string, id uint) (*Transaction, error)
	Delete(userID string, id uint) error
	GetAcoesPositions(userID string) ([]*AcoesPosition, error)
}
