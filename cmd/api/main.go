package main

import (
	"log"
	"os"

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
	if err := persistence.SeedIfEmpty(db); err != nil {
		log.Fatalf("failed to seed database: %v", err)
	}

	stockRepo := persistence.NewGormStockRepository(db)
	stockService := application.NewStockService(stockRepo)
	stockHandler := handler.NewStockHandler(stockService)

	dividendRepo := persistence.NewGormDividendRepository(db)
	dividendService := application.NewDividendService(dividendRepo, stockRepo)
	dividendHandler := handler.NewDividendHandler(dividendService)

	transactionRepo := persistence.NewGormTransactionRepository(db)
	transactionService := application.NewTransactionService(transactionRepo)
	transactionHandler := handler.NewTransactionHandler(transactionService)

	quoteHandler := handler.NewQuoteHandler()

	goalRepo := persistence.NewGormGoalRepository(db)
	goalService := application.NewGoalService(goalRepo)
	goalHandler := handler.NewGoalHandler(goalService)

	r := router.SetupRouter(stockHandler, dividendHandler, transactionHandler, quoteHandler, goalHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
