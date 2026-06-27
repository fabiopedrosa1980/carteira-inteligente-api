package application_test

import (
	"testing"

	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"
)

type mockTxRepo struct {
	created []*domain.Transaction
	updated []*domain.Transaction
}

func (m *mockTxRepo) Create(t *domain.Transaction) error {
	m.created = append(m.created, t)
	return nil
}
func (m *mockTxRepo) Update(t *domain.Transaction) error {
	m.updated = append(m.updated, t)
	return nil
}
func (m *mockTxRepo) List(userID, ticker string) ([]*domain.Transaction, error) { return nil, nil }
func (m *mockTxRepo) GetByID(userID string, id uint) (*domain.Transaction, error) {
	return &domain.Transaction{ID: id, UserID: userID}, nil
}
func (m *mockTxRepo) Delete(userID string, id uint) error { return nil }
func (m *mockTxRepo) DeleteAll(userID string) error       { return nil }
func (m *mockTxRepo) ImportOverwrite(userID string, txs []*domain.Transaction) error {
	return nil
}
func (m *mockTxRepo) GetAcoesPositions(userID string) ([]*domain.AcoesPosition, error) {
	return nil, nil
}
func (m *mockTxRepo) GetFiisPositions(userID string) ([]*domain.AcoesPosition, error) {
	return nil, nil
}
func (m *mockTxRepo) GetEtfsPositions(userID string) ([]*domain.AcoesPosition, error) {
	return nil, nil
}
func (m *mockTxRepo) GetAllPositions(userID string) ([]*domain.AcoesPosition, error) {
	return nil, nil
}

func TestNormalizeTicker(t *testing.T) {
	cases := map[string]string{
		" petr4 ":  "PETR4",
		"petr4":    "PETR4",
		"PETR4":    "PETR4",
		"  mxrf11": "MXRF11",
	}
	for in, want := range cases {
		if got := domain.NormalizeTicker(in); got != want {
			t.Errorf("NormalizeTicker(%q) = %q, esperado %q", in, got, want)
		}
	}
}

func TestTransactionService_CreateNormalizaTicker(t *testing.T) {
	repo := &mockTxRepo{}
	svc := application.NewTransactionService(repo)

	if err := svc.Create(&domain.Transaction{UserID: "u1", Ticker: " petr4 "}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(repo.created) != 1 || repo.created[0].Ticker != "PETR4" {
		t.Fatalf("esperado ticker normalizado PETR4, obtido %+v", repo.created)
	}
}

func TestTransactionService_UpdateNormalizaTicker(t *testing.T) {
	repo := &mockTxRepo{}
	svc := application.NewTransactionService(repo)

	if err := svc.Update(&domain.Transaction{ID: 1, UserID: "u1", Ticker: "petr4"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if len(repo.updated) != 1 || repo.updated[0].Ticker != "PETR4" {
		t.Fatalf("esperado ticker normalizado PETR4, obtido %+v", repo.updated)
	}
}
