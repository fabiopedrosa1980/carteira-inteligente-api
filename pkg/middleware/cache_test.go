package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"carteira-inteligente-api/internal/infrastructure/cache"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// withUser injeta um userID no contexto, simulando o AuthRequired.
func withUser(uid string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if uid != "" {
			c.Set("userID", uid)
		}
		c.Next()
	}
}

func TestCacheResponse_ServesSecondCallFromCache(t *testing.T) {
	store := cache.New()
	var calls int64

	r := gin.New()
	r.GET("/x", CacheResponse(store, time.Minute), func(c *gin.Context) {
		n := atomic.AddInt64(&calls, 1)
		c.JSON(http.StatusOK, gin.H{"n": n})
	})

	do := func() string {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/x", nil)
		r.ServeHTTP(w, req)
		return w.Body.String()
	}

	first := do()
	second := do()
	if first != second {
		t.Fatalf("segunda chamada deveria vir do cache: %q != %q", first, second)
	}
	if got := atomic.LoadInt64(&calls); got != 1 {
		t.Fatalf("handler deveria ter executado 1 vez, executou %d", got)
	}
}

func TestCacheResponse_IsolatesByUser(t *testing.T) {
	store := cache.New()
	var calls int64

	// Monta a cadeia manualmente por requisição para variar o userID.
	handler := func(uid string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		e := gin.New()
		e.GET("/x", withUser(uid), CacheResponse(store, time.Minute), func(c *gin.Context) {
			atomic.AddInt64(&calls, 1)
			c.JSON(http.StatusOK, gin.H{"u": uid})
		})
		req, _ := http.NewRequest(http.MethodGet, "/x", nil)
		e.ServeHTTP(w, req)
		return w
	}

	handler("user-1")
	handler("user-2")
	if got := atomic.LoadInt64(&calls); got != 2 {
		t.Fatalf("usuários distintos não devem compartilhar cache; esperado 2 execuções, obtido %d", got)
	}
}

func TestInvalidateOnWrite_FlushesUserScope(t *testing.T) {
	store := cache.New()
	var reads int64

	build := func() *gin.Engine {
		e := gin.New()
		g := e.Group("/r")
		g.Use(withUser("user-1"))
		g.Use(CacheResponse(store, time.Minute), InvalidateOnWrite(store))
		g.GET("", func(c *gin.Context) {
			atomic.AddInt64(&reads, 1)
			c.JSON(http.StatusOK, gin.H{"reads": atomic.LoadInt64(&reads)})
		})
		g.POST("", func(c *gin.Context) { c.JSON(http.StatusCreated, gin.H{"ok": true}) })
		return e
	}
	e := build()

	get := func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/r", nil)
		e.ServeHTTP(w, req)
	}
	post := func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/r", nil)
		e.ServeHTTP(w, req)
	}

	get() // executa handler (reads=1) e cacheia
	get() // servido do cache (reads continua 1)
	if got := atomic.LoadInt64(&reads); got != 1 {
		t.Fatalf("após cache esperado 1 execução, obtido %d", got)
	}
	post() // mutação invalida o cache do user-1
	get()  // deve reexecutar o handler (reads=2)
	if got := atomic.LoadInt64(&reads); got != 2 {
		t.Fatalf("após invalidação o handler deveria reexecutar; reads=%d", got)
	}
}
