package dto

import "carteira-inteligente-api/internal/domain"

type CreateDividendRequest struct {
	Amount  float64            `json:"amount"   binding:"gt=0"`
	Month   int                `json:"month"    binding:"required,min=1,max=12"`
	Year    int                `json:"year"     binding:"required,min=2000"`
	Type    domain.DividendType `json:"type"    binding:"required"`
	ExDate  string             `json:"ex_date"`
	PayDate string             `json:"pay_date"`
}

type DividendResponse struct {
	ID      uint               `json:"id"`
	StockID uint               `json:"stock_id"`
	Amount  float64            `json:"amount"`
	Month   int                `json:"month"`
	Year    int                `json:"year"`
	Type    domain.DividendType `json:"type"`
	ExDate  string             `json:"ex_date"`
	PayDate string             `json:"pay_date"`
}

type MonthSummaryResponse struct {
	Month      int      `json:"month"`
	MonthName  string   `json:"month_name"`
	Tickers    []string `json:"tickers"`
	StockCount int      `json:"stock_count"`
	AvgTotal   float64  `json:"avg_total"`
	AvgYield   float64  `json:"avg_yield"`
}

func DividendFromDomain(d *domain.Dividend) DividendResponse {
	return DividendResponse{
		ID:      d.ID,
		StockID: d.StockID,
		Amount:  d.Amount,
		Month:   d.Month,
		Year:    d.Year,
		Type:    d.Type,
		ExDate:  d.ExDate,
		PayDate: d.PayDate,
	}
}

func DividendListFromDomain(dividends []domain.Dividend) []DividendResponse {
	out := make([]DividendResponse, len(dividends))
	for i := range dividends {
		out[i] = DividendFromDomain(&dividends[i])
	}
	return out
}
