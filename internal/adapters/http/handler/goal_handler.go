package handler

import (
	"errors"
	"net/http"

	"carteira-inteligente-api/internal/adapters/http/dto"
	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type GoalHandler struct {
	service *application.GoalService
}

func NewGoalHandler(service *application.GoalService) *GoalHandler {
	return &GoalHandler{service: service}
}

func (h *GoalHandler) ListGoals(c *gin.Context) {
	goals, err := h.service.ListGoals()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, dto.GoalListFromDomain(goals))
}

func (h *GoalHandler) CreateGoal(c *gin.Context) {
	var req dto.GoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	g := &domain.Goal{
		Name:        req.Name,

		TargetValue: req.TargetValue,
		Type:        req.Type,
		Ticker:      req.Ticker,
	}
	if err := h.service.CreateGoal(g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, dto.GoalFromDomain(g))
}

func (h *GoalHandler) UpdateGoal(c *gin.Context) {
	id := c.Param("id")
	var req dto.GoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.service.UpdateGoal(id, &domain.Goal{
		Name:        req.Name,

		TargetValue: req.TargetValue,
		Type:        req.Type,
		Ticker:      req.Ticker,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, dto.GoalFromDomain(updated))
}

func (h *GoalHandler) DeleteGoal(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteGoal(id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.Status(http.StatusNoContent)
}
