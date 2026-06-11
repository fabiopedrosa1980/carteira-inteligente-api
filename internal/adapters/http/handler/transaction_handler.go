package handler

import (
	"encoding/json"
	"errors"
	"fmt"
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
			price, changePercent, dividendYield, name := fetchYahooQuote(p.Ticker)
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

	c.JSON(http.StatusOK, items)
}

// fetchYahooQuote busca cotação atual, variação do dia e DY anual no Yahoo Finance.
// Usa o endpoint v7/finance/quote que retorna todos os dados em uma única chamada.
// changePercent já vem como percentual (ex: 1.54 = +1.54%).
// trailingAnnualDividendYield vem como decimal (ex: 0.0854) e é convertido para percentual.
func fetchYahooQuote(ticker string) (price, changePercent, dividendYield float64, name string) {
	client := &http.Client{Timeout: 6 * time.Second}
	url := fmt.Sprintf(
		"https://query2.finance.yahoo.com/v7/finance/quote?symbols=%s.SA&fields=regularMarketPrice,regularMarketChangePercent,trailingAnnualDividendYield,longName,shortName",
		ticker,
	)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, 0, ticker
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; carteira-inteligente/1.0)")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return 0, 0, 0, ticker
	}
	defer resp.Body.Close()

	var qr struct {
		QuoteResponse struct {
			Result []struct {
				LongName                    string  `json:"longName"`
				ShortName                   string  `json:"shortName"`
				RegularMarketPrice          float64 `json:"regularMarketPrice"`
				RegularMarketChangePercent  float64 `json:"regularMarketChangePercent"`
				TrailingAnnualDividendYield float64 `json:"trailingAnnualDividendYield"`
			} `json:"result"`
		} `json:"quoteResponse"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil || len(qr.QuoteResponse.Result) == 0 {
		return 0, 0, 0, ticker
	}

	r := qr.QuoteResponse.Result[0]
	n := r.LongName
	if n == "" {
		n = r.ShortName
	}
	if n == "" {
		n = ticker
	}

	// DY vem como decimal (0.0854 = 8.54%); converte para percentual
	return r.RegularMarketPrice, r.RegularMarketChangePercent, r.TrailingAnnualDividendYield * 100, n
}
