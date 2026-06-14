package application_test

import (
	"errors"
	"testing"

	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"
)

type mockRepo struct {
	stocks map[uint]*domain.Stock
	nextID uint
}

func newMockRepo() *mockRepo {
	return &mockRepo{stocks: make(map[uint]*domain.Stock), nextID: 1}
}

func (m *mockRepo) Create(stock *domain.Stock) error {
	for _, s := range m.stocks {
		if s.Ticker == stock.Ticker {
			return domain.ErrDuplicate
		}
	}
	stock.ID = m.nextID
	m.nextID++
	copy := *stock
	m.stocks[stock.ID] = &copy
	return nil
}

func (m *mockRepo) FindByID(id uint) (*domain.Stock, error) {
	s, ok := m.stocks[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copy := *s
	return &copy, nil
}

func (m *mockRepo) FindAll(query domain.StockQuery) ([]domain.Stock, error) {
	var result []domain.Stock
	for _, s := range m.stocks {
		if query.Sector == "" || s.Sector == query.Sector {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (m *mockRepo) Update(stock *domain.Stock) error {
	if _, ok := m.stocks[stock.ID]; !ok {
		return domain.ErrNotFound
	}
	copy := *stock
	m.stocks[stock.ID] = &copy
	return nil
}

func (m *mockRepo) Delete(id uint) error {
	if _, ok := m.stocks[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.stocks, id)
	return nil
}

func (m *mockRepo) UpdateHistoryReady(id uint, ready bool) error {
	s, ok := m.stocks[id]
	if !ok {
		return domain.ErrNotFound
	}
	s.HistoryReady = ready
	return nil
}

func TestCreateStock_Success(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	stock := &domain.Stock{Ticker: "PETR4", Name: "Petrobras", Sector: "Energia", Score: 8.0, CurrentPrice: 35.50}
	if err := svc.CreateStock(stock); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if stock.ID == 0 {
		t.Fatal("expected ID to be set")
	}
}

func TestCreateStock_Duplicate(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	svc.CreateStock(&domain.Stock{Ticker: "PETR4", Name: "Petrobras", Score: 8.0, CurrentPrice: 35.50})
	err := svc.CreateStock(&domain.Stock{Ticker: "PETR4", Name: "Petrobras 2", Score: 7.0, CurrentPrice: 30.00})
	if !errors.Is(err, domain.ErrDuplicate) {
		t.Fatalf("expected ErrDuplicate, got %v", err)
	}
}

func TestCreateStock_InvalidNota(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	err := svc.CreateStock(&domain.Stock{Ticker: "PETR4", Name: "Petrobras", Score: 11.0, CurrentPrice: 35.50})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestCreateStock_InvalidPreco(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	err := svc.CreateStock(&domain.Stock{Ticker: "PETR4", Name: "Petrobras", Score: 8.0, CurrentPrice: -1.0})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestCreateStock_MissingTicker(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	err := svc.CreateStock(&domain.Stock{Name: "Petrobras", Score: 8.0, CurrentPrice: 35.50})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestGetStockByID_NotFound(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	_, err := svc.GetStockByID(999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestListStocks_FilterBySetor(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	svc.CreateStock(&domain.Stock{Ticker: "PETR4", Name: "Petrobras", Sector: "Energia", Score: 8.0, CurrentPrice: 35.50})
	svc.CreateStock(&domain.Stock{Ticker: "VALE3", Name: "Vale", Sector: "Mineração", Score: 7.0, CurrentPrice: 65.00})

	stocks, err := svc.ListStocks(domain.StockQuery{Sector: "Energia"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stocks) != 1 || stocks[0].Ticker != "PETR4" {
		t.Fatalf("expected 1 stock PETR4, got %v", stocks)
	}
}

func TestListStocks_EmptySetor(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	stocks, err := svc.ListStocks(domain.StockQuery{Sector: "Inexistente"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stocks) != 0 {
		t.Fatalf("expected empty slice, got %v", stocks)
	}
}

func TestUpdateStock_NotFound(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	_, err := svc.UpdateStock(999, &domain.Stock{Ticker: "X", Name: "Y", Score: 5, CurrentPrice: 10})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateStock_DY_Propagated(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	stock := &domain.Stock{Ticker: "PETR4", Name: "Petrobras", Score: 8.0, CurrentPrice: 35.50, DY: 3.0}
	if err := svc.CreateStock(stock); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated, err := svc.UpdateStock(stock.ID, &domain.Stock{Ticker: "PETR4", Name: "Petrobras", Score: 8.0, CurrentPrice: 35.50, DY: 7.5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.DY != 7.5 {
		t.Fatalf("expected DY=7.5, got %v", updated.DY)
	}
}

func TestDeleteStock_NotFound(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	err := svc.DeleteStock(999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteStock_Success(t *testing.T) {
	svc := application.NewStockService(newMockRepo())
	stock := &domain.Stock{Ticker: "PETR4", Name: "Petrobras", Score: 8.0, CurrentPrice: 35.50}
	svc.CreateStock(stock)
	if err := svc.DeleteStock(stock.ID); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
