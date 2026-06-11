package dto

import (
	"carteira-inteligente-api/internal/domain"
	"time"
)

type CreateTransactionRequest struct {
	Ticker    string            `json:"ticker"     binding:"required"`
	AssetType domain.AssetType  `json:"asset_type" binding:"required"`
	Quantity  float64           `json:"quantity"   binding:"required,gt=0"`
	Price     float64           `json:"price"      binding:"required,gt=0"`
	Date      string            `json:"date"       binding:"required"`
}

type TransactionResponse struct {
	ID        uint              `json:"id"`
	Ticker    string            `json:"ticker"`
	AssetType domain.AssetType  `json:"asset_type"`
	Quantity  float64           `json:"quantity"`
	Price     float64           `json:"price"`
	Date      time.Time         `json:"date"`
	CreatedAt time.Time         `json:"created_at"`
}

func TransactionFromDomain(t *domain.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:        t.ID,
		Ticker:    t.Ticker,
		AssetType: t.AssetType,
		Quantity:  t.Quantity,
		Price:     t.Price,
		Date:      t.Date,
		CreatedAt: t.CreatedAt,
	}
}

func TransactionListFromDomain(list []*domain.Transaction) []TransactionResponse {
	out := make([]TransactionResponse, len(list))
	for i, t := range list {
		out[i] = TransactionFromDomain(t)
	}
	return out
}

type PortfolioItemResponse struct {
	Ticker        string           `json:"ticker"`
	AssetType     domain.AssetType `json:"asset_type"`
	TotalQuantity float64          `json:"total_quantity"`
	AvgPrice      float64          `json:"avg_price"`
}

func PortfolioItemFromDomain(p *domain.PortfolioItem) PortfolioItemResponse {
	return PortfolioItemResponse{
		Ticker:        p.Ticker,
		AssetType:     p.AssetType,
		TotalQuantity: p.TotalQuantity,
		AvgPrice:      p.AvgPrice,
	}
}

func PortfolioListFromDomain(list []*domain.PortfolioItem) []PortfolioItemResponse {
	out := make([]PortfolioItemResponse, len(list))
	for i, p := range list {
		out[i] = PortfolioItemFromDomain(p)
	}
	return out
}
