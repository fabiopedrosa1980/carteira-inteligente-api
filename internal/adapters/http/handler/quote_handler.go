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

type brapiResult struct {
	Symbol                     string  `json:"symbol"`
	ShortName                  string  `json:"shortName"`
	LongName                   string  `json:"longName"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
}

type brapiResponse struct {
	Results []brapiResult `json:"results"`
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

func (h *QuoteHandler) GetQuote(c *gin.Context) {
	ticker := strings.ToUpper(c.Param("ticker"))

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet,
		fmt.Sprintf("https://brapi.dev/api/quote/%s", ticker), nil)
	if err != nil {
		c.JSON(http.StatusOK, QuoteResponse{Ticker: ticker, Found: false})
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "carteira-inteligente-api/1.0")

	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, QuoteResponse{Ticker: ticker, Found: false})
		return
	}
	defer resp.Body.Close()

	var br brapiResponse
	if err := json.NewDecoder(resp.Body).Decode(&br); err != nil || len(br.Results) == 0 {
		c.JSON(http.StatusOK, QuoteResponse{Ticker: ticker, Found: false})
		return
	}

	r := br.Results[0]
	name := r.LongName
	if name == "" {
		name = r.ShortName
	}
	if name == "" {
		name = ticker
	}

	c.JSON(http.StatusOK, QuoteResponse{
		Ticker:        ticker,
		Name:          name,
		Price:         r.RegularMarketPrice,
		ChangePercent: r.RegularMarketChangePercent,
		PrevClose:     r.RegularMarketPreviousClose,
		Found:         r.RegularMarketPrice > 0,
	})
}
