package handler

import (
	"errors"
	"net/http"
	"sync"

	"carteira-inteligente-api/internal/adapters/http/dto"
	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type GoalHandler struct {
	service     *application.GoalService
	txService   application.TransactionUseCase
	stockRepo   domain.StockRepository
}

func NewGoalHandler(service *application.GoalService, txService application.TransactionUseCase, stockRepo domain.StockRepository) *GoalHandler {
	return &GoalHandler{service: service, txService: txService, stockRepo: stockRepo}
}

func (h *GoalHandler) ListGoals(c *gin.Context) {
	userID := c.GetString("userID")
	goals, err := h.service.ListGoals(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, h.buildGoalResponses(userID, goals))
}

func (h *GoalHandler) buildGoalResponses(userID string, goals []domain.Goal) []dto.GoalResponse {
	positions, err := h.txService.GetAcoesPositions(userID)
	if err != nil || len(positions) == 0 {
		out := make([]dto.GoalResponse, len(goals))
		for i := range goals {
			out[i] = dto.GoalFromDomain(&goals[i], 0, 0)
		}
		return out
	}

	// Fetch live prices for all positions in parallel.
	priceMap := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, pos := range positions {
		wg.Add(1)
		go func(ticker string) {
			defer wg.Done()
			if q := fetchYahoo(ticker); q != nil {
				mu.Lock()
				priceMap[ticker] = q.Price
				mu.Unlock()
			}
		}(pos.Ticker)
	}
	wg.Wait()

	// Total patrimônio = sum(qty * currentPrice).
	patrimonio := 0.0
	for _, pos := range positions {
		price := pos.AvgPrice
		if p, ok := priceMap[pos.Ticker]; ok {
			price = p
		}
		patrimonio += pos.TotalQuantity * price
	}

	// DY from stocks table (stored as percentage, e.g. 8.5 for 8.5%).
	stocks, _ := h.stockRepo.FindAll(domain.StockQuery{})
	dyMap := make(map[string]float64)
	for _, s := range stocks {
		dyMap[s.Ticker] = s.DY / 100
	}

	// Estimated monthly income = sum(qty * price * dy / 12).
	rendaMensal := 0.0
	for _, pos := range positions {
		price := pos.AvgPrice
		if p, ok := priceMap[pos.Ticker]; ok {
			price = p
		}
		rendaMensal += pos.TotalQuantity * price * dyMap[pos.Ticker] / 12
	}

	// Average purchase price per ticker.
	avgMap := make(map[string]float64)
	for _, pos := range positions {
		avgMap[pos.Ticker] = pos.AvgPrice
	}

	out := make([]dto.GoalResponse, len(goals))
	for i, g := range goals {
		var currentValue float64
		switch g.Type {
		case "patrimonio":
			currentValue = patrimonio
		case "renda_mensal":
			currentValue = rendaMensal
		case "preco_medio":
			currentValue = avgMap[g.Ticker]
		}

		progressPercent := 0.0
		if g.TargetValue > 0 {
			progressPercent = (currentValue / g.TargetValue) * 100
			if progressPercent > 100 {
				progressPercent = 100
			}
		}

		out[i] = dto.GoalFromDomain(&g, currentValue, progressPercent)
	}
	return out
}

func (h *GoalHandler) CreateGoal(c *gin.Context) {
	userID := c.GetString("userID")
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
	if err := h.service.CreateGoal(userID, g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, dto.GoalFromDomain(g, 0, 0))
}

func (h *GoalHandler) UpdateGoal(c *gin.Context) {
	userID := c.GetString("userID")
	id := c.Param("id")
	var req dto.GoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.service.UpdateGoal(userID, id, &domain.Goal{
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
	c.JSON(http.StatusOK, dto.GoalFromDomain(updated, 0, 0))
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
