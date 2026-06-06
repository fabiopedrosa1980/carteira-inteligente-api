package router

import (
	"carteira-inteligente-api/internal/adapters/http/handler"
	"carteira-inteligente-api/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(stockHandler *handler.StockHandler, dividendHandler *handler.DividendHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	v1 := r.Group("/api/v1")
	{
		stocks := v1.Group("/stocks")
		{
			stocks.POST("", stockHandler.CreateStock)
			stocks.GET("", stockHandler.ListStocks)
			stocks.GET("/:id", stockHandler.GetStock)
			stocks.PUT("/:id", stockHandler.UpdateStock)
			stocks.DELETE("/:id", stockHandler.DeleteStock)
			stocks.POST("/:id/dividends", dividendHandler.CreateDividend)
			stocks.GET("/:id/dividends", dividendHandler.ListDividends)
		}

		v1.GET("/dividends/monthly", dividendHandler.GetMonthlySummary)
	}

	return r
}
