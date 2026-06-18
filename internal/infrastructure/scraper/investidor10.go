package scraper

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"carteira-inteligente-api/internal/domain"

	"golang.org/x/net/html"
)

// ScrapedDividend holds raw dividend data extracted from Investidor10.
type ScrapedDividend struct {
	Type    domain.DividendType
	ExDate  string // DD/MM/YYYY
	PayDate string // DD/MM/YYYY
	Amount  float64
	Month   int
	Year    int
}

var httpClient = &http.Client{Timeout: 15 * time.Second}

// FetchDividends scrapes the dividend history for a B3 ação ticker.
// Mantido por compatibilidade; delega para FetchDividendsForType (ação).
func FetchDividends(ticker string, since time.Time) ([]ScrapedDividend, error) {
	return FetchDividendsForType(ticker, since, false)
}

// FetchDividendsForType scrapes the dividend history for a B3 ticker from
// investidor10.com.br and returns records with pay_date >= since. Quando fii é
// true usa a seção de FIIs (proventos = rendimento). On any fetch or parse
// error it returns nil, err; callers should treat this as a soft failure.
func FetchDividendsForType(ticker string, since time.Time, fii bool) ([]ScrapedDividend, error) {
	segment := "acoes"
	if fii {
		segment = "fiis"
	}
	url := fmt.Sprintf("https://investidor10.com.br/%s/%s/", segment, strings.ToLower(ticker))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("investidor10 fetch %s: %w", ticker, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("investidor10 %s: HTTP %d", ticker, resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("investidor10 HTML parse %s: %w", ticker, err)
	}

	rows := extractDividendRows(doc)
	var result []ScrapedDividend
	for _, row := range rows {
		d, ok := parseRow(row, since)
		if ok {
			// Para FIIs, proventos sem marcação de JCP são rendimentos.
			if fii && d.Type == domain.DividendTypeDividendo {
				d.Type = domain.DividendTypeRendimento
			}
			result = append(result, d)
		}
	}
	return result, nil
}

// extractDividendRows finds all <tr> rows inside the dividend history table.
// The table is located first by its id "table-dividends-history" (the markup
// rendered server-side by Investidor10) and, as a fallback, by a header row
// containing "tipo", "data" and "valor" columns.
func extractDividendRows(doc *html.Node) [][]string {
	dividendTable := findTableByID(doc, "table-dividends-history")

	if dividendTable == nil {
		// Fallback: locate by header content.
		var findTable func(*html.Node)
		findTable = func(n *html.Node) {
			if dividendTable != nil {
				return
			}
			if n.Type == html.ElementNode && n.Data == "table" {
				if tableHasDividendHeaders(n) {
					dividendTable = n
					return
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				findTable(c)
			}
		}
		findTable(doc)
	}

	if dividendTable == nil {
		log.Println("[investidor10] dividend table not found in HTML")
		return nil
	}

	var rows [][]string
	var extractRows func(*html.Node)
	extractRows = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			cells := extractCells(n)
			if len(cells) >= 4 {
				rows = append(rows, cells)
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractRows(c)
		}
	}
	extractRows(dividendTable)

	// Drop the header row (first row).
	if len(rows) > 0 {
		rows = rows[1:]
	}
	return rows
}

// findTableByID returns the first <table> element whose id attribute equals id,
// or nil if none is found.
func findTableByID(n *html.Node, id string) *html.Node {
	if n.Type == html.ElementNode && n.Data == "table" {
		for _, a := range n.Attr {
			if a.Key == "id" && a.Val == id {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findTableByID(c, id); found != nil {
			return found
		}
	}
	return nil
}

// tableHasDividendHeaders returns true if the table's first row contains
// headers that match the expected dividend columns.
func tableHasDividendHeaders(table *html.Node) bool {
	var firstRow *html.Node
	var find func(*html.Node)
	find = func(n *html.Node) {
		if firstRow != nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "tr" {
			firstRow = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			find(c)
		}
	}
	find(table)
	if firstRow == nil {
		return false
	}

	cells := extractCells(firstRow)
	if len(cells) < 4 {
		return false
	}
	row := strings.Join(cells, " ")
	row = strings.ToLower(row)
	return strings.Contains(row, "tipo") && strings.Contains(row, "data") && strings.Contains(row, "valor")
}

func extractCells(tr *html.Node) []string {
	var cells []string
	for c := tr.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
			cells = append(cells, strings.TrimSpace(nodeText(c)))
		}
	}
	return cells
}

func nodeText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(nodeText(c))
	}
	return sb.String()
}

// parseRow converts a table row ([]string) into a ScrapedDividend.
// Returns (zero, false) if the row is malformed or outside the date range.
func parseRow(cells []string, since time.Time) (ScrapedDividend, bool) {
	if len(cells) < 4 {
		return ScrapedDividend{}, false
	}

	rawType := strings.TrimSpace(cells[0])
	exDateStr := strings.TrimSpace(cells[1])
	payDateStr := strings.TrimSpace(cells[2])
	rawAmount := strings.TrimSpace(cells[3])

	// Map type.
	var dtype domain.DividendType
	upper := strings.ToUpper(rawType)
	if strings.Contains(upper, "JCP") || strings.Contains(upper, "JSCP") {
		dtype = domain.DividendTypeJCP
	} else {
		dtype = domain.DividendTypeDividendo
	}

	// Parse pay date (DD/MM/YYYY).
	payDate, err := time.Parse("02/01/2006", payDateStr)
	if err != nil {
		return ScrapedDividend{}, false
	}

	// Filter by since date.
	if payDate.Before(since) {
		return ScrapedDividend{}, false
	}

	// Parse ex date (best-effort; use pay date if empty/invalid).
	exDate, err := time.Parse("02/01/2006", exDateStr)
	if err != nil {
		exDate = payDate
	}

	// Parse Brazilian decimal amount (comma → dot).
	amtStr := strings.ReplaceAll(rawAmount, ".", "")
	amtStr = strings.ReplaceAll(amtStr, ",", ".")
	amount, err := strconv.ParseFloat(amtStr, 64)
	if err != nil || amount <= 0 {
		return ScrapedDividend{}, false
	}

	return ScrapedDividend{
		Type:    dtype,
		ExDate:  exDate.Format("2006-01-02"),
		PayDate: payDate.Format("2006-01-02"),
		Amount:  amount,
		Month:   int(payDate.Month()),
		Year:    payDate.Year(),
	}, true
}

// FetchProfile baixa a página do ativo no Investidor10 uma única vez e extrai
// tanto os Indicadores Fundamentalistas (`#table-indicators`) quanto as
// Informações sobre a empresa (`#table-indicators-company`), como pares
// rótulo/valor. Best-effort: em falha de rede/parse retorna (nil, nil, err);
// listas vazias quando uma seção não existe. Quando fii é true usa a seção de FIIs.
func FetchProfile(ticker string, fii bool) (indicators, companyInfo []domain.Indicator, err error) {
	segment := "acoes"
	if fii {
		segment = "fiis"
	}
	url := fmt.Sprintf("https://investidor10.com.br/%s/%s/", segment, strings.ToLower(ticker))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("investidor10 perfil %s: %w", ticker, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("investidor10 perfil %s: HTTP %d", ticker, resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("investidor10 perfil HTML parse %s: %w", ticker, err)
	}

	return extractIndicators(doc), extractCompanyInfo(doc), nil
}

// extractIndicators coleta os Indicadores Fundamentalistas do Investidor10. Eles
// ficam na seção `#table-indicators` (P/L, P/VP, ROE, DY, margens, EV/EBITDA,
// Dív./PL etc.) — não na seção "Informações sobre a empresa". Para cada `.cell`:
// o rótulo é o primeiro `<span>` da célula e o valor é o primeiro `<span>` dentro
// de `.value` (o número do próprio ativo, ignorando as médias de Setor/Subsetor).
func extractIndicators(doc *html.Node) []domain.Indicator {
	table := findFirstByID(doc, "table-indicators")
	if table == nil {
		return nil
	}

	var out []domain.Indicator
	seen := map[string]bool{}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && hasClass(n, "cell") {
			label := collapseSpaces(firstSpanText(n))
			value := ""
			if valueNode := findFirstByClass(n, "value"); valueNode != nil {
				value = collapseSpaces(firstSpanText(valueNode))
				if value == "" {
					value = collapseSpaces(nodeText(valueNode))
				}
			}
			if label != "" && value != "" && !seen[label] {
				seen[label] = true
				out = append(out, domain.Indicator{Label: label, Value: value})
			}
			return // não descer em células aninhadas
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(table)
	return out
}

// extractCompanyInfo coleta as "Informações sobre a empresa" do Investidor10,
// na seção `#table-indicators-company`. Cada `.cell` tem `.title` (rótulo) e
// `.value` com `.simple-value` (valor abreviado) + `.detail-value` (completo);
// usamos o `.simple-value` para evitar valor duplicado.
func extractCompanyInfo(doc *html.Node) []domain.Indicator {
	table := findFirstByID(doc, "table-indicators-company")
	if table == nil {
		return nil
	}

	var out []domain.Indicator
	seen := map[string]bool{}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && hasClass(n, "cell") {
			titleNode := findFirstByClass(n, "title")
			valueNode := findFirstByClass(n, "value")
			if titleNode != nil && valueNode != nil {
				label := collapseSpaces(nodeText(titleNode))
				valueSource := valueNode
				if simple := findFirstByClass(valueNode, "simple-value"); simple != nil {
					valueSource = simple
				}
				value := collapseSpaces(nodeText(valueSource))
				if label != "" && value != "" && !seen[label] {
					seen[label] = true
					out = append(out, domain.Indicator{Label: label, Value: value})
				}
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(table)
	return out
}

// findFirstByID retorna o primeiro elemento com o id informado.
func findFirstByID(n *html.Node, id string) *html.Node {
	if n.Type == html.ElementNode {
		for _, a := range n.Attr {
			if a.Key == "id" && a.Val == id {
				return n
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstByID(c, id); found != nil {
			return found
		}
	}
	return nil
}

// firstSpanText retorna o texto do primeiro elemento <span> encontrado (em
// profundidade), normalizado. Vazio se não houver <span>.
func firstSpanText(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "span" {
		return nodeText(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := firstSpanText(c); t != "" {
			return t
		}
	}
	return ""
}

// hasClass retorna true se o elemento possui a classe css informada.
func hasClass(n *html.Node, class string) bool {
	for _, a := range n.Attr {
		if a.Key == "class" {
			for _, c := range strings.Fields(a.Val) {
				if c == class {
					return true
				}
			}
		}
	}
	return false
}

// findFirstByClass retorna o primeiro descendente (ou o próprio nó) com a classe.
func findFirstByClass(n *html.Node, class string) *html.Node {
	if n.Type == html.ElementNode && hasClass(n, class) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findFirstByClass(c, class); found != nil {
			return found
		}
	}
	return nil
}

// collapseSpaces normaliza espaços em branco (incluindo quebras de linha) em um
// único espaço e remove as extremidades.
func collapseSpaces(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}
