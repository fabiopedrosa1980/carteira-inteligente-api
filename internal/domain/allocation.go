package domain

import "time"

// AllocationConfig é a configuração de alocação da carteira de UM usuário
// (singleton por usuário): alvos por classe + limite de concentração por ativo.
type AllocationConfig struct {
	UserID             string    `gorm:"primaryKey" json:"-"`
	AcoesTarget        float64   `json:"-"`
	FIIsTarget         float64   `json:"-"`
	ETFsTarget         float64   `json:"-"`
	ConcentrationLimit float64   `json:"-"`
	UpdatedAt          time.Time `json:"-"`
}

type AllocationRepository interface {
	// Get retorna a config do usuário, ou (nil, nil) quando ainda não existe.
	Get(userID string) (*AllocationConfig, error)
	// Upsert cria ou atualiza a config (chave = UserID).
	Upsert(cfg *AllocationConfig) error
}
