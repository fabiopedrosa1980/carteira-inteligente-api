package dto

import (
	"carteira-inteligente-api/internal/domain"
)

type CreateStockRequest struct {
	Ticker       string  `json:"ticker"        binding:"required"`
	Nome         string  `json:"nome"          binding:"required"`
	Setor        string  `json:"setor"`
	Nota         float64 `json:"nota"          binding:"min=0,max=10"`
	PrecoAtual   float64 `json:"preco_atual"   binding:"gt=0"`
	VariacaoHoje float64 `json:"variacao_hoje"`
	DY           float64 `json:"dy"`
}

type UpdateStockRequest struct {
	Ticker       string  `json:"ticker"        binding:"required"`
	Nome         string  `json:"nome"          binding:"required"`
	Setor        string  `json:"setor"`
	Nota         float64 `json:"nota"          binding:"min=0,max=10"`
	PrecoAtual   float64 `json:"preco_atual"   binding:"gt=0"`
	VariacaoHoje float64 `json:"variacao_hoje"`
	DY           float64 `json:"dy"`
}

type StockResponse struct {
	ID           uint    `json:"id"`
	Ticker       string  `json:"ticker"`
	Nome         string  `json:"nome"`
	Setor        string  `json:"setor"`
	Nota         float64 `json:"nota"`
	PrecoAtual   float64 `json:"preco_atual"`
	VariacaoHoje float64 `json:"variacao_hoje"`
	DY           float64 `json:"dy"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func FromDomain(s *domain.Stock) StockResponse {
	return StockResponse{
		ID:           s.ID,
		Ticker:       s.Ticker,
		Nome:         s.Nome,
		Setor:        s.Setor,
		Nota:         s.Nota,
		PrecoAtual:   s.PrecoAtual,
		VariacaoHoje: s.VariacaoHoje,
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
