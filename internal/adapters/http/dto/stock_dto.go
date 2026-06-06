package dto

import (
	"carteira-inteligente-api/internal/domain"
)

type CreateStockRequest struct {
	Ticker       string  `json:"ticker"         binding:"required"`
	Name         string  `json:"name"           binding:"required"`
	Sector       string  `json:"sector"`
	Score        float64 `json:"score"          binding:"min=0,max=10"`
	CurrentPrice float64 `json:"current_price"  binding:"gt=0"`
	DailyChange  float64 `json:"daily_change"`
	DY           float64 `json:"dy"`
}

type UpdateStockRequest struct {
	Ticker       string  `json:"ticker"         binding:"required"`
	Name         string  `json:"name"           binding:"required"`
	Sector       string  `json:"sector"`
	Score        float64 `json:"score"          binding:"min=0,max=10"`
	CurrentPrice float64 `json:"current_price"  binding:"gt=0"`
	DailyChange  float64 `json:"daily_change"`
	DY           float64 `json:"dy"`
}

type StockResponse struct {
	ID           uint    `json:"id"`
	Ticker       string  `json:"ticker"`
	Name         string  `json:"name"`
	Sector       string  `json:"sector"`
	Score        float64 `json:"score"`
	CurrentPrice float64 `json:"current_price"`
	DailyChange  float64 `json:"daily_change"`
	DY           float64 `json:"dy"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func FromDomain(s *domain.Stock) StockResponse {
	return StockResponse{
		ID:           s.ID,
		Ticker:       s.Ticker,
		Name:         s.Name,
		Sector:       s.Sector,
		Score:        s.Score,
		CurrentPrice: s.CurrentPrice,
		DailyChange:  s.DailyChange,
		DY:           s.DY,
		CreatedAt:    s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    s.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func FromDomainList(stocks []domain.Stock) []StockResponse {
	responses := make([]StockResponse, len(stocks))
	for i := range stocks {
		responses[i] = FromDomain(&stocks[i])
	}
	return responses
}
