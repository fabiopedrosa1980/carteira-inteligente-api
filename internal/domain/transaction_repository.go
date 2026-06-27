package domain

type TransactionRepository interface {
	Create(t *Transaction) error
	Update(t *Transaction) error
	List(userID, ticker string) ([]*Transaction, error)
	GetByID(userID string, id uint) (*Transaction, error)
	Delete(userID string, id uint) error
	DeleteAll(userID string) error
	// ImportOverwrite substitui atomicamente todos os lançamentos do usuário
	// pela lista informada (apaga tudo e insere os novos numa única transação).
	ImportOverwrite(userID string, txs []*Transaction) error
	GetAcoesPositions(userID string) ([]*AcoesPosition, error)
	GetFiisPositions(userID string) ([]*AcoesPosition, error)
	GetEtfsPositions(userID string) ([]*AcoesPosition, error)
	GetAllPositions(userID string) ([]*AcoesPosition, error)
}
