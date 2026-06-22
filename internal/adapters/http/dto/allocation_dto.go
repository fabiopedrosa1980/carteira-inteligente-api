package dto

import "carteira-inteligente-api/internal/domain"

// Alvos por classe (percentuais).
type AllocationTargets struct {
	Acoes float64 `json:"Acoes"`
	FIIs  float64 `json:"FIIs"`
	ETFs  float64 `json:"ETFs"`
}

type AllocationRequest struct {
	Targets            AllocationTargets `json:"targets"`
	ConcentrationLimit float64           `json:"concentrationLimit"`
}

type AllocationResponse struct {
	Targets            AllocationTargets `json:"targets"`
	ConcentrationLimit float64           `json:"concentrationLimit"`
}

func AllocationFromDomain(c *domain.AllocationConfig) AllocationResponse {
	return AllocationResponse{
		Targets: AllocationTargets{
			Acoes: c.AcoesTarget,
			FIIs:  c.FIIsTarget,
			ETFs:  c.ETFsTarget,
		},
		ConcentrationLimit: c.ConcentrationLimit,
	}
}

func (r *AllocationRequest) ToDomain() *domain.AllocationConfig {
	return &domain.AllocationConfig{
		AcoesTarget:        r.Targets.Acoes,
		FIIsTarget:         r.Targets.FIIs,
		ETFsTarget:         r.Targets.ETFs,
		ConcentrationLimit: r.ConcentrationLimit,
	}
}
