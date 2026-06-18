package handler

import (
	"os"
	"testing"

	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"
	"carteira-inteligente-api/internal/infrastructure/persistence"
)

// TestE2E_TransactionCreatesStockAndImportsDividends exercita o fluxo dinamico
// completo contra servicos reais (Yahoo Finance + Investidor10):
// criar transacao de Acao -> stock criado -> dividendos importados.
//
// Rode com: E2E=1 go test ./internal/adapters/http/handler -run TestE2E -v
func TestE2E_TransactionCreatesStockAndImportsDividends(t *testing.T) {
	if os.Getenv("E2E") != "1" {
		t.Skip("teste E2E (rede): defina E2E=1 para executar")
	}

	db, err := persistence.NewDBWithDSN("file:e2e_tx?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("db: %v", err)
	}

	stockRepo := persistence.NewGormStockRepository(db)
	stockSvc := application.NewStockService(stockRepo)
	dividendRepo := persistence.NewGormDividendRepository(db)
	dividendSvc := application.NewDividendService(dividendRepo, stockRepo)
	txRepo := persistence.NewGormTransactionRepository(db)
	txSvc := application.NewTransactionService(txRepo)

	h := NewTransactionHandler(txSvc, stockRepo, stockSvc, dividendSvc)

	const ticker = "BBAS3"

	// Simula o efeito colateral da criacao de uma transacao de Acao.
	h.ensureStockAndImport(ticker, "Ações", false)

	// 1) Stock deve existir no catalogo com history_ready = true.
	stocks, err := stockRepo.FindAll(domain.StockQuery{})
	if err != nil {
		t.Fatalf("listar stocks: %v", err)
	}
	var created *domain.Stock
	for i := range stocks {
		if stocks[i].Ticker == ticker {
			created = &stocks[i]
		}
	}
	if created == nil {
		t.Fatalf("stock %s nao foi criado pela transacao", ticker)
	}
	if created.ID == 0 {
		t.Fatalf("stock %s criado com ID invalido", ticker)
	}
	if !created.HistoryReady {
		t.Errorf("stock %s deveria estar com history_ready=true apos importacao", ticker)
	}
	t.Logf("stock criado: id=%d ticker=%s preco=%.2f history_ready=%v", created.ID, created.Ticker, created.CurrentPrice, created.HistoryReady)

	// 2) Dividendos devem ter sido importados.
	divs, err := dividendRepo.FindByStockID(created.ID)
	if err != nil {
		t.Fatalf("listar dividendos: %v", err)
	}
	if len(divs) == 0 {
		t.Fatalf("nenhum dividendo importado para %s", ticker)
	}
	t.Logf("%s: %d dividendos importados", ticker, len(divs))

	// 3) Idempotencia: chamar de novo nao duplica stock nem dividendos.
	h.ensureStockAndImport(ticker, "Ações", false)
	stocks2, _ := stockRepo.FindAll(domain.StockQuery{})
	count := 0
	for _, s := range stocks2 {
		if s.Ticker == ticker {
			count++
		}
	}
	if count != 1 {
		t.Errorf("esperado 1 stock %s apos segunda chamada, obtido %d", ticker, count)
	}
	divs2, _ := dividendRepo.FindByStockID(created.ID)
	if len(divs2) != len(divs) {
		t.Errorf("reimportacao duplicou dividendos: antes=%d depois=%d", len(divs), len(divs2))
	}
}
