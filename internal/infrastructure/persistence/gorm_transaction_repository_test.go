package persistence_test

import (
	"math"
	"testing"
	"time"

	"carteira-inteligente-api/internal/domain"
	"carteira-inteligente-api/internal/infrastructure/persistence"
)

func setupTransactionRepo(t *testing.T) domain.TransactionRepository {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := persistence.NewDBWithDSN(dsn)
	if err != nil {
		t.Fatalf("setup db: %v", err)
	}
	return persistence.NewGormTransactionRepository(db)
}

func tx(userID, ticker string, at domain.AssetType, qty, price float64) *domain.Transaction {
	return &domain.Transaction{
		UserID:    userID,
		Ticker:    ticker,
		AssetType: at,
		Quantity:  qty,
		Price:     price,
		Date:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// Vários lançamentos do mesmo ticker DEVEM ser consolidados em uma única
// posição, somando a quantidade e calculando o preço médio ponderado.
func TestGetAcoesPositions_ConsolidaMesmoTicker(t *testing.T) {
	repo := setupTransactionRepo(t)
	const user = "u1"

	if err := repo.Create(tx(user, "PETR4", domain.AssetTypeAcoes, 10, 30)); err != nil {
		t.Fatalf("create 1: %v", err)
	}
	if err := repo.Create(tx(user, "PETR4", domain.AssetTypeAcoes, 20, 36)); err != nil {
		t.Fatalf("create 2: %v", err)
	}

	positions, err := repo.GetAcoesPositions(user)
	if err != nil {
		t.Fatalf("positions: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("esperado 1 posição consolidada, obtido %d", len(positions))
	}
	p := positions[0]
	if p.Ticker != "PETR4" {
		t.Fatalf("ticker esperado PETR4, obtido %s", p.Ticker)
	}
	if p.TotalQuantity != 30 {
		t.Fatalf("quantidade total esperada 30, obtida %g", p.TotalQuantity)
	}
	// preço médio ponderado: (10*30 + 20*36) / 30 = 34
	if math.Abs(p.AvgPrice-34) > 1e-9 {
		t.Fatalf("preço médio esperado 34, obtido %g", p.AvgPrice)
	}
	if p.TransactionCount != 2 {
		t.Fatalf("contagem de lançamentos esperada 2, obtida %d", p.TransactionCount)
	}
}

// Dados legados gravados com caixa/espaço divergentes DEVEM ser consolidados
// numa única posição (regressão do bug que duplicava o ativo).
func TestGetAcoesPositions_ConsolidaCaixaDivergente(t *testing.T) {
	repo := setupTransactionRepo(t)
	const user = "u1"

	// Insere direto via Create do repositório (sem passar pelo service que
	// normaliza), simulando dados legados gravados com caixa divergente.
	if err := repo.Create(tx(user, "PETR4", domain.AssetTypeAcoes, 10, 30)); err != nil {
		t.Fatalf("create 1: %v", err)
	}
	if err := repo.Create(tx(user, "petr4", domain.AssetTypeAcoes, 5, 40)); err != nil {
		t.Fatalf("create 2: %v", err)
	}

	positions, err := repo.GetAcoesPositions(user)
	if err != nil {
		t.Fatalf("positions: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("esperado 1 posição consolidada (sem duplicação), obtido %d", len(positions))
	}
	if positions[0].TotalQuantity != 15 {
		t.Fatalf("quantidade total esperada 15, obtida %g", positions[0].TotalQuantity)
	}
}

// GetAllPositions DEVE somar posições de todos os tipos de ativo (Ações, FIIs
// e ETFs) — usado no cálculo do patrimônio das Metas.
func TestGetAllPositions_TodosOsTipos(t *testing.T) {
	repo := setupTransactionRepo(t)
	const user = "u1"

	if err := repo.Create(tx(user, "PETR4", domain.AssetTypeAcoes, 10, 30)); err != nil {
		t.Fatalf("create acao: %v", err)
	}
	if err := repo.Create(tx(user, "MXRF11", domain.AssetTypeFIIs, 100, 10)); err != nil {
		t.Fatalf("create fii: %v", err)
	}
	if err := repo.Create(tx(user, "IVVB11", domain.AssetTypeETFs, 5, 300)); err != nil {
		t.Fatalf("create etf: %v", err)
	}

	all, err := repo.GetAllPositions(user)
	if err != nil {
		t.Fatalf("all positions: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("esperadas 3 posições (Ações, FIIs, ETFs), obtidas %d", len(all))
	}

	// GetAcoesPositions continua restrito a Ações.
	acoes, err := repo.GetAcoesPositions(user)
	if err != nil {
		t.Fatalf("acoes positions: %v", err)
	}
	if len(acoes) != 1 {
		t.Fatalf("esperada 1 posição de Ações, obtida %d", len(acoes))
	}
}

// DeleteAll DEVE remover todos os lançamentos do usuário, sem afetar os de
// outros usuários, e ser idempotente quando não há lançamentos.
func TestDeleteAll(t *testing.T) {
	repo := setupTransactionRepo(t)
	const user = "u1"
	const other = "u2"

	if err := repo.Create(tx(user, "PETR4", domain.AssetTypeAcoes, 10, 30)); err != nil {
		t.Fatalf("create 1: %v", err)
	}
	if err := repo.Create(tx(user, "MXRF11", domain.AssetTypeFIIs, 100, 10)); err != nil {
		t.Fatalf("create 2: %v", err)
	}
	if err := repo.Create(tx(other, "VALE3", domain.AssetTypeAcoes, 5, 60)); err != nil {
		t.Fatalf("create other: %v", err)
	}

	if err := repo.DeleteAll(user); err != nil {
		t.Fatalf("delete all: %v", err)
	}

	list, err := repo.List(user, "")
	if err != nil {
		t.Fatalf("list user: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("esperado 0 lançamentos após DeleteAll, obtido %d", len(list))
	}

	// Lançamentos de outro usuário permanecem intactos.
	otherList, err := repo.List(other, "")
	if err != nil {
		t.Fatalf("list other: %v", err)
	}
	if len(otherList) != 1 {
		t.Fatalf("esperado 1 lançamento do outro usuário preservado, obtido %d", len(otherList))
	}

	// Idempotência: chamar de novo com a carteira já vazia não retorna erro.
	if err := repo.DeleteAll(user); err != nil {
		t.Fatalf("delete all idempotente: %v", err)
	}
}
