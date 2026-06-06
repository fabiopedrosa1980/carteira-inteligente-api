package handler

import (
	"errors"
	"net/http"
	"strconv"

	"carteira-inteligente-api/internal/adapters/http/dto"
	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	service application.StockUseCase
}

func NewStockHandler(service application.StockUseCase) *StockHandler {
	return &StockHandler{service: service}
}

func (h *StockHandler) CreateStock(c *gin.Context) {
	var req dto.CreateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stock := &domain.Stock{
		Ticker:       req.Ticker,
		Nome:         req.Nome,
		Setor:        req.Setor,
		Nota:         req.Nota,
		PrecoAtual:   req.PrecoAtual,
		VariacaoHoje: req.VariacaoHoje,
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
	if sort != "" && sort != "nota" && sort != "variacao" && sort != "dy" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sort value; accepted: nota, variacao, dy"})
		return
	}

	query := domain.StockQuery{
		Setor: c.Query("setor"),
		Sort:  sort,
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
		Nome:         req.Nome,
		Setor:        req.Setor,
		Nota:         req.Nota,
		PrecoAtual:   req.PrecoAtual,
		VariacaoHoje: req.VariacaoHoje,
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
