package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func createStock(t *testing.T, r interface{ ServeHTTP(http.ResponseWriter, *http.Request) }, ticker string, preco float64) int {
	t.Helper()
	body := toJSON(t, map[string]any{"ticker": ticker, "nome": ticker, "preco_atual": preco})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("createStock: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	return int(resp["id"].(float64))
}

func TestCreateDividend_201(t *testing.T) {
	r := setupRouter(t)
	id := createStock(t, r, "BBAS3", 55.20)

	body := toJSON(t, map[string]any{
		"amount":   0.45,
		"month":    3,
		"year":     2024,
		"type":     "dividendo",
		"ex_date":  "2024-03-15",
		"pay_date": "2024-03-20",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks/"+strconv.Itoa(id)+"/dividends", body)
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
	if resp["amount"].(float64) != 0.45 {
		t.Fatalf("expected amount=0.45, got %v", resp["amount"])
	}
}

func TestCreateDividend_404_StockNotFound(t *testing.T) {
	r := setupRouter(t)
	body := toJSON(t, map[string]any{"amount": 0.45, "month": 3, "year": 2024, "type": "dividendo"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks/9999/dividends", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCreateDividend_400_InvalidType(t *testing.T) {
	r := setupRouter(t)
	id := createStock(t, r, "PETR4", 37.90)
	body := toJSON(t, map[string]any{"amount": 0.45, "month": 3, "year": 2024, "type": "invalido"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks/"+strconv.Itoa(id)+"/dividends", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListDividends_200(t *testing.T) {
	r := setupRouter(t)
	id := createStock(t, r, "VALE3", 65.0)

	for _, month := range []int{1, 4, 7} {
		body := toJSON(t, map[string]any{"amount": 0.30, "month": month, "year": 2024, "type": "dividendo"})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks/"+strconv.Itoa(id)+"/dividends", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stocks/"+strconv.Itoa(id)+"/dividends", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 3 {
		t.Fatalf("expected 3 dividends, got %d", len(resp))
	}
}

func TestListDividends_FilterByYear(t *testing.T) {
	r := setupRouter(t)
	id := createStock(t, r, "CMIG4", 11.95)

	for _, year := range []int{2023, 2024} {
		body := toJSON(t, map[string]any{"amount": 0.22, "month": 6, "year": year, "type": "dividendo"})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks/"+strconv.Itoa(id)+"/dividends", body)
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(httptest.NewRecorder(), req)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/stocks/"+strconv.Itoa(id)+"/dividends?year=2024", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 dividend for 2024, got %d", len(resp))
	}
	if int(resp[0]["year"].(float64)) != 2024 {
		t.Fatalf("expected year=2024, got %v", resp[0]["year"])
	}
}

func TestGetMonthlySummary_Always12Months(t *testing.T) {
	r := setupRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/dividends/monthly", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp []map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 12 {
		t.Fatalf("expected 12 months, got %d", len(resp))
	}
	if resp[0]["month_name"] == nil || resp[0]["month_name"].(string) != "Janeiro" {
		t.Fatalf("expected first month to be Janeiro, got %v", resp[0]["month_name"])
	}
}

func TestGetMonthlySummary_WithData(t *testing.T) {
	r := setupRouter(t)
	id := createStock(t, r, "BBSE3", 34.80)

	body := toJSON(t, map[string]any{"amount": 0.45, "month": 6, "year": 2024, "type": "dividendo"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/stocks/"+strconv.Itoa(id)+"/dividends", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), req)

	w := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/dividends/monthly?year=2024", nil)
	r.ServeHTTP(w, req2)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp []map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 12 {
		t.Fatalf("expected 12 months, got %d", len(resp))
	}

	june := resp[5]
	if int(june["stock_count"].(float64)) != 1 {
		t.Fatalf("expected stock_count=1 in June, got %v", june["stock_count"])
	}
	if june["avg_total"].(float64) != 0.45 {
		t.Fatalf("expected avg_total=0.45 in June, got %v", june["avg_total"])
	}

	jan := resp[0]
	if int(jan["stock_count"].(float64)) != 0 {
		t.Fatalf("expected stock_count=0 in January (no data), got %v", jan["stock_count"])
	}
}

func TestGetMonthlySummary_InvalidYear(t *testing.T) {
	r := setupRouter(t)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/dividends/monthly?year=abc", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
