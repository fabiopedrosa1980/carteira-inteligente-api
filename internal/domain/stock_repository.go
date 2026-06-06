package domain

type StockQuery struct {
	Setor string
	Sort  string
}

type StockRepository interface {
	Create(stock *Stock) error
	FindByID(id uint) (*Stock, error)
	FindAll(query StockQuery) ([]Stock, error)
	Update(stock *Stock) error
	Delete(id uint) error
}
