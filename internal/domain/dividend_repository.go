package domain

type DividendRepository interface {
	Create(d *Dividend) error
	CreateIfNotExists(d *Dividend) error
	FindByStockID(stockID uint) ([]Dividend, error)
	FindByStockIDAndYear(stockID uint, year int) ([]Dividend, error)
	FindByYear(year int) ([]Dividend, error)
}
