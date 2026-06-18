package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrTransactionNotFound = errors.New("transaction not found")

// NormalizeTicker padroniza o ticker removendo espaços nas extremidades e
// convertendo para caixa alta, de modo que variações de digitação do mesmo
// ativo (ex.: " petr4 ", "PETR4") sejam tratadas como um único ticker.
func NormalizeTicker(ticker string) string {
	return strings.ToUpper(strings.TrimSpace(ticker))
}

type AssetType string

const (
	AssetTypeAcoes AssetType = "Acoes"
	AssetTypeFIIs  AssetType = "FIIs"
	AssetTypeETFs  AssetType = "ETFs"
)

type Transaction struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"not null;index"           json:"user_id"`
	Ticker    string    `gorm:"not null;index"           json:"ticker"`
	AssetType AssetType `gorm:"not null"                 json:"asset_type"`
	Quantity  float64   `gorm:"not null"                 json:"quantity"`
	Price     float64   `gorm:"not null"                 json:"price"`
	Date      time.Time `gorm:"not null"                 json:"date"`
	CreatedAt time.Time `json:"created_at"`
}

type AcoesPosition struct {
	Ticker           string
	TotalQuantity    float64
	AvgPrice         float64
	TransactionCount int
}

type AcaoItem struct {
	Ticker           string      `json:"ticker"`
	Name             string      `json:"name"`
	TotalQuantity    float64     `json:"total_quantity"`
	AvgPrice         float64     `json:"avg_price"`
	CurrentPrice     float64     `json:"current_price"`
	ChangePercent    float64     `json:"change_percent"`
	DividendYield    float64     `json:"dividend_yield"`
	Nota             float64     `json:"nota"`
	HistoryReady     bool        `json:"history_ready"`
	StockID          uint        `json:"stock_id"`
	TransactionCount int         `json:"transaction_count"`
	Indicators       []Indicator `json:"indicators,omitempty"`
	CompanyInfo      []Indicator `json:"company_info,omitempty"`
}

// Indicator é um indicador fundamentalista no formato rótulo/valor, com o valor
// já formatado como exibido na fonte (ex.: "12,34", "8,50%").
type Indicator struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
