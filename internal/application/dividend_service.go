package application

import (
	"carteira-inteligente-api/internal/adapters/http/dto"
	"carteira-inteligente-api/internal/domain"
)

var monthNames = [12]string{
	"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
	"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
}

type DividendUseCase interface {
	CreateDividend(stockID uint, d *domain.Dividend) error
	ListDividendsByStock(stockID uint, year *int) ([]domain.Dividend, error)
	GetMonthlySummary(year int) ([]dto.MonthSummaryResponse, error)
}

type DividendService struct {
	repo      domain.DividendRepository
	stockRepo domain.StockRepository
}

func NewDividendService(repo domain.DividendRepository, stockRepo domain.StockRepository) *DividendService {
	return &DividendService{repo: repo, stockRepo: stockRepo}
}

func (s *DividendService) CreateDividend(stockID uint, d *domain.Dividend) error {
	if _, err := s.stockRepo.FindByID(stockID); err != nil {
		return err
	}
	d.StockID = stockID
	if err := d.Validate(); err != nil {
		return err
	}
	return s.repo.Create(d)
}

func (s *DividendService) ListDividendsByStock(stockID uint, year *int) ([]domain.Dividend, error) {
	if _, err := s.stockRepo.FindByID(stockID); err != nil {
		return nil, err
	}
	if year != nil {
		return s.repo.FindByStockIDAndYear(stockID, *year)
	}
	return s.repo.FindByStockID(stockID)
}

func (s *DividendService) GetMonthlySummary(year int) ([]dto.MonthSummaryResponse, error) {
	dividends, err := s.repo.FindByYear(year)
	if err != nil {
		return nil, err
	}

	type stockEntry struct {
		amount float64
		price  float64
	}
	// month (1-12) → ticker → best entry
	byMonth := make(map[int]map[uint]stockEntry)
	stockIDMap := make(map[uint]string) // stockID → ticker

	for i := range dividends {
		d := &dividends[i]
		if _, ok := byMonth[d.Month]; !ok {
			byMonth[d.Month] = make(map[uint]stockEntry)
		}
		if _, seen := stockIDMap[d.StockID]; !seen {
			stock, err := s.stockRepo.FindByID(d.StockID)
			if err == nil {
				stockIDMap[d.StockID] = stock.Ticker
			}
		}
		prev := byMonth[d.Month][d.StockID]
		byMonth[d.Month][d.StockID] = stockEntry{
			amount: prev.amount + d.Amount,
			price:  byMonth[d.Month][d.StockID].price,
		}
	}

	// attach prices
	for stockID := range stockIDMap {
		stock, err := s.stockRepo.FindByID(stockID)
		if err != nil {
			continue
		}
		for m := range byMonth {
			if entry, ok := byMonth[m][stockID]; ok {
				entry.price = stock.PrecoAtual
				byMonth[m][stockID] = entry
			}
		}
	}

	summaries := make([]dto.MonthSummaryResponse, 12)
	for i := 0; i < 12; i++ {
		month := i + 1
		entries := byMonth[month]
		tickers := make([]string, 0, len(entries))
		var totalAmount, totalYield float64

		for stockID, entry := range entries {
			ticker, ok := stockIDMap[stockID]
			if !ok {
				continue
			}
			tickers = append(tickers, ticker)
			totalAmount += entry.amount
			if entry.price > 0 {
				totalYield += (entry.amount / entry.price) * 100
			}
		}

		count := len(tickers)
		var avgTotal, avgYield float64
		if count > 0 {
			avgTotal = totalAmount / float64(count)
			avgYield = totalYield / float64(count)
		}

		summaries[i] = dto.MonthSummaryResponse{
			Month:      month,
			MonthName:  monthNames[i],
			Tickers:    tickers,
			StockCount: count,
			AvgTotal:   avgTotal,
			AvgYield:   avgYield,
		}
	}

	return summaries, nil
}
