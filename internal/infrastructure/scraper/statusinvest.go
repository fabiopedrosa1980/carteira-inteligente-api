package scraper

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"carteira-inteligente-api/internal/domain"
)

// FetchIndicators obtém indicadores fundamentais (P/L, P/VP, DY, ROE, Payout)
// de um ativo no Status Invest. É best-effort: em qualquer falha de rede/parse
// retorna (nil, err) e o chamador deve tratar como ausência de indicadores
// (não é erro fatal). Tenta primeiro a página de ações e, em seguida, a de FIIs.
func FetchIndicators(ticker string) (*domain.StockIndicators, error) {
	for _, segment := range []string{"acoes", "fundos-imobiliarios"} {
		ind, err := fetchIndicatorsFrom(segment, ticker)
		if err == nil && ind != nil && !indicatorsEmpty(ind) {
			return ind, nil
		}
	}
	return nil, fmt.Errorf("statusinvest: indicadores não encontrados para %s", ticker)
}

func fetchIndicatorsFrom(segment, ticker string) (*domain.StockIndicators, error) {
	url := fmt.Sprintf("https://statusinvest.com.br/%s/%s", segment, strings.ToLower(ticker))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("statusinvest fetch %s: %w", ticker, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("statusinvest %s: HTTP %d", ticker, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseIndicators(string(body)), nil
}

// labelValueRe captura, para um rótulo de indicador, o primeiro valor numérico
// renderizado em um elemento com classe "value" logo após o título. O HTML do
// Status Invest expõe cada indicador como um título seguido de <strong class="value">N</strong>.
func labelValueRe(label string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(label)
	return regexp.MustCompile(`(?is)>\s*` + escaped + `\s*<.*?class="[^"]*\bvalue\b[^"]*"[^>]*>\s*([-0-9.,]+)\s*<`)
}

func parseIndicators(htmlBody string) *domain.StockIndicators {
	ind := &domain.StockIndicators{}
	ind.PL = matchFloat(htmlBody, "P/L")
	ind.PVP = matchFloat(htmlBody, "P/VP")
	ind.DY = matchFloat(htmlBody, "DY")
	ind.ROE = matchFloat(htmlBody, "ROE")
	ind.Payout = matchFloat(htmlBody, "PAYOUT")
	if indicatorsEmpty(ind) {
		return nil
	}
	return ind
}

func matchFloat(htmlBody, label string) *float64 {
	m := labelValueRe(label).FindStringSubmatch(htmlBody)
	if len(m) < 2 {
		return nil
	}
	return parseBRNumber(m[1])
}

// parseBRNumber converte um número no formato brasileiro ("1.234,56", "10,52",
// "12,3") para float64. Retorna nil quando não é um número válido.
func parseBRNumber(raw string) *float64 {
	s := strings.TrimSpace(raw)
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", ".")
	if s == "" || s == "-" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

func indicatorsEmpty(i *domain.StockIndicators) bool {
	return i == nil || (i.PL == nil && i.PVP == nil && i.DY == nil && i.ROE == nil && i.Payout == nil)
}
