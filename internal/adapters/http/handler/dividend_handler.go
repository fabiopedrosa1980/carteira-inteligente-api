package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"carteira-inteligente-api/internal/adapters/http/dto"
	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type DividendHandler struct {
	service application.DividendUseCase
}

func NewDividendHandler(service application.DividendUseCase) *DividendHandler {
	return &DividendHandler{service: service}
}

func (h *DividendHandler) CreateDividend(c *gin.Context) {
	stockID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req dto.CreateDividendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d := &domain.Dividend{
		Amount:  req.Amount,
		Month:   req.Month,
		Year:    req.Year,
		Type:    req.Type,
		ExDate:  req.ExDate,
		PayDate: req.PayDate,
	}

	if err := h.service.CreateDividend(stockID, d); err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "stock not found"})
		case errors.Is(err, domain.ErrValidation), errors.Is(err, domain.ErrInvalidDividendType):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.DividendFromDomain(d))
}

func (h *DividendHandler) ListDividends(c *gin.Context) {
	stockID, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var year *int
	if y := c.Query("year"); y != "" {
		parsed, err := strconv.Atoi(y)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
			return
		}
		year = &parsed
	}

	dividends, err := h.service.ListDividendsByStock(stockID, year)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "stock not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, dto.DividendListFromDomain(dividends))
}

func (h *DividendHandler) GetMonthlySummary(c *gin.Context) {
	year := time.Now().Year()
	if y := c.Query("year"); y != "" {
		parsed, err := strconv.Atoi(y)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
			return
		}
		year = parsed
	}

	summaries, err := h.service.GetMonthlySummary(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, summaries)
}
