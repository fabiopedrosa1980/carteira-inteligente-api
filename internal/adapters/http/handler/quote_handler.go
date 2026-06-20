package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// truncate2 corta o valor em 2 casas decimais sem arredondar (ex.: 36.8399 → 36.83).
func truncate2(v float64) float64 {
	return math.Trunc(v*100) / 100
}

type QuoteHandler struct{}

func NewQuoteHandler() *QuoteHandler {
	return &QuoteHandler{}
}

type QuoteResponse struct {
	Ticker        string  `json:"ticker"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	ChangePercent float64 `json:"changePercent"`
	PrevClose     float64 `json:"prevClose"`
	DividendYield float64 `json:"dividendYield"`
	Sector        string  `json:"sector"`
	Found         bool    `json:"found"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

// fetchYahoo usa a API do Yahoo Finance com sufixo .SA para tickers da B3.
func fetchYahoo(ticker string) *QuoteResponse {
	type chartMeta struct {
		LongName           string  `json:"longName"`
		ShortName          string  `json:"shortName"`
		RegularMarketPrice float64 `json:"regularMarketPrice"`
		ChartPreviousClose float64 `json:"chartPreviousClose"`
	}
	type chartResult struct {
		Meta chartMeta `json:"meta"`
	}
	type chart struct {
		Result []chartResult `json:"result"`
	}
	type yahooResp struct {
		Chart chart `json:"chart"`
	}

	for _, host := range []string{"query2", "query1"} {
		url := fmt.Sprintf("https://%s.finance.yahoo.com/v8/finance/chart/%s.SA?interval=1d&range=1d", host, ticker)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36")
		req.Header.Set("Accept", "*/*")

		resp, err := httpClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			continue
		}

		var yr yahooResp
		err = json.NewDecoder(resp.Body).Decode(&yr)
		resp.Body.Close()
		if err != nil || len(yr.Chart.Result) == 0 {
			continue
		}

		meta := yr.Chart.Result[0].Meta
		if meta.RegularMarketPrice == 0 {
			continue
		}

		prev := meta.ChartPreviousClose
		if prev == 0 {
			prev = meta.RegularMarketPrice
		}
		changePercent := 0.0
		if prev > 0 {
			changePercent = ((meta.RegularMarketPrice - prev) / prev) * 100
		}

		name := meta.LongName
		if name == "" {
			name = meta.ShortName
		}
		return &QuoteResponse{
			Ticker:        ticker,
			Name:          name,
			Price:         meta.RegularMarketPrice,
			ChangePercent: changePercent,
			PrevClose:     prev,
			Found:         true,
		}
	}
	return nil
}

// fetchYahooOnDate retorna o fechamento do ativo na data informada (ou no pregão
// anterior mais próximo, cobrindo fim de semana/feriado). dateStr no formato
// YYYY-MM-DD. Reaproveita o sufixo .SA e os hosts/headers do fetchYahoo.
func fetchYahooOnDate(ticker, dateStr string) *QuoteResponse {
	d, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil
	}
	// Janela: data-7d até data+1d (cobre dias sem pregão; limite superior inclusivo).
	p1 := d.AddDate(0, 0, -7).Unix()
	p2 := d.AddDate(0, 0, 1).Unix()

	type chartMeta struct {
		LongName  string `json:"longName"`
		ShortName string `json:"shortName"`
	}
	type indicators struct {
		Quote []struct {
			Close []float64 `json:"close"`
		} `json:"quote"`
	}
	type chartResult struct {
		Meta       chartMeta  `json:"meta"`
		Timestamp  []int64    `json:"timestamp"`
		Indicators indicators `json:"indicators"`
	}
	type chart struct {
		Result []chartResult `json:"result"`
	}
	type yahooResp struct {
		Chart chart `json:"chart"`
	}

	for _, host := range []string{"query2", "query1"} {
		url := fmt.Sprintf("https://%s.finance.yahoo.com/v8/finance/chart/%s.SA?interval=1d&period1=%d&period2=%d", host, ticker, p1, p2)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36")
		req.Header.Set("Accept", "*/*")

		resp, err := httpClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			continue
		}

		var yr yahooResp
		err = json.NewDecoder(resp.Body).Decode(&yr)
		resp.Body.Close()
		if err != nil || len(yr.Chart.Result) == 0 {
			continue
		}

		res := yr.Chart.Result[0]
		if len(res.Indicators.Quote) == 0 {
			continue
		}
		closes := res.Indicators.Quote[0].Close

		// Último fechamento cujo dia (UTC) seja <= data pedida.
		price := 0.0
		found := false
		for i, ts := range res.Timestamp {
			if i >= len(closes) {
				break
			}
			day := time.Unix(ts, 0).UTC().Format("2006-01-02")
			if day <= dateStr && closes[i] > 0 {
				price = closes[i]
				found = true
			}
		}
		if !found {
			continue
		}

		name := res.Meta.LongName
		if name == "" {
			name = res.Meta.ShortName
		}
		return &QuoteResponse{
			Ticker: ticker,
			Name:   name,
			Price:  truncate2(price),
			Found:  true,
		}
	}
	return nil
}

func (h *QuoteHandler) GetQuote(c *gin.Context) {
	ticker := strings.ToUpper(c.Param("ticker"))
	date := c.Query("date")

	// Data anterior a hoje → preço de fechamento naquela data.
	if date != "" && date < time.Now().Format("2006-01-02") {
		if q := fetchYahooOnDate(ticker, date); q != nil {
			c.JSON(http.StatusOK, q)
			return
		}
		c.JSON(http.StatusOK, QuoteResponse{Ticker: ticker, Found: false})
		return
	}

	if q := fetchYahoo(ticker); q != nil {
		c.JSON(http.StatusOK, q)
		return
	}

	c.JSON(http.StatusOK, QuoteResponse{Ticker: ticker, Found: false})
}
