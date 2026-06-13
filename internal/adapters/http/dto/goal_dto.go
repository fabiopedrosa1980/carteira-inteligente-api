package dto

import "carteira-inteligente-api/internal/domain"

type GoalRequest struct {
	Name        string  `json:"name"        binding:"required"`
	TargetValue float64 `json:"targetValue"`
	Type        string  `json:"type"        binding:"required,oneof=patrimonio renda_mensal preco_medio"`
	Ticker      string  `json:"ticker"`
}

type GoalResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	TargetValue     float64 `json:"targetValue"`
	Type            string  `json:"type"`
	Ticker          string  `json:"ticker"`
	CreatedAt       string  `json:"createdAt"`
	CurrentValue    float64 `json:"currentValue"`
	ProgressPercent float64 `json:"progressPercent"`
}

func GoalFromDomain(g *domain.Goal, currentValue, progressPercent float64) GoalResponse {
	return GoalResponse{
		ID:              g.ID,
		Name:            g.Name,
		TargetValue:     g.TargetValue,
		Type:            g.Type,
		Ticker:          g.Ticker,
		CreatedAt:       g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CurrentValue:    currentValue,
		ProgressPercent: progressPercent,
	}
}
