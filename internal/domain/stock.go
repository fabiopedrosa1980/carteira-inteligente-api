package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound   = errors.New("stock not found")
	ErrDuplicate  = errors.New("ticker already exists")
	ErrValidation = errors.New("validation error")
)

type Stock struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Ticker       string    `gorm:"uniqueIndex;not null"     json:"ticker"`
	Name         string    `gorm:"not null"                 json:"name"`
	Sector       string    `json:"sector"`
	Score        float64   `json:"score"`
	CurrentPrice float64   `json:"current_price"`
	DailyChange  float64   `json:"daily_change"`
	DY           float64   `json:"dy"`
	HistoryReady bool      `gorm:"default:false" json:"history_ready"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *Stock) Validate() error {
	if strings.TrimSpace(s.Ticker) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(s.Name) == "" {
		return ErrValidation
	}
	if s.Score < 0 || s.Score > 10 {
		return ErrValidation
	}
	if s.CurrentPrice <= 0 {
		return ErrValidation
	}
	if s.DY < 0 {
		return ErrValidation
	}
	return nil
}
