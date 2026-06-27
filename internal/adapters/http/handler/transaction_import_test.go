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

	// Aba Empréstimos (layout próprio: ticker em Produto, coluna Natureza).
	// BBSE3 Doador (incluído), ITUB3 Tomador (ignorado), PETR4 Doador (somado
	// à posição de PETR4 da aba Acoes → 121 + 50 = 171).
	empHeader := []string{
		"Produto", "Instituição", "Natureza", "Número de Contrato", "Modalidade",
		"OPA", "Liquidação antecipada", "Taxa", "Comissão", "Data de registro",
		"Data de vencimento", "Quantidade", "Preço de Fechamento", "Valor Atualizado",
	}
	f.NewSheet("Empréstimos")
	for c, h := range empHeader {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		f.SetCellStr("Empréstimos", cell, h)
	}
	empRows := [][4]string{
		{"BBSE3 - BB SEGURIDADE PARTICIPAÇÕES S.A.", "Doador", "71", "38.87"},
		{"ITUB3 - ITAU UNIBANCO HOLDING S.A.", "Tomador", "150", "44.08"},
		{"PETR4 - PETROLEO BRASILEIRO S.A. PETROBRAS", "Doador", "50", "38.45"},
	}
	for i, r := range empRows {
		rowNum := i + 2
		set := func(col int, v string) {
			cell, _ := excelize.CoordinatesToCellName(col+1, rowNum)
			f.SetCellStr("Empréstimos", cell, v)
		}
		set(0, r[0])  // Produto
		set(2, r[1])  // Natureza
		set(11, r[2]) // Quantidade
		set(12, r[3]) // Preço de Fechamento
	}

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
	// Acoes: BBAS3, PETR4, BBSE3 (Doador, classificado como Acoes sem catálogo) = 3;
	// ETFs: IVVB11 = 1; ITUB3 (Tomador) é ignorado e não conta.
	if resp.Created.Acoes != 3 || resp.Created.ETFs != 1 || resp.Created.FIIs != 0 {
		t.Fatalf("resumo inesperado: %+v", resp.Created)
	}

	// A base deve conter apenas os 4 importados — MANUAL3 sobreposto, ITUB3 (Tomador) fora.
	list, err := txSvc.List(userID, "")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 4 {
		t.Fatalf("esperado 4 lançamentos após sobreposição, obtido %d", len(list))
	}

	byTicker := map[string]*domain.Transaction{}
	for _, tx := range list {
		if tx.Ticker == "MANUAL3" {
			t.Fatal("lançamento manual não foi sobreposto")
		}
		if tx.Ticker == "ITUB3" {
			t.Fatal("ITUB3 (Tomador) não deveria ter sido importado")
		}
		byTicker[tx.Ticker] = tx
		// A data deve vir do nome do arquivo (2026-06-27).
		if tx.Date.Format("2006-01-02") != "2026-06-27" {
			t.Fatalf("data esperada do nome do arquivo, obtido %s", tx.Date.Format("2006-01-02"))
		}
	}

	if byTicker["IVVB11"] == nil || byTicker["IVVB11"].AssetType != domain.AssetTypeETFs {
		t.Fatalf("IVVB11 deveria ser ETFs: %+v", byTicker["IVVB11"])
	}
	if byTicker["BBSE3"] == nil || byTicker["BBSE3"].Quantity != 71 {
		t.Fatalf("BBSE3 (Doador) deveria ter 71 cotas: %+v", byTicker["BBSE3"])
	}
	// PETR4: 121 (Acoes) + 50 (Empréstimos Doador) = 171.
	if byTicker["PETR4"] == nil || byTicker["PETR4"].Quantity != 171 {
		t.Fatalf("PETR4 deveria somar 171 cotas (121 + 50): %+v", byTicker["PETR4"])
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
