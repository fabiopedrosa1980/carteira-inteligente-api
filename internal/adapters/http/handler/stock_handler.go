package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	go h.importDividends(stock.ID, stock.Ticker, isFIISector(stock.Sector))
}

func (h *StockHandler) importDividends(stockID uint, ticker string, fii bool) {
	importDividendsForStock(h.dividendSvc, h.service, stockID, ticker, fii)
}

// isFIISector indica se o setor corresponde a um Fundo Imobiliário.
func isFIISector(sector string) bool {
	return strings.EqualFold(sector, "FIIs")
}

// importDividendsForStock busca o histórico de dividendos do Investidor10 para
// o ticker e persiste cada registro, marcando o stock como history_ready ao
// final. Reutilizado pelo cadastro de stock e pela criação de transações.
func importDividendsForStock(dividendSvc application.DividendUseCase, stockSvc application.StockUseCase, stockID uint, ticker string, fii bool) {
	// Janela de 5 anos-calendário: do dia 1º de janeiro de (ano atual - 4) em
	// diante. Usar AddDate(-5,...) abrangia 6 anos-calendário (o ano de 5 anos
	// atrás entrava parcialmente), divergindo dos filtros de 5 anos da UI.
	since := time.Date(time.Now().Year()-4, time.January, 1, 0, 0, 0, 0, time.UTC)
	dividends, err := scraper.FetchDividendsForType(ticker, since, fii)
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
		if err := dividendSvc.CreateIfNotExists(div); err != nil {
			log.Printf("[scraper] insert %s %s: %v", ticker, d.ExDate, err)
		}
	}
	if err := stockSvc.UpdateHistoryReady(stockID, true); err != nil {
		log.Printf("[scraper] mark history_ready %s: %v", ticker, err)
	}
	log.Printf("[scraper] %s: imported %d dividends", ticker, len(dividends))
}

// StartDividendSync reimporta periodicamente o histórico de dividendos de todos
// os stocks cadastrados. O import inicial ocorre apenas no cadastro do stock, de
// modo que proventos publicados depois (ex.: novos JCP/dividendos do ano) nunca
// entrariam. Este job roda imediatamente e a cada `interval`, mantendo a base
// atualizada. É idempotente: CreateIfNotExists evita duplicar registros.
func StartDividendSync(stockSvc application.StockUseCase, dividendSvc application.DividendUseCase, interval time.Duration) {
	go func() {
		for {
			refreshAllDividends(stockSvc, dividendSvc)
			time.Sleep(interval)
		}
	}()
}

func refreshAllDividends(stockSvc application.StockUseCase, dividendSvc application.DividendUseCase) {
	stocks, err := stockSvc.ListStocks(domain.StockQuery{})
	if err != nil {
		log.Printf("[scraper] sync: list stocks: %v", err)
		return
	}
	for _, s := range stocks {
		importDividendsForStock(dividendSvc, stockSvc, s.ID, s.Ticker, isFIISector(s.Sector))
	}
	log.Printf("[scraper] sync: refreshed dividends for %d stocks", len(stocks))
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
