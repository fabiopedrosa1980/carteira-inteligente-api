package b3import

import (
	"bytes"
	"testing"

	"github.com/xuri/excelize/v2"
)

// buildPosicaoXLSX monta um .xlsx no formato do relatório de Posição da B3 com
// as abas informadas. Cada aba recebe a linha de cabeçalho real e as linhas de
// dados passadas (já como strings de célula).
func buildPosicaoXLSX(t *testing.T, sheets map[string][][]string) *bytes.Reader {
	t.Helper()
	f := excelize.NewFile()
	header := []string{
		"Produto", "Instituição", "Conta", "Código de Negociação", "CNPJ da Empresa",
		"Código ISIN / Distribuição", "Tipo", "Escriturador", "Quantidade",
		"Quantidade Disponível", "Quantidade Indisponível", "Motivo",
		"Preço de Fechamento", "Valor Atualizado",
	}
	first := true
	for name, rows := range sheets {
		if first {
			f.SetSheetName("Sheet1", name)
			first = false
		} else {
			f.NewSheet(name)
		}
		all := append([][]string{header}, rows...)
		for r, row := range all {
			for c, val := range row {
				cell, _ := excelize.CoordinatesToCellName(c+1, r+1)
				f.SetCellStr(name, cell, val)
			}
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("WriteToBuffer: %v", err)
	}
	return bytes.NewReader(buf.Bytes())
}

// row monta uma linha de dados posicionando ticker, quantidade e preço nas
// colunas reais (índices 3, 8 e 12 do cabeçalho).
func row(ticker, qty, price string) []string {
	r := make([]string, 14)
	r[3] = ticker
	r[8] = qty
	r[12] = price
	return r
}

func TestParsePosicao_LeAcoesEEtf(t *testing.T) {
	src := buildPosicaoXLSX(t, map[string][][]string{
		"Acoes": {row("BBAS3", "183", "20.05"), row("PETR4", "121", "38.45")},
		"ETF":   {row("IVVB11", "6", "429.63")},
	})

	ps, err := ParsePosicao(src)
	if err != nil {
		t.Fatalf("ParsePosicao: %v", err)
	}
	if len(ps) != 3 {
		t.Fatalf("esperado 3 posições, obtido %d: %+v", len(ps), ps)
	}

	bySheet := map[string]int{}
	for _, p := range ps {
		bySheet[p.Sheet]++
	}
	if bySheet["Acoes"] != 2 || bySheet["ETF"] != 1 {
		t.Fatalf("contagem por aba inesperada: %v", bySheet)
	}
}

func TestParsePosicao_IgnoraLinhasInvalidasEAbasForaDeEscopo(t *testing.T) {
	src := buildPosicaoXLSX(t, map[string][][]string{
		"Acoes": {
			row("BBAS3", "183", "20.05"),
			row("", "10", "5.00"),     // sem ticker
			row("Total", "0", ""),     // linha de total
			row("ZERO3", "0", "10.0"), // quantidade zero
		},
		"Tesouro Direto": {row("LFT", "1", "100.0")}, // aba ignorada
	})

	ps, err := ParsePosicao(src)
	if err != nil {
		t.Fatalf("ParsePosicao: %v", err)
	}
	if len(ps) != 1 || ps[0].Ticker != "BBAS3" {
		t.Fatalf("esperado apenas BBAS3, obtido %+v", ps)
	}
}

func TestParsePosicao_NumeroPtBR(t *testing.T) {
	src := buildPosicaoXLSX(t, map[string][][]string{
		"Acoes": {row("ABCD3", "1234", "1.234,56")},
	})

	ps, err := ParsePosicao(src)
	if err != nil {
		t.Fatalf("ParsePosicao: %v", err)
	}
	if len(ps) != 1 {
		t.Fatalf("esperado 1 posição, obtido %d", len(ps))
	}
	// Preço em pt-BR ("1.234,56" → 1234.56); quantidade inteira simples.
	if ps[0].Quantity != 1234 || ps[0].ClosingPrice != 1234.56 {
		t.Fatalf("parsing pt-BR incorreto: qty=%g price=%g", ps[0].Quantity, ps[0].ClosingPrice)
	}
}

// buildEmprestimosXLSX monta um .xlsx com uma aba Empréstimos no layout real da
// B3 (ticker na coluna Produto, coluna Natureza). rows: {produto, natureza, qtd, preço}.
func buildEmprestimosXLSX(t *testing.T, rows [][4]string) *bytes.Reader {
	t.Helper()
	header := []string{
		"Produto", "Instituição", "Natureza", "Número de Contrato", "Modalidade",
		"OPA", "Liquidação antecipada", "Taxa", "Comissão", "Data de registro",
		"Data de vencimento", "Quantidade", "Preço de Fechamento", "Valor Atualizado",
	}
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", "Empréstimos")
	for c, h := range header {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		f.SetCellStr("Empréstimos", cell, h)
	}
	for i, r := range rows {
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
	return bytes.NewReader(buf.Bytes())
}

func TestParsePosicao_Emprestimos(t *testing.T) {
	src := buildEmprestimosXLSX(t, [][4]string{
		{"BBSE3 - BB SEGURIDADE PARTICIPAÇÕES S.A.", "Doador", "71", "38.87"},
		{"ITUB3 - ITAU UNIBANCO HOLDING S.A.", "Tomador", "150", "44.08"},
	})

	ps, err := ParsePosicao(src)
	if err != nil {
		t.Fatalf("ParsePosicao: %v", err)
	}
	// O parser não filtra natureza — devolve ambas as linhas com ticker do Produto.
	if len(ps) != 2 {
		t.Fatalf("esperado 2 posições, obtido %d: %+v", len(ps), ps)
	}
	if ps[0].Ticker != "BBSE3" || ps[0].Natureza != "Doador" || ps[0].Sheet != SheetEmprestimos {
		t.Fatalf("posição Doador inesperada: %+v", ps[0])
	}
	if ps[0].Quantity != 71 || ps[0].ClosingPrice != 38.87 {
		t.Fatalf("qtd/preço incorretos: %+v", ps[0])
	}
	if ps[1].Ticker != "ITUB3" || ps[1].Natureza != "Tomador" {
		t.Fatalf("posição Tomador inesperada: %+v", ps[1])
	}
}

func TestParsePosicao_ArquivoInvalido(t *testing.T) {
	if _, err := ParsePosicao(bytes.NewReader([]byte("não é um xlsx"))); err == nil {
		t.Fatal("esperado erro para arquivo inválido")
	}
}
