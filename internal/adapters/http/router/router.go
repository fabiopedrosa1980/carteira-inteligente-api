package router

import (
	"time"

	"carteira-inteligente-api/internal/adapters/http/handler"
	"carteira-inteligente-api/internal/infrastructure/cache"
	"carteira-inteligente-api/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// TTLs dos baldes de cache. Volátil curto (dados do usuário, invalidados em toda
// escrita); catálogo longo (resolução ticker→ativo, estável); busca externa
// curto (autocomplete do Yahoo).
const (
	volatileTTL = 60 * time.Second
	catalogTTL  = 24 * time.Hour
	searchTTL   = 5 * time.Minute
)

func SetupRouter(stockHandler *handler.StockHandler, dividendHandler *handler.DividendHandler, transactionHandler *handler.TransactionHandler, quoteHandler *handler.QuoteHandler, goalHandler *handler.GoalHandler, searchHandler *handler.SearchHandler, allocationHandler *handler.AllocationHandler, assetHandler *handler.AssetHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// Três baldes de cache in-memory com políticas distintas. Instância única no
	// Render torna o cache em memória suficiente (sem Redis).
	volatile := cache.New() // invalidado em toda mutação, escopado por usuário
	catalog := cache.New()  // ticker→ativo; imune a lançamentos, só TTL/refresh
	searchCache := cache.New()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		stocks := v1.Group("/stocks")
		// GETs cacheados no balde volátil; escritas (incl. dividendos do stock)
		// invalidam o cache público.
		stocks.Use(middleware.CacheResponse(volatile, volatileTTL), middleware.InvalidateOnWrite(volatile))
		{
			stocks.POST("", stockHandler.CreateStock)
			stocks.GET("", stockHandler.ListStocks)
			stocks.GET("/:id", stockHandler.GetStock)
			stocks.PUT("/:id", stockHandler.UpdateStock)
			stocks.DELETE("/:id", stockHandler.DeleteStock)
			stocks.POST("/:id/dividends", dividendHandler.CreateDividend)
			stocks.GET("/:id/dividends", dividendHandler.ListDividends)
		}

		v1.GET("/dividends/monthly", middleware.CacheResponse(volatile, volatileTTL), dividendHandler.GetMonthlySummary)
		// /quote/:ticker NÃO é cacheado no nível de resposta: a cotação ao vivo
		// passa pelo cache por ticker (TTL curto) dentro do handler.
		v1.GET("/quote/:ticker", quoteHandler.GetQuote)
		v1.GET("/search", middleware.CacheResponse(searchCache, searchTTL), searchHandler.Search)

		// Catálogo da B3 (b3_assets): resolução/busca local, sem site externo.
		// Cache persistente (TTL longo) e imune a lançamentos/importação.
		// A rota estática /assets/search é registrada antes do parâmetro /:ticker.
		assets := v1.Group("/assets")
		assets.Use(middleware.CacheResponse(catalog, catalogTTL))
		{
			assets.GET("/search", assetHandler.SearchAssets)
			assets.GET("/:ticker", assetHandler.GetAsset)
		}
		// Acionamento sob demanda da ingestão (protegido por autenticação).
		v1.POST("/admin/catalog/refresh", middleware.AuthRequired(), assetHandler.RefreshCatalog)

		goals := v1.Group("/goals")
		goals.Use(middleware.AuthRequired())
		goals.Use(middleware.CacheResponse(volatile, volatileTTL), middleware.InvalidateOnWrite(volatile))
		{
			goals.GET("", goalHandler.ListGoals)
			goals.POST("", goalHandler.CreateGoal)
			goals.PUT("/:id", goalHandler.UpdateGoal)
			goals.DELETE("/:id", goalHandler.DeleteGoal)
		}

		transactions := v1.Group("/transactions")
		transactions.Use(middleware.AuthRequired())
		// Invalida o cache volátil do usuário em toda escrita. CacheResponse é
		// aplicado só na lista (GET ""): acoes/fiis/etfs dependem de cotação ao
		// vivo e não são cacheados no nível de resposta.
		transactions.Use(middleware.InvalidateOnWrite(volatile))
		{
			transactions.GET("/acoes", transactionHandler.GetAcoes)
			transactions.GET("/fiis", transactionHandler.GetFiis)
			transactions.GET("/etfs", transactionHandler.GetEtfs)
			transactions.POST("", transactionHandler.CreateTransaction)
			transactions.POST("/import", transactionHandler.ImportTransactions)
			transactions.GET("", middleware.CacheResponse(volatile, volatileTTL), transactionHandler.ListTransactions)
			transactions.PUT("/:id", transactionHandler.UpdateTransaction)
			transactions.DELETE("", transactionHandler.DeleteAllTransactions)
			transactions.DELETE("/:id", transactionHandler.DeleteTransaction)
		}

		allocation := v1.Group("/allocation")
		allocation.Use(middleware.AuthRequired())
		allocation.Use(middleware.CacheResponse(volatile, volatileTTL), middleware.InvalidateOnWrite(volatile))
		{
			allocation.GET("", allocationHandler.GetAllocation)
			allocation.PUT("", allocationHandler.PutAllocation)
		}
	}

	return r
}
