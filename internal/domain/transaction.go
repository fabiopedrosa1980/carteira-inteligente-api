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
	Ticker    string    `gorm:"not null;index"           json:"ticker"`
	AssetType AssetType `gorm:"not null"                 json:"asset_type"`
	Quantity  float64   `gorm:"not null"                 json:"quantity"`
	Price     float64   `gorm:"not null"                 json:"price"`
	Date      time.Time `gorm:"not null"                 json:"date"`
	CreatedAt time.Time                                  `json:"created_at"`
}
