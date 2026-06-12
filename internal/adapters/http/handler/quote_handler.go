package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

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

func (h *QuoteHandler) GetQuote(c *gin.Context) {
	ticker := strings.ToUpper(c.Param("ticker"))

	if q := fetchYahoo(ticker); q != nil {
		c.JSON(http.StatusOK, q)
		return
	}

	c.JSON(http.StatusOK, QuoteResponse{Ticker: ticker, Found: false})
}
