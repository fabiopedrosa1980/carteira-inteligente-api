package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"carteira-inteligente-api/internal/adapters/http/dto"
	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"
	"carteira-inteligente-api/internal/infrastructure/scraper"

	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	service     application.StockUseCase
	dividendSvc application.DividendUseCase
}

func NewStockHandler(service application.StockUseCase, dividendSvc application.DividendUseCase) *StockHandler {
	return &StockHandler{service: service, dividendSvc: dividendSvc}
}

func (h *StockHandler) CreateStock(c *gin.Context) {
	var req dto.CreateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stock := &domain.Stock{
		Ticker:       req.Ticker,
		Name:         req.Name,
		Sector:       req.Sector,
		Score:        req.Score,
		CurrentPrice: req.CurrentPrice,
		DailyChange:  req.DailyChange,
		DY:           req.DY,
	}

	if err := h.service.CreateStock(stock); err != nil {
		switch {
		case errors.Is(err, domain.ErrDuplicate):
			c.JSON(http.StatusConflict, gin.H{"error": "ticker already exists"})
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.FromDomain(stock))

	// Import dividend history from Investidor10 in background.
	go h.importDividends(stock.ID, stock.Ticker)
}

func (h *StockHandler) importDividends(stockID uint, ticker string) {
	since := time.Now().AddDate(-5, 0, 0)
	dividends, err := scraper.FetchDividends(ticker, since)
	if err != nil {
		log.Printf("[scraper] %s: %v", ticker, err)
		return
	}
	for _, d := range dividends {
		div := &domain.Dividend{
			StockID: stockID,
			Amount:  d.Amount,
			Month:   d.Month,
			Year:    d.Year,
			Type:    d.Type,
			ExDate:  d.ExDate,
			PayDate: d.PayDate,
		}
		if err := h.dividendSvc.CreateIfNotExists(div); err != nil {
			log.Printf("[scraper] insert %s %s: %v", ticker, d.ExDate, err)
		}
	}
	if err := h.service.UpdateHistoryReady(stockID, true); err != nil {
		log.Printf("[scraper] mark history_ready %s: %v", ticker, err)
	}
	log.Printf("[scraper] %s: imported %d dividends", ticker, len(dividends))
}

func (h *StockHandler) GetStock(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	stock, err := h.service.GetStockByID(id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "stock not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, dto.FromDomain(stock))
}

func (h *StockHandler) ListStocks(c *gin.Context) {
	sort := c.Query("sort")
	if sort != "" && sort != "score" && sort != "daily_change" && sort != "dy" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sort value; accepted: score, daily_change, dy"})
		return
	}

	query := domain.StockQuery{
		Sector: c.Query("sector"),
		Sort:   sort,
	}

	stocks, err := h.service.ListStocks(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, dto.FromDomainList(stocks))
}

func (h *StockHandler) UpdateStock(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req dto.UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated := &domain.Stock{
		Ticker:       req.Ticker,
		Name:         req.Name,
		Sector:       req.Sector,
		Score:        req.Score,
		CurrentPrice: req.CurrentPrice,
		DailyChange:  req.DailyChange,
		DY:           req.DY,
	}

	stock, err := h.service.UpdateStock(id, updated)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "stock not found"})
		case errors.Is(err, domain.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.FromDomain(stock))
}

func (h *StockHandler) DeleteStock(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteStock(id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "stock not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
