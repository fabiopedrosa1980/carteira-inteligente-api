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
	Nome         string    `gorm:"not null"                 json:"nome"`
	Setor        string    `json:"setor"`
	Nota         float64   `json:"nota"`
	PrecoAtual   float64   `json:"preco_atual"`
	VariacaoHoje float64   `json:"variacao_hoje"`
	DY           float64   `json:"dy"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (s *Stock) Validate() error {
	if strings.TrimSpace(s.Ticker) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(s.Nome) == "" {
		return ErrValidation
	}
	if s.Nota < 0 || s.Nota > 10 {
		return ErrValidation
	}
	if s.PrecoAtual <= 0 {
		return ErrValidation
	}
	if s.DY < 0 {
		return ErrValidation
	}
	return nil
}
