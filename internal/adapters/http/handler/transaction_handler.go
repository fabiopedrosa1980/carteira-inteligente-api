package handler

import (
	"errors"
	"net/http"
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

func (h *TransactionHandler) GetPortfolio(c *gin.Context) {
	userID := c.GetString("userID")
	items, err := h.service.GetPortfolio(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, dto.PortfolioListFromDomain(items))
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
