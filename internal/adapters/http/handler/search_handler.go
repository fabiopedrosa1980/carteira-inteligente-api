package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct{}

func NewSearchHandler() *SearchHandler {
	return &SearchHandler{}
}

// TickerSuggestion é o item de sugestão retornado pelo autocomplete.
type TickerSuggestion struct {
	Ticker string `json:"ticker"`
	Name   string `json:"name"`
}

// Search retorna sugestões de tickers da B3 a partir de um termo (>= 3 letras),
// consultando a API de busca do Yahoo Finance e filtrando símbolos `.SA`.
func (h *SearchHandler) Search(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if len(q) < 3 {
		c.JSON(http.StatusOK, []TickerSuggestion{})
		return
	}

	suggestions := searchYahooTickers(q)
	c.JSON(http.StatusOK, suggestions)
}

func searchYahooTickers(query string) []TickerSuggestion {
	out := []TickerSuggestion{}

	url := fmt.Sprintf(
		"https://query2.finance.yahoo.com/v1/finance/search?q=%s&quotesCount=15&newsCount=0&lang=pt-BR&region=BR",
		strings.ToUpper(query),
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return out
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return out
	}
	defer resp.Body.Close()

	var yr struct {
		Quotes []struct {
			Symbol    string `json:"symbol"`
			ShortName string `json:"shortname"`
			LongName  string `json:"longname"`
			QuoteType string `json:"quoteType"`
		} `json:"quotes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&yr); err != nil {
		return out
	}

	seen := map[string]bool{}
	for _, qt := range yr.Quotes {
		// Apenas símbolos da B3 (sufixo .SA).
		if !strings.HasSuffix(qt.Symbol, ".SA") {
			continue
		}
		ticker := strings.TrimSuffix(qt.Symbol, ".SA")
		if ticker == "" || seen[ticker] {
			continue
		}
		seen[ticker] = true
		name := qt.LongName
		if name == "" {
			name = qt.ShortName
		}
		if name == "" {
			name = ticker
		}
		out = append(out, TickerSuggestion{Ticker: ticker, Name: name})
	}
	return out
}
