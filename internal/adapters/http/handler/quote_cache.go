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

// histQuoteTTL é longo: o fechamento de uma data passada é imutável; só usamos
// TTL para limitar o uso de memória entre reinícios.
const histQuoteTTL = 24 * time.Hour

// quoteTuple é o retorno de fetchYahooQuote guardado no cache.
type quoteTuple struct {
	price         float64
	changePercent float64
	name          string
	dividendYield float64
	high52        float64
	low52         float64
}

// cachedYahooQuote envolve fetchYahooQuote com o cache por ticker. Só guarda
// respostas válidas (preço > 0) para não fixar falhas transitórias do provedor.
func cachedYahooQuote(ticker string) (price, changePercent float64, name string, dividendYield, high52, low52 float64) {
	key := "quote:" + ticker
	if v, ok := quoteCache.Get(key); ok {
		if t, ok := v.(quoteTuple); ok {
			return t.price, t.changePercent, t.name, t.dividendYield, t.high52, t.low52
		}
	}
	price, changePercent, name, dividendYield, high52, low52 = fetchYahooQuote(ticker)
	if price > 0 {
		quoteCache.Set(key, quoteTuple{price, changePercent, name, dividendYield, high52, low52}, quoteTTL)
	}
	return price, changePercent, name, dividendYield, high52, low52
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

// cachedYahooOnDate envolve fetchYahooOnDate com o cache por ticker+data. Como o
// chamador só consulta datas passadas (fechamento imutável), usa TTL longo.
func cachedYahooOnDate(ticker, dateStr string) *QuoteResponse {
	key := "hist:" + ticker + ":" + dateStr
	if v, ok := quoteCache.Get(key); ok {
		if q, ok := v.(*QuoteResponse); ok {
			return q
		}
	}
	q := fetchYahooOnDate(ticker, dateStr)
	if q != nil {
		quoteCache.Set(key, q, histQuoteTTL)
	}
	return q
}
