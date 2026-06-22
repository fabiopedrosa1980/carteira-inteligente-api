package handler

import (
	"net/http"

	"carteira-inteligente-api/internal/adapters/http/dto"
	"carteira-inteligente-api/internal/application"

	"github.com/gin-gonic/gin"
)

type AllocationHandler struct {
	service *application.AllocationService
}

func NewAllocationHandler(service *application.AllocationService) *AllocationHandler {
	return &AllocationHandler{service: service}
}

func (h *AllocationHandler) GetAllocation(c *gin.Context) {
	userID := c.GetString("userID")
	cfg, err := h.service.Get(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, dto.AllocationFromDomain(cfg))
}

func (h *AllocationHandler) PutAllocation(c *gin.Context) {
	userID := c.GetString("userID")
	var req dto.AllocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saved, err := h.service.Save(userID, req.ToDomain())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, dto.AllocationFromDomain(saved))
}
