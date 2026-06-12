package router

import (
	"carteira-inteligente-api/internal/adapters/http/handler"
	"carteira-inteligente-api/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(stockHandler *handler.StockHandler, dividendHandler *handler.DividendHandler, transactionHandler *handler.TransactionHandler, quoteHandler *handler.QuoteHandler, goalHandler *handler.GoalHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

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
		v1.GET("/quote/:ticker", quoteHandler.GetQuote)

		goals := v1.Group("/goals")
		{
			goals.GET("", goalHandler.ListGoals)
			goals.POST("", goalHandler.CreateGoal)
			goals.PUT("/:id", goalHandler.UpdateGoal)
			goals.DELETE("/:id", goalHandler.DeleteGoal)
		}

		transactions := v1.Group("/transactions")
		transactions.Use(middleware.AuthRequired())
		{
			transactions.GET("/acoes", transactionHandler.GetAcoes)
			transactions.POST("", transactionHandler.CreateTransaction)
			transactions.GET("", transactionHandler.ListTransactions)
			transactions.DELETE("/:id", transactionHandler.DeleteTransaction)
		}
	}

	return r
}
