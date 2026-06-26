package handler

import (
	"errors"
	"net/http"
	"strings"

	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"

	"github.com/gin-gonic/gin"
)

type AssetHandler struct {
	svc application.AssetUseCase
}

func NewAssetHandler(svc application.AssetUseCase) *AssetHandler {
	return &AssetHandler{svc: svc}
}

// AssetResponse é o ativo resolvido pelo catálogo local (b3_assets).
type AssetResponse struct {
	Ticker string `json:"ticker"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Sector string `json:"sector"`
	Found  bool   `json:"found"`
}

// GetAsset resolve um ticker exclusivamente pelo catálogo local — sem consulta
// externa. Ausência do catálogo → found:false com HTTP 404.
func (h *AssetHandler) GetAsset(c *gin.Context) {
	ticker := strings.ToUpper(strings.TrimSpace(c.Param("ticker")))
	a, err := h.svc.GetByTicker(ticker)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, AssetResponse{Ticker: ticker, Found: false})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, AssetResponse{
		Ticker: a.Ticker,
		Name:   a.Name,
		Type:   a.Type,
		Sector: a.Sector,
		Found:  true,
	})
}

// SearchAssets retorna sugestões do catálogo local por prefixo de ticker/nome.
// Reaproveita o shape TickerSuggestion do autocomplete (ticker + name).
func (h *AssetHandler) SearchAssets(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if len(q) < 2 {
		c.JSON(http.StatusOK, []TickerSuggestion{})
		return
	}
	assets, err := h.svc.Search(q, 15)
	if err != nil {
		c.JSON(http.StatusOK, []TickerSuggestion{})
		return
	}
	out := make([]TickerSuggestion, 0, len(assets))
	for _, a := range assets {
		out = append(out, TickerSuggestion{Ticker: a.Ticker, Name: a.Name})
	}
	c.JSON(http.StatusOK, out)
}

// RefreshCatalog dispara a ingestão sob demanda do catálogo (acionamento admin).
func (h *AssetHandler) RefreshCatalog(c *gin.Context) {
	n, err := h.svc.Refresh()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"refreshed": n})
}
