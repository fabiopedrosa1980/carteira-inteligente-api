package handler

import (
	"time"

	"carteira-inteligente-api/internal/infrastructure/cache"
)

// quoteCache guarda cotações por ticker por uma janela curta (quoteTTL). É
// compartilhado entre /quote/:ticker e o enriquecimento de acoes/fiis/etfs,
// eliminando o fan-out repetido ao Yahoo a cada recarregamento. A janela curta
// mantém o dado praticamente em tempo real. Não é invalidado por mutações —
// expira apenas por TTL (uma cotação não muda quando o usuário lança uma compra).
var quoteCache = cache.New()

const quoteTTL = 45 * time.Second

// quoteTuple é o retorno de fetchYahooQuote guardado no cache.
type quoteTuple struct {
	price         float64
	changePercent float64
	name          string
	dividendYield float64
}

// cachedYahooQuote envolve fetchYahooQuote com o cache por ticker. Só guarda
// respostas válidas (preço > 0) para não fixar falhas transitórias do provedor.
func cachedYahooQuote(ticker string) (price, changePercent float64, name string, dividendYield float64) {
	key := "quote:" + ticker
	if v, ok := quoteCache.Get(key); ok {
		if t, ok := v.(quoteTuple); ok {
			return t.price, t.changePercent, t.name, t.dividendYield
		}
	}
	price, changePercent, name, dividendYield = fetchYahooQuote(ticker)
	if price > 0 {
		quoteCache.Set(key, quoteTuple{price, changePercent, name, dividendYield}, quoteTTL)
	}
	return price, changePercent, name, dividendYield
}

// cachedYahoo envolve fetchYahoo (cotação ao vivo de /quote) com o cache por
// ticker. Só guarda respostas encontradas.
func cachedYahoo(ticker string) *QuoteResponse {
	key := "chart:" + ticker
	if v, ok := quoteCache.Get(key); ok {
		if q, ok := v.(*QuoteResponse); ok {
			return q
		}
	}
	q := fetchYahoo(ticker)
	if q != nil {
		quoteCache.Set(key, q, quoteTTL)
	}
	return q
}
