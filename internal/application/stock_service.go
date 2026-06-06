package application

import "carteira-inteligente-api/internal/domain"

type StockUseCase interface {
	CreateStock(stock *domain.Stock) error
	GetStockByID(id uint) (*domain.Stock, error)
	ListStocks(query domain.StockQuery) ([]domain.Stock, error)
	UpdateStock(id uint, updated *domain.Stock) (*domain.Stock, error)
	DeleteStock(id uint) error
}

type StockService struct {
	repo domain.StockRepository
}

func NewStockService(repo domain.StockRepository) *StockService {
	return &StockService{repo: repo}
}

func (s *StockService) CreateStock(stock *domain.Stock) error {
	if err := stock.Validate(); err != nil {
		return err
	}
	return s.repo.Create(stock)
}

func (s *StockService) GetStockByID(id uint) (*domain.Stock, error) {
	return s.repo.FindByID(id)
}

func (s *StockService) ListStocks(query domain.StockQuery) ([]domain.Stock, error) {
	return s.repo.FindAll(query)
}

func (s *StockService) UpdateStock(id uint, updated *domain.Stock) (*domain.Stock, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	existing.Ticker = updated.Ticker
	existing.Name = updated.Name
	existing.Sector = updated.Sector
	existing.Score = updated.Score
	existing.CurrentPrice = updated.CurrentPrice
	existing.DailyChange = updated.DailyChange
	existing.DY = updated.DY

	if err := existing.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *StockService) DeleteStock(id uint) error {
	return s.repo.Delete(id)
}
