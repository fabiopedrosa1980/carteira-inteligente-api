package persistence_test

import (
	"testing"

	"carteira-inteligente-api/internal/domain"
	"carteira-inteligente-api/internal/infrastructure/persistence"
)

func setupDividendRepo(t *testing.T) (domain.DividendRepository, uint) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := persistence.NewDBWithDSN(dsn)
	if err != nil {
		t.Fatalf("setup db: %v", err)
	}
	stockRepo := persistence.NewGormStockRepository(db)
	stock := &domain.Stock{Ticker: "BBAS3", Name: "Banco do Brasil", CurrentPrice: 55.2}
	if err := stockRepo.Create(stock); err != nil {
		t.Fatalf("create stock: %v", err)
	}
	return persistence.NewGormDividendRepository(db), stock.ID
}

// Dois dividendos com mesmo tipo, mesma data com e mesmo pagamento, diferindo
// apenas no valor, DEVEM ser ambos persistidos (caso real BBAS3/PETR4).
func TestCreateIfNotExists_SameDatesDifferentAmount(t *testing.T) {
	repo, stockID := setupDividendRepo(t)

	a := &domain.Dividend{StockID: stockID, Amount: 0.00259905, Month: 3, Year: 2023, Type: domain.DividendTypeDividendo, ExDate: "2023-02-23", PayDate: "2023-03-03"}
	b := &domain.Dividend{StockID: stockID, Amount: 0.00632990, Month: 3, Year: 2023, Type: domain.DividendTypeDividendo, ExDate: "2023-02-23", PayDate: "2023-03-03"}

	if err := repo.CreateIfNotExists(a); err != nil {
		t.Fatalf("create a: %v", err)
	}
	if err := repo.CreateIfNotExists(b); err != nil {
		t.Fatalf("create b: %v", err)
	}

	got, err := repo.FindByStockID(stockID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("esperado 2 dividendos distintos persistidos, obtido %d", len(got))
	}
}

// Reimportar o mesmo registro (todos os campos da chave iguais) NAO deve duplicar.
func TestCreateIfNotExists_ExactDuplicateDeduped(t *testing.T) {
	repo, stockID := setupDividendRepo(t)

	d := &domain.Dividend{StockID: stockID, Amount: 0.39, Month: 1, Year: 2025, Type: domain.DividendTypeJCP, ExDate: "2025-01-10", PayDate: "2025-01-20"}

	if err := repo.CreateIfNotExists(d); err != nil {
		t.Fatalf("create 1: %v", err)
	}
	dup := *d
	dup.ID = 0
	if err := repo.CreateIfNotExists(&dup); err != nil {
		t.Fatalf("create dup: %v", err)
	}

	got, err := repo.FindByStockID(stockID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("esperado 1 dividendo (duplicata deduplicada), obtido %d", len(got))
	}
}
