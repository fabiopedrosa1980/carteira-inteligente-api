package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"carteira-inteligente-api/internal/adapters/http/dto"
	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	service application.TransactionUseCase
}

func NewTransactionHandler(service application.TransactionUseCase) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	userID := c.GetString("userID")

	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	t := &domain.Transaction{
		UserID:    userID,
		Ticker:    req.Ticker,
		AssetType: req.AssetType,
		Quantity:  req.Quantity,
		Price:     req.Price,
		Date:      date,
	}

	if err := h.service.Create(t); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, dto.TransactionFromDomain(t))
}

func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	userID := c.GetString("userID")
	ticker := c.Query("ticker")
	list, err := h.service.List(userID, ticker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, dto.TransactionListFromDomain(list))
}

func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	userID := c.GetString("userID")

	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.Delete(userID, id); err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TransactionHandler) GetAcoes(c *gin.Context) {
	userID := c.GetString("userID")

	positions, err := h.service.GetAcoesPositions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	items := make([]*domain.AcaoItem, len(positions))
	var wg sync.WaitGroup
	for i, pos := range positions {
		wg.Add(1)
		go func(idx int, p *domain.AcoesPosition) {
			defer wg.Done()
			price, changePercent, name, dividendYield := fetchYahooQuote(p.Ticker)
			items[idx] = &domain.AcaoItem{
				Ticker:        p.Ticker,
				Name:          name,
				TotalQuantity: p.TotalQuantity,
				AvgPrice:      p.AvgPrice,
				CurrentPrice:  price,
				ChangePercent: changePercent,
				DividendYield: dividendYield,
			}
		}(i, pos)
	}
	wg.Wait()

	computeNotas(items)

	c.JSON(http.StatusOK, items)
}

// computeNotas atribui uma nota de 1 a 10 para cada posicao, comparando
// rendimento e dividend yield entre todas as posicoes do usuario via
// normalizacao min-max.
func computeNotas(items []*domain.AcaoItem) {
	if len(items) == 0 {
		return
	}

	rendimentos := make([]float64, len(items))
	dys := make([]float64, len(items))
	for i, it := range items {
		r := 0.0
		if it.AvgPrice != 0 {
			r = (it.CurrentPrice - it.AvgPrice) / it.AvgPrice * 100
		}
		rendimentos[i] = r
		dys[i] = it.DividendYield
	}

	normRend := minMaxNormalize(rendimentos)
	normDY := minMaxNormalize(dys)

	for i, it := range items {
		combined := 0.5*normRend[i] + 0.5*normDY[i]
		it.Nota = round1(1 + 9*combined)
	}
}

// minMaxNormalize normaliza os valores para [0,1]. Se max == min,
// retorna 0.5 para todos.
func minMaxNormalize(values []float64) []float64 {
	out := make([]float64, len(values))
	if len(values) == 0 {
		return out
	}
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	if max == min {
		for i := range out {
			out[i] = 0.5
		}
		return out
	}
	for i, v := range values {
		out[i] = (v - min) / (max - min)
	}
	return out
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func fetchYahooQuote(ticker string) (price, changePercent float64, name string, dividendYield float64) {
	client := &http.Client{Timeout: 6 * time.Second}
	url := fmt.Sprintf("https://query2.finance.yahoo.com/v8/finance/chart/%s.SA?interval=1d&range=1y&events=div", ticker)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, ticker, 0
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return 0, 0, ticker, 0
	}
	defer resp.Body.Close()

	var yr struct {
		Chart struct {
			Result []struct {
				Meta struct {
					RegularMarketPrice float64 `json:"regularMarketPrice"`
					ChartPreviousClose float64 `json:"chartPreviousClose"`
					LongName           string  `json:"longName"`
					ShortName          string  `json:"shortName"`
				} `json:"meta"`
				Events struct {
					Dividends map[string]struct {
						Amount float64 `json:"amount"`
						Date   int64   `json:"date"`
					} `json:"dividends"`
				} `json:"events"`
			} `json:"result"`
		} `json:"chart"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&yr); err != nil || len(yr.Chart.Result) == 0 {
		return 0, 0, ticker, 0
	}

	result := yr.Chart.Result[0]
	meta := result.Meta
	n := meta.LongName
	if n == "" {
		n = meta.ShortName
	}
	if n == "" {
		n = ticker
	}

	prev := meta.ChartPreviousClose
	cp := 0.0
	if prev > 0 {
		cp = (meta.RegularMarketPrice - prev) / prev * 100
	}

	sumDividends := 0.0
	for _, d := range result.Events.Dividends {
		sumDividends += d.Amount
	}
	dy := 0.0
	if meta.RegularMarketPrice > 0 {
		dy = sumDividends / meta.RegularMarketPrice * 100
	}

	return meta.RegularMarketPrice, cp, n, dy
}
