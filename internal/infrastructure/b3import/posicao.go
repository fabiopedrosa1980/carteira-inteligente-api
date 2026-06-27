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
	Sheet        string // "Acoes" | "ETF"
}

// importedSheets são as abas processadas. As demais abas do relatório de Posição
// (Empréstimos, Tesouro Direto) são deliberadamente ignoradas.
var importedSheets = []string{"Acoes", "ETF"}

// Cabeçalhos das colunas no relatório de Posição (mesmos nas abas Acoes e ETF).
const (
	headerTicker = "Código de Negociação"
	headerQty    = "Quantidade"
	headerPrice  = "Preço de Fechamento"
)

// ParsePosicao lê as abas Acoes e ETF de um .xlsx de Posição da B3 e devolve as
// posições válidas. Linhas sem ticker, de total/rodapé, ou com quantidade ≤ 0
// são ignoradas. Retorna erro apenas quando o arquivo não é um .xlsx legível.
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
		iQty := headerIndex(header, headerQty)
		iPrice := headerIndex(header, headerPrice)
		if iTicker < 0 || iQty < 0 {
			continue
		}
		for _, row := range rows[1:] {
			ticker := strings.ToUpper(strings.TrimSpace(cell(row, iTicker)))
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
			positions = append(positions, Position{
				Ticker:       ticker,
				Quantity:     qty,
				ClosingPrice: price,
				Sheet:        sheet,
			})
		}
	}
	return positions, nil
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
