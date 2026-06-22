package application

import (
	"time"

	"carteira-inteligente-api/internal/domain"
)

// Defaults aplicados quando o usuário ainda não salvou nenhuma configuração.
// Devem coincidir com os defaults do frontend (AllocationService).
var defaultAllocation = domain.AllocationConfig{
	AcoesTarget:        50,
	FIIsTarget:         40,
	ETFsTarget:         10,
	ConcentrationLimit: 20,
}

type AllocationService struct {
	repo domain.AllocationRepository
}

func NewAllocationService(repo domain.AllocationRepository) *AllocationService {
	return &AllocationService{repo: repo}
}

// Get retorna a config do usuário; quando ausente, devolve os defaults (sem persistir).
func (s *AllocationService) Get(userID string) (*domain.AllocationConfig, error) {
	cfg, err := s.repo.Get(userID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		d := defaultAllocation
		d.UserID = userID
		return &d, nil
	}
	return cfg, nil
}

// Save grava (upsert) a config do usuário.
func (s *AllocationService) Save(userID string, cfg *domain.AllocationConfig) (*domain.AllocationConfig, error) {
	cfg.UserID = userID
	cfg.UpdatedAt = time.Now()
	if err := s.repo.Upsert(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
