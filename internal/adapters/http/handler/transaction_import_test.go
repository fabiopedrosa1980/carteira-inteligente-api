package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"carteira-inteligente-api/internal/application"
	"carteira-inteligente-api/internal/domain"
	"carteira-inteligente-api/internal/infrastructure/persistence"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// buildImportXLSX gera um .xlsx no formato de Posição da B3 com as abas Acoes e
// ETF para uso no teste do endpoint de importação.
func buildImportXLSX(t *testing.T) []byte {
	t.Helper()
	f := excelize.NewFile()
	header := []string{
		"Produto", "Instituição", "Conta", "Código de Negociação", "CNPJ da Empresa",
		"Código ISIN / Distribuição", "Tipo", "Escriturador", "Quantidade",
		"Quantidade Disponível", "Quantidade Indisponível", "Motivo",
		"Preço de Fechamento", "Valor Atualizado",
	}
	write := func(sheet string, rows [][3]string) {
		f.SetCellStr(sheet, "A1", header[0])
		for c, h := range header {
			cell, _ := excelize.CoordinatesToCellName(c+1, 1)
			f.SetCellStr(sheet, cell, h)
		}
		for i, r := range rows {
			rowNum := i + 2
			set := func(col int, v string) {
				cell, _ := excelize.CoordinatesToCellName(col+1, rowNum)
				f.SetCellStr(sheet, cell, v)
			}
			set(3, r[0])  // ticker
			set(8, r[1])  // quantidade
			set(12, r[2]) // preço de fechamento
		}
	}
	f.SetSheetName("Sheet1", "Acoes")
	write("Acoes", [][3]string{{"BBAS3", "183", "20.05"}, {"PETR4", "121", "38.45"}})
	f.NewSheet("ETF")
	write("ETF", [][3]string{{"IVVB11", "6", "429.63"}})

	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("WriteToBuffer: %v", err)
	}
	return buf.Bytes()
}

func TestImportTransactions_SobrepoeERetornaResumo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := persistence.NewDBWithDSN("file:import_tx?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	txRepo := persistence.NewGormTransactionRepository(db)
	txSvc := application.NewTransactionService(txRepo)
	stockRepo := persistence.NewGormStockRepository(db)
	// stockSvc/dividendSvc nil → o efeito colateral em background é desativado
	// (sem rede); assetSvc nil → classificação por aba (ETF→ETFs, resto→Acoes).
	h := NewTransactionHandler(txSvc, stockRepo, nil, nil, nil)

	const userID = "user-1"

	// Lançamento manual preexistente, que deve ser sobreposto pela importação.
	if err := txSvc.Create(&domain.Transaction{
		UserID: userID, Ticker: "MANUAL3", AssetType: domain.AssetTypeAcoes,
		Quantity: 10, Price: 1.0, Date: time.Now(),
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Monta o upload multipart com nome de arquivo no padrão da B3.
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, _ := mw.CreateFormFile("file", "posicao-2026-06-27-02-37-19.xlsx")
	fw.Write(buildImportXLSX(t))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/import", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID)

	h.ImportTransactions(c)

	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtido %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Created struct{ Acoes, FIIs, ETFs int }
		Ignored []struct {
			Ticker string
			Reason string
		}
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v (%s)", err, w.Body.String())
	}
	if resp.Created.Acoes != 2 || resp.Created.ETFs != 1 || resp.Created.FIIs != 0 {
		t.Fatalf("resumo inesperado: %+v", resp.Created)
	}

	// A base deve conter apenas os 3 importados — MANUAL3 foi sobreposto.
	list, err := txSvc.List(userID, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("esperado 3 lançamentos após sobreposição, obtido %d", len(list))
	}
	for _, tx := range list {
		if tx.Ticker == "MANUAL3" {
			t.Fatal("lançamento manual não foi sobreposto")
		}
	}

	// A data deve vir do nome do arquivo (2026-06-27).
	for _, tx := range list {
		if tx.Date.Format("2006-01-02") != "2026-06-27" {
			t.Fatalf("data esperada do nome do arquivo, obtido %s", tx.Date.Format("2006-01-02"))
		}
		if tx.Ticker == "IVVB11" && tx.AssetType != domain.AssetTypeETFs {
			t.Fatalf("IVVB11 deveria ser ETFs, obtido %s", tx.AssetType)
		}
	}
}

func TestImportTransactions_ArquivoAusente(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewTransactionHandler(nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/import", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", "user-1")

	h.ImportTransactions(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400 sem arquivo, obtido %d", w.Code)
	}
}
