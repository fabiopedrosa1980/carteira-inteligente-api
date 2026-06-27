// Package b3import parseia planilhas exportadas da área do investidor da B3.
package b3import

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Position é uma linha de posição lida do relatório de Posição da B3.
type Position struct {
	Ticker       string
	Quantity     float64
	ClosingPrice float64
	Sheet        string // "Acoes" | "ETF" | "Empréstimos"
	// Natureza só é preenchida na aba Empréstimos ("Doador" | "Tomador"); vazia
	// nas demais abas.
	Natureza string
}

// SheetEmprestimos é a aba de empréstimo de ativos. Nela o ticker vem da coluna
// "Produto" (não há "Código de Negociação") e a "Natureza" distingue o investidor
// doador (continua dono) do tomador.
const SheetEmprestimos = "Empréstimos"

// importedSheets são as abas processadas. A aba "Tesouro Direto" é deliberadamente
// ignorada.
var importedSheets = []string{"Acoes", "ETF", SheetEmprestimos}

// Cabeçalhos das colunas no relatório de Posição.
const (
	headerTicker   = "Código de Negociação"
	headerProduto  = "Produto"
	headerQty      = "Quantidade"
	headerPrice    = "Preço de Fechamento"
	headerNatureza = "Natureza"
)

// ParsePosicao lê as abas Acoes, ETF e Empréstimos de um .xlsx de Posição da B3
// e devolve as posições válidas. Linhas sem ticker, de total/rodapé, ou com
// quantidade ≤ 0 são ignoradas. O filtro por natureza (Doador/Tomador) é
// responsabilidade do chamador. Retorna erro apenas quando o arquivo não é um
// .xlsx legível.
func ParsePosicao(r io.Reader) ([]Position, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("arquivo .xlsx inválido ou ilegível")
	}
	defer f.Close()

	exists := map[string]bool{}
	for _, s := range f.GetSheetList() {
		exists[s] = true
	}

	var positions []Position
	for _, sheet := range importedSheets {
		if !exists[sheet] {
			continue
		}
		rows, err := f.GetRows(sheet)
		if err != nil || len(rows) == 0 {
			continue
		}
		header := rows[0]
		iTicker := headerIndex(header, headerTicker)
		iProduto := headerIndex(header, headerProduto)
		iQty := headerIndex(header, headerQty)
		iPrice := headerIndex(header, headerPrice)
		iNatureza := headerIndex(header, headerNatureza)
		// Precisa de quantidade e de alguma fonte de ticker (coluna própria ou
		// "Produto", usado na aba Empréstimos).
		if iQty < 0 || (iTicker < 0 && iProduto < 0) {
			continue
		}
		for _, row := range rows[1:] {
			ticker := ""
			if iTicker >= 0 {
				ticker = strings.ToUpper(strings.TrimSpace(cell(row, iTicker)))
			} else {
				ticker = tickerFromProduto(cell(row, iProduto))
			}
			if ticker == "" || strings.EqualFold(ticker, "Total") {
				continue
			}
			qty := parseNumber(cell(row, iQty))
			if qty <= 0 {
				continue
			}
			price := 0.0
			if iPrice >= 0 {
				price = parseNumber(cell(row, iPrice))
			}
			natureza := ""
			if iNatureza >= 0 {
				natureza = strings.TrimSpace(cell(row, iNatureza))
			}
			positions = append(positions, Position{
				Ticker:       ticker,
				Quantity:     qty,
				ClosingPrice: price,
				Sheet:        sheet,
				Natureza:     natureza,
			})
		}
	}
	return positions, nil
}

// tickerFromProduto extrai o ticker da coluna "Produto" (formato
// "BBSE3 - BB SEGURIDADE ..."), usado na aba Empréstimos, que não tem coluna
// "Código de Negociação".
func tickerFromProduto(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if i := strings.Index(s, "-"); i >= 0 {
		s = s[:i]
	}
	return strings.ToUpper(strings.TrimSpace(s))
}

// headerIndex acha o índice da coluna cujo cabeçalho corresponde exatamente a
// name (case-insensitive, ignorando espaços), evitando casar "Quantidade" com
// "Quantidade Disponível".
func headerIndex(header []string, name string) int {
	for i, h := range header {
		if strings.EqualFold(strings.TrimSpace(h), name) {
			return i
		}
	}
	return -1
}

func cell(row []string, idx int) string {
	if idx < 0 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

// parseNumber interpreta valores numéricos tolerando formatação pt-BR
// ("1.234,56") e en-US ("1234.56"). Retorna 0 para vazio, "-" ou inválido.
func parseNumber(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}
	s = strings.ReplaceAll(s, " ", "")
	hasComma := strings.Contains(s, ",")
	hasDot := strings.Contains(s, ".")
	switch {
	case hasComma && hasDot:
		// pt-BR: ponto de milhar, vírgula decimal.
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	case hasComma:
		s = strings.ReplaceAll(s, ",", ".")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}
