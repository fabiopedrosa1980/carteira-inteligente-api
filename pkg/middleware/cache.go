package middleware

import (
	"bytes"
	"net/http"
	"time"

	"carteira-inteligente-api/internal/infrastructure/cache"

	"github.com/gin-gonic/gin"
)

// cachedResponse é a resposta serializada guardada no cache.
type cachedResponse struct {
	status      int
	contentType string
	body        []byte
}

// bodyCapture intercepta o corpo escrito pelo handler para que o middleware
// possa guardá-lo no cache, sem deixar de escrever na resposta real.
type bodyCapture struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w *bodyCapture) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyCapture) WriteString(s string) (int, error) {
	w.buf.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// cacheKey compõe a chave do cache. O userID vem primeiro (vazio em rotas
// públicas) para permitir invalidação por prefixo de usuário; em seguida método,
// caminho e querystring distinguem cada recurso e seus filtros.
func cacheKey(c *gin.Context) string {
	return c.GetString("userID") + "|" + c.Request.Method + ":" +
		c.Request.URL.Path + "?" + c.Request.URL.RawQuery
}

// CacheResponse serve respostas GET a partir de um cache em memória com o TTL
// informado. Em hit, devolve o corpo guardado sem executar o handler; em miss,
// executa o handler e guarda apenas respostas 2xx. Requisições não-GET passam
// direto.
func CacheResponse(store *cache.Store, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		key := cacheKey(c)
		if v, ok := store.Get(key); ok {
			if cr, ok := v.(cachedResponse); ok {
				c.Data(cr.status, cr.contentType, cr.body)
				c.Abort()
				return
			}
		}

		bw := &bodyCapture{ResponseWriter: c.Writer, buf: &bytes.Buffer{}}
		c.Writer = bw
		c.Next()

		status := c.Writer.Status()
		if status >= 200 && status < 300 {
			store.Set(key, cachedResponse{
				status:      status,
				contentType: c.Writer.Header().Get("Content-Type"),
				body:        bw.buf.Bytes(),
			}, ttl)
		}
	}
}

// InvalidateOnWrite invalida o cache volátil após uma mutação bem-sucedida
// (POST/PUT/PATCH/DELETE 2xx). O flush é escopado pelo prefixo do usuário
// autenticado; em rotas públicas (sem userID) limpa o prefixo público "|".
// Aplicar este middleware apenas a grupos cujos dados estão no balde volátil —
// nunca ao catálogo de ativos.
func InvalidateOnWrite(store *cache.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			if status := c.Writer.Status(); status < 200 || status >= 300 {
				return
			}
			if uid := c.GetString("userID"); uid != "" {
				store.FlushPrefix(uid + "|")
			} else {
				store.FlushPrefix("|")
			}
		}
	}
}
