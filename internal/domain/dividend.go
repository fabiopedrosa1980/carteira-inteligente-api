package domain

import (
	"errors"
	"time"
)

var ErrInvalidDividendType = errors.New("invalid dividend type")

type DividendType string

const (
	DividendTypeDividendo  DividendType = "dividendo"
	DividendTypeJCP        DividendType = "jcp"
	DividendTypeRendimento DividendType = "rendimento"
)

type Dividend struct {
	ID        uint         `gorm:"primaryKey;autoIncrement"                 json:"id"`
	StockID   uint         `gorm:"not null;uniqueIndex:idx_dividend_unique" json:"stock_id"`
	Amount    float64      `gorm:"not null"                                 json:"amount"`
	Month     int          `gorm:"not null"                                 json:"month"`
	Year      int          `gorm:"not null"                                 json:"year"`
	Type      DividendType `gorm:"not null;uniqueIndex:idx_dividend_unique" json:"type"`
	ExDate    string       `gorm:"uniqueIndex:idx_dividend_unique"          json:"ex_date"`
	PayDate   string       `gorm:"uniqueIndex:idx_dividend_unique"          json:"pay_date"`
	CreatedAt time.Time    `json:"created_at"`
}

func (d *Dividend) Validate() error {
	if d.Amount <= 0 {
		return ErrValidation
	}
	if d.Month < 1 || d.Month > 12 {
		return ErrValidation
	}
	if d.Year < 2000 {
		return ErrValidation
	}
	switch d.Type {
	case DividendTypeDividendo, DividendTypeJCP, DividendTypeRendimento:
	default:
		return ErrInvalidDividendType
	}
	return nil
}
