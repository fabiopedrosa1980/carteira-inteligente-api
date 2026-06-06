package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"carteira-inteligente-api/internal/adapters/http/handler"
	"carteira-inteligente-api/internal/adapters/http/router"
	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/infrastructure/persistence"

	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := persistence.NewDBWithDSN(dsn)
	if err != nil {
		t.Fatalf("failed to setup db: %v", err)
	}
	stockRepo := persistence.NewGormStockRepository(db)
	svc := application.NewStockService(stockRepo)
	h := handler.NewStockHandler(svc)
	dividendRepo := persistence.NewGormDividendRepository(db)
	divSvc := application.NewDividendService(dividendRepo, stockRepo)
	divH := handler.NewDividendHandler(divSvc)
	return router.SetupRouter(h, divH)
}

func toJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return bytes.NewBuffer(b)
}

func TestCreateStock_201(t *testing.T) {
	r := setupRouter(t)
	body := toJSON(t, map[string]any{
		"ticker":        "PETR4",
		"nome":          "Petrobras",
		"setor":         "Energia",
		"nota":          8.5,
		"preco_atual":   35.50,
		"variacao_hoje": -1.2,
		"dy":            6.5,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["id"] == nil || resp["id"].(float64) == 0 {
		t.Fatal("expected id in response")
	}
	if resp["dy"] == nil || resp["dy"].(float64) != 6.5 {
		t.Fatalf("expected dy=6.5 in response, got %v", resp["dy"])
	}
}

func TestCreateStock_201_WithoutDY(t *testing.T) {
	r := setupRouter(t)
	body := toJSON(t, map[string]any{
		"ticker":      "BBDC4",
		"nome":        "Bradesco",
		"preco_atual": 20.0,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["dy"] == nil || resp["dy"].(float64) != 0.0 {
		t.Fatalf("expected dy=0.0 when omitted, got %v", resp["dy"])
	}
}

func TestCreateStock_400_MissingTicker(t *testing.T) {
	r := setupRouter(t)
	body := toJSON(t, map[string]any{"nome": "Petrobras", "preco_atual": 35.50})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateStock_400_InvalidNota(t *testing.T) {
	r := setupRouter(t)
	body := toJSON(t, map[string]any{"ticker": "PETR4", "nome": "Petrobras", "preco_atual": 35.50, "nota": 15.0})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateStock_400_PrecoNegativo(t *testing.T) {
	r := setupRouter(t)
	body := toJSON(t, map[string]any{"ticker": "PETR4", "nome": "Petrobras", "preco_atual": -5.0})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateStock_409_Duplicate(t *testing.T) {
	r := setupRouter(t)
	payload := map[string]any{"ticker": "VALE3", "nome": "Vale", "preco_atual": 65.0}
	body := toJSON(t, payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), req)

	body = toJSON(t, payload)
	w := httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetStock_200(t *testing.T) {
	r := setupRouter(t)
	body := toJSON(t, map[string]any{"ticker": "ITUB4", "nome": "Itaú", "preco_atual": 30.0})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var created map[string]any
	json.Unmarshal(w.Body.Bytes(), &created)
	id := int(created["id"].(float64))

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/stocks/"+strconv.Itoa(id), nil)
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}
}

func TestGetStock_404(t *testing.T) {
	r := setupRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stocks/9999", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestListStocks_200_Empty(t *testing.T) {
	r := setupRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stocks", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 0 {
		t.Fatalf("expected empty array, got %v", resp)
	}
}

func TestListStocks_FilterSetor(t *testing.T) {
	r := setupRouter(t)
	create := func(ticker, nome, setor string, preco float64) {
		body := toJSON(t, map[string]any{"ticker": ticker, "nome": nome, "setor": setor, "preco_atual": preco})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
	}
	create("PETR4", "Petrobras", "Energia", 35.0)
	create("VALE3", "Vale", "Mineração", 65.0)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stocks?setor=Energia", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 1 || resp[0]["ticker"] != "PETR4" {
		t.Fatalf("expected 1 stock PETR4, got %v", resp)
	}
}

func TestListStocks_400_InvalidSort(t *testing.T) {
	r := setupRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stocks?sort=invalido", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListStocks_SortByNota(t *testing.T) {
	r := setupRouter(t)
	for _, s := range []map[string]any{
		{"ticker": "A1", "nome": "A", "preco_atual": 10.0, "nota": 5.0},
		{"ticker": "B2", "nome": "B", "preco_atual": 10.0, "nota": 9.0},
		{"ticker": "C3", "nome": "C", "preco_atual": 10.0, "nota": 3.0},
	} {
		body := toJSON(t, s)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stocks?sort=nota", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 3 || resp[0]["ticker"] != "B2" {
		t.Fatalf("expected first stock B2 (nota 9), got %v", resp)
	}
}

func TestUpdateStock_DY(t *testing.T) {
	r := setupRouter(t)
	body := toJSON(t, map[string]any{"ticker": "ITSA4", "nome": "Itaúsa", "preco_atual": 10.0, "dy": 3.0})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
	req.Header.Set("Content-Type", "application/json")
	wc := httptest.NewRecorder()
	r.ServeHTTP(wc, req)

	var created map[string]any
	json.Unmarshal(wc.Body.Bytes(), &created)
	id := int(created["id"].(float64))

	updateBody := toJSON(t, map[string]any{"ticker": "ITSA4", "nome": "Itaúsa", "preco_atual": 10.0, "dy": 8.5})
	w := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPut, "/api/v1/stocks/"+strconv.Itoa(id), updateBody)
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req2)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["dy"] == nil || resp["dy"].(float64) != 8.5 {
		t.Fatalf("expected dy=8.5 after update, got %v", resp["dy"])
	}
}

func TestListStocks_SortByDY(t *testing.T) {
	r := setupRouter(t)
	for _, s := range []map[string]any{
		{"ticker": "X1", "nome": "X1", "preco_atual": 10.0, "dy": 2.0},
		{"ticker": "X2", "nome": "X2", "preco_atual": 10.0, "dy": 9.0},
		{"ticker": "X3", "nome": "X3", "preco_atual": 10.0, "dy": 5.0},
	} {
		body := toJSON(t, s)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stocks?sort=dy", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 3 || resp[0]["ticker"] != "X2" {
		t.Fatalf("expected first stock X2 (dy 9.0), got %v", resp)
	}
}

func TestDeleteStock_204(t *testing.T) {
	r := setupRouter(t)
	body := toJSON(t, map[string]any{"ticker": "BBAS3", "nome": "Banco do Brasil", "preco_atual": 25.0})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
	req.Header.Set("Content-Type", "application/json")
	wc := httptest.NewRecorder()
	r.ServeHTTP(wc, req)

	var created map[string]any
	json.Unmarshal(wc.Body.Bytes(), &created)
	id := int(created["id"].(float64))

	w := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodDelete, "/api/v1/stocks/"+strconv.Itoa(id), nil)
	r.ServeHTTP(w, req2)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestDeleteStock_404(t *testing.T) {
	r := setupRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/stocks/9999", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
