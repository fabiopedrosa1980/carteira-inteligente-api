package handler

import (
	"testing"
	"time"

	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"
	"carteira-inteligente-api/internal/infrastructure/persistence"
)

// Verifica que buildGoalResponses retorna progressPercent > 0 quando o usuario
// tem posicoes em acoes — reproduz a queixa "percentual nao aparece".
func TestBuildGoalResponses_ComputesProgress(t *testing.T) {
	db, err := persistence.NewDBWithDSN("file:goal_progress?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("db: %v", err)
	}

	stockRepo := persistence.NewGormStockRepository(db)
	txRepo := persistence.NewGormTransactionRepository(db)
	txSvc := application.NewTransactionService(txRepo)
	goalRepo := persistence.NewGormGoalRepository(db)
	goalSvc := application.NewGoalService(goalRepo)

	const userID = "user-1"

	// Posicao: 100 acoes a R$ 50 -> patrimonio minimo R$ 5.000 (fallback avgPrice).
	tx := &domain.Transaction{
		UserID:    userID,
		Ticker:    "BBAS3",
		AssetType: domain.AssetTypeAcoes,
		Quantity:  100,
		Price:     50,
		Date:      time.Now(),
	}
	if err := txSvc.Create(tx); err != nil {
		t.Fatalf("criar transacao: %v", err)
	}

	// Meta de R$ 10.000.
	goal := &domain.Goal{Name: "Primeiros 10 mil", TargetValue: 10000}
	if err := goalSvc.CreateGoal(userID, goal); err != nil {
		t.Fatalf("criar meta: %v", err)
	}

	h := NewGoalHandler(goalSvc, txSvc, stockRepo)
	goals, _ := goalSvc.ListGoals(userID)
	resp := h.buildGoalResponses(userID, goals)

	if len(resp) != 1 {
		t.Fatalf("esperava 1 meta, obteve %d", len(resp))
	}
	r := resp[0]
	t.Logf("currentValue=%.2f targetValue=%.2f progressPercent=%.2f", r.CurrentValue, r.TargetValue, r.ProgressPercent)

	if r.CurrentValue <= 0 {
		t.Errorf("currentValue deveria ser > 0 (ha posicoes), obteve %.2f", r.CurrentValue)
	}
	if r.ProgressPercent <= 0 {
		t.Errorf("progressPercent deveria ser > 0, obteve %.2f", r.ProgressPercent)
	}
}
