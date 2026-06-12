package dto

import "carteira-inteligente-api/internal/domain"

type GoalRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Description string  `json:"description"`
	TargetValue float64 `json:"targetValue"`
	Type        string  `json:"type"        binding:"required,oneof=patrimonio renda_mensal preco_medio"`
	Ticker      string  `json:"ticker"`
}

type GoalResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	TargetValue float64 `json:"targetValue"`
	Type        string  `json:"type"`
	Ticker      string  `json:"ticker"`
	CreatedAt   string  `json:"createdAt"`
}

func GoalFromDomain(g *domain.Goal) GoalResponse {
	return GoalResponse{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		TargetValue: g.TargetValue,
		Type:        g.Type,
		Ticker:      g.Ticker,
		CreatedAt:   g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func GoalListFromDomain(goals []domain.Goal) []GoalResponse {
	out := make([]GoalResponse, len(goals))
	for i := range goals {
		out[i] = GoalFromDomain(&goals[i])
	}
	return out
}
