package main

import (
	"log"
	"os"
	"time"

	"carteira-inteligente-api/internal/adapters/http/handler"
	"carteira-inteligente-api/internal/adapters/http/router"
	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/infrastructure/persistence"
)

func main() {
	db, err := persistence.NewDB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	stockRepo := persistence.NewGormStockRepository(db)
	stockService := application.NewStockService(stockRepo)

	dividendRepo := persistence.NewGormDividendRepository(db)
	dividendService := application.NewDividendService(dividendRepo, stockRepo)
	dividendHandler := handler.NewDividendHandler(dividendService)

	stockHandler := handler.NewStockHandler(stockService, dividendService)

	transactionRepo := persistence.NewGormTransactionRepository(db)
	transactionService := application.NewTransactionService(transactionRepo)
	transactionHandler := handler.NewTransactionHandler(transactionService, stockRepo, stockService, dividendService)

	quoteHandler := handler.NewQuoteHandler()
	searchHandler := handler.NewSearchHandler()

	goalRepo := persistence.NewGormGoalRepository(db)
	goalService := application.NewGoalService(goalRepo)
	goalHandler := handler.NewGoalHandler(goalService, transactionService, stockRepo)

	allocationRepo := persistence.NewGormAllocationRepository(db)
	allocationService := application.NewAllocationService(allocationRepo)
	allocationHandler := handler.NewAllocationHandler(allocationService)

	r := router.SetupRouter(stockHandler, dividendHandler, transactionHandler, quoteHandler, goalHandler, searchHandler, allocationHandler)

	// Mantém o histórico de dividendos atualizado: reimporta no startup e
	// periodicamente, capturando proventos publicados após o cadastro do stock.
	handler.StartDividendSync(stockService, dividendService, 12*time.Hour)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
