package scraper

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"carteira-inteligente-api/internal/domain"
)

// catalogSegments mapeia a seção do Investidor10 (e seu sitemap) para o tipo de
// ativo gravado no catálogo. A categoria vem da seção de origem — não de
// heurística por sufixo —, classificando corretamente units de ação terminadas
// em 11 (TAEE11/SANB11 em "acoes"), FIIs (MXRF11 em "fiis") e ETFs (BOVA11 em
// "etfs"). Os valores de tipo casam com o AssetType do frontend.
var catalogSegments = []struct {
	segment string
	typ     string
}{
	{"acoes", "Acoes"},
	{"fiis", "FIIs"},
	{"etfs", "ETFs"},
}

// sitemapURLSet decodifica o <urlset> dos sitemaps do Investidor10.
type sitemapURLSet struct {
	URLs []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

// tickerRe valida o slug da URL como ticker da B3 (ex.: TAEE11, BOVA11, PETR4),
// descartando páginas que não são de papel (categorias, slugs textuais).
var tickerRe = regexp.MustCompile(`^[A-Z0-9]{4,}[0-9]{1,2}$`)

// FetchB3Catalog baixa o catálogo completo da B3 a partir dos sitemaps por
// categoria do Investidor10 e retorna os ativos com o tipo dado pela seção.
// Best-effort por seção: falha em uma categoria não derruba as demais; só
// retorna erro se nenhum ativo for coletado.
func FetchB3Catalog() ([]domain.Asset, error) {
	var all []domain.Asset
	seen := map[string]bool{}
	var firstErr error

	for _, s := range catalogSegments {
		assets, err := fetchCatalogSegment(s.segment, s.typ)
		if err != nil {
			log.Printf("[catalog] %s: %v", s.segment, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		added := 0
		for _, a := range assets {
			if seen[a.Ticker] {
				continue
			}
			seen[a.Ticker] = true
			all = append(all, a)
			added++
		}
		log.Printf("[catalog] %s: %d ativos", s.segment, added)
	}

	if len(all) == 0 {
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, fmt.Errorf("catálogo vazio")
	}
	return all, nil
}

func fetchCatalogSegment(segment, typ string) ([]domain.Asset, error) {
	url := fmt.Sprintf("https://investidor10.com.br/sitemap-%s.xml", segment)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36")
	req.Header.Set("Accept", "application/xml,text/xml,*/*")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sitemap %s: %w", segment, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sitemap %s: HTTP %d", segment, resp.StatusCode)
	}

	var set sitemapURLSet
	if err := xml.NewDecoder(resp.Body).Decode(&set); err != nil {
		return nil, fmt.Errorf("sitemap %s parse: %w", segment, err)
	}

	prefix := "/" + segment + "/"
	var out []domain.Asset
	for _, u := range set.URLs {
		if ticker := tickerFromLoc(u.Loc, prefix); ticker != "" {
			out = append(out, domain.Asset{Ticker: ticker, Type: typ})
		}
	}
	return out, nil
}

// tickerFromLoc extrai e valida o ticker do slug da URL do Investidor10
// (.../acoes/taee11/ → TAEE11). Retorna "" quando o slug não é um ticker B3.
func tickerFromLoc(loc, prefix string) string {
	i := strings.Index(loc, prefix)
	if i < 0 {
		return ""
	}
	rest := strings.Trim(loc[i+len(prefix):], "/")
	if rest == "" || strings.Contains(rest, "/") {
		return ""
	}
	ticker := strings.ToUpper(rest)
	if !tickerRe.MatchString(ticker) {
		return ""
	}
	return ticker
}
