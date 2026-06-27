package dto

import (
	"carteira-inteligente-api/internal/domain"
	"time"
)

// ImportCreated contém a contagem de lançamentos criados por classe na
// importação da planilha de Posição da B3.
type ImportCreated struct {
	Acoes int `json:"Acoes"`
	FIIs  int `json:"FIIs"`
	ETFs  int `json:"ETFs"`
}

// ImportIgnored descreve um ticker que não virou lançamento e o motivo.
type ImportIgnored struct {
	Ticker string `json:"ticker"`
	Reason string `json:"reason,omitempty"`
}

// ImportResultResponse é o resumo devolvido pela importação da planilha.
type ImportResultResponse struct {
	Created ImportCreated   `json:"created"`
	Ignored []ImportIgnored `json:"ignored"`
}

type CreateTransactionRequest struct {
	Ticker    string           `json:"ticker"     binding:"required"`
	AssetType domain.AssetType `json:"asset_type" binding:"required"`
	Quantity  float64          `json:"quantity"   binding:"required,gt=0"`
	Price     float64          `json:"price"      binding:"required,gt=0"`
	Date      string           `json:"date"       binding:"required"`
}

type UpdateTransactionRequest struct {
	AssetType domain.AssetType `json:"asset_type" binding:"required"`
	Quantity  float64          `json:"quantity"   binding:"required,gt=0"`
	Price     float64          `json:"price"      binding:"required,gt=0"`
	Date      string           `json:"date"       binding:"required"`
}

type TransactionResponse struct {
	ID        uint             `json:"id"`
	Ticker    string           `json:"ticker"`
	AssetType domain.AssetType `json:"asset_type"`
	Quantity  float64          `json:"quantity"`
	Price     float64          `json:"price"`
	Date      time.Time        `json:"date"`
	CreatedAt time.Time        `json:"created_at"`
	Message   string           `json:"message,omitempty"`
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

// TransactionWithMessage adiciona uma mensagem de resultado da ação à resposta,
// para que a tela de Meus Ativos informe o usuário sobre o lançamento.
func TransactionWithMessage(t *domain.Transaction, message string) TransactionResponse {
	resp := TransactionFromDomain(t)
	resp.Message = message
	return resp
}

func TransactionListFromDomain(list []*domain.Transaction) []TransactionResponse {
	out := make([]TransactionResponse, len(list))
	for i, t := range list {
		out[i] = TransactionFromDomain(t)
	}
	return out
}
