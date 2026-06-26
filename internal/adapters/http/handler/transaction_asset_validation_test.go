package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"
	"carteira-inteligente-api/internal/infrastructure/persistence"

	"github.com/gin-gonic/gin"
)

// setupTxHandler monta um TransactionHandler com um catálogo b3_assets semeado.
func setupTxHandler(t *testing.T, seed []domain.Asset) *TransactionHandler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := persistence.NewDBWithDSN("file:" + t.Name() + "?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	stockRepo := persistence.NewGormStockRepository(db)
	stockSvc := application.NewStockService(stockRepo)
	dividendRepo := persistence.NewGormDividendRepository(db)
	dividendSvc := application.NewDividendService(dividendRepo, stockRepo)
	txRepo := persistence.NewGormTransactionRepository(db)
	txSvc := application.NewTransactionService(txRepo)

	assetRepo := persistence.NewGormAssetRepository(db)
	if len(seed) > 0 {
		if err := assetRepo.Upsert(seed); err != nil {
			t.Fatalf("seed catálogo: %v", err)
		}
	}
	assetSvc := application.NewAssetService(assetRepo)

	return NewTransactionHandler(txSvc, stockRepo, stockSvc, dividendSvc, assetSvc)
}

func postTransaction(t *testing.T, h *TransactionHandler, ticker, assetType string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"ticker":     ticker,
		"asset_type": assetType,
		"quantity":   10,
		"price":      30.5,
		"date":       "2024-01-10",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "user-1")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.CreateTransaction(c)
	return w
}

func TestCreateTransaction_AssetTypeMatchesCatalog(t *testing.T) {
	h := setupTxHandler(t, []domain.Asset{{Ticker: "TAEE11", Type: "Acoes"}})
	w := postTransaction(t, h, "TAEE11", "Acoes")
	if w.Code != http.StatusCreated {
		t.Fatalf("esperava 201 para tipo correto, veio %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateTransaction_AssetTypeDivergesCatalog(t *testing.T) {
	h := setupTxHandler(t, []domain.Asset{{Ticker: "BOVA11", Type: "ETFs"}})
	w := postTransaction(t, h, "BOVA11", "FIIs")
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperava 422 para tipo divergente, veio %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateTransaction_TickerOutsideCatalogIsAccepted(t *testing.T) {
	h := setupTxHandler(t, nil)
	w := postTransaction(t, h, "ZZZZ99", "FIIs")
	if w.Code != http.StatusCreated {
		t.Fatalf("esperava 201 para ticker fora do catálogo, veio %d: %s", w.Code, w.Body.String())
	}
}
