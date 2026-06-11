package domain

import (
	"errors"
	"time"
)

var ErrTransactionNotFound = errors.New("transaction not found")

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
	CreatedAt time.Time                                  `json:"created_at"`
}

type AcoesPosition struct {
	Ticker        string
	TotalQuantity float64
	AvgPrice      float64
}

type AcaoItem struct {
	Ticker        string  `json:"ticker"`
	Name          string  `json:"name"`
	TotalQuantity float64 `json:"total_quantity"`
	AvgPrice      float64 `json:"avg_price"`
	CurrentPrice  float64 `json:"current_price"`
	ChangePercent float64 `json:"change_percent"`
}
