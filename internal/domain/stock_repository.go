package domain

type StockQuery struct {
	Sector string
	Sort   string
}

type StockRepository interface {
	Create(stock *Stock) error
	FindByID(id uint) (*Stock, error)
	FindAll(query StockQuery) ([]Stock, error)
	Update(stock *Stock) error
	Delete(id uint) error
	UpdateHistoryReady(id uint, ready bool) error
}
