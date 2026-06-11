package domain

type TransactionRepository interface {
	Create(t *Transaction) error
	List(ticker string) ([]*Transaction, error)
	GetByID(id uint) (*Transaction, error)
	Delete(id uint) error
}
