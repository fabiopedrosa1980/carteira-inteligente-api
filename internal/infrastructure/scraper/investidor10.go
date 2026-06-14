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

// FetchDividends scrapes the dividend history for a B3 ticker from
// investidor10.com.br and returns records with pay_date >= since.
// On any fetch or parse error it returns nil, err; callers should treat
// this as a soft failure (log and continue).
func FetchDividends(ticker string, since time.Time) ([]ScrapedDividend, error) {
	url := fmt.Sprintf("https://investidor10.com.br/acoes/%s/", strings.ToLower(ticker))
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
