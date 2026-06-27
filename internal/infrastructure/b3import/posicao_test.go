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

func TestParsePosicao_ArquivoInvalido(t *testing.T) {
	if _, err := ParsePosicao(bytes.NewReader([]byte("não é um xlsx"))); err == nil {
		t.Fatal("esperado erro para arquivo inválido")
	}
}
