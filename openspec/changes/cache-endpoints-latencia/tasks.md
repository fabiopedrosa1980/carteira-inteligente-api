## 1. Pacote de cache in-memory

- [x] 1.1 Criar store TTL thread-safe (`sync.RWMutex` + map) com `Get/Set(key, value, ttl)`, expiração preguiçosa por TTL e `FlushPrefix(prefix)` / `Flush()`
- [x] 1.2 Suportar namespaces/baldes independentes: volátil, cotação, catálogo/externo
- [x] 1.3 Testes unitários do store: hit/miss, expiração por TTL, flush por prefixo e flush total

## 2. Middleware de resposta

- [x] 2.1 Implementar `CacheResponse(store, ttl)` com wrapper de `gin.ResponseWriter` que captura status + body
- [x] 2.2 Compor a chave a partir de método + path + querystring; incluir `userID` quando presente no contexto (rotas autenticadas)
- [x] 2.3 Em hit, devolver body cacheado sem executar o handler; em miss, executar e gravar apenas respostas 2xx
- [x] 2.4 Testes do middleware: segunda chamada serve do cache; usuários diferentes não compartilham entrada

## 3. Cache de cotação (camada Yahoo)

- [x] 3.1 Introduzir cache por ticker (TTL ~30–60s) na camada `fetchYahoo`/`fetchYahooQuote`
- [x] 3.2 Compartilhar o cache entre `quote_handler.GetQuote` e `transaction_handler.respondPositions`
- [x] 3.3 Garantir que cotação **não** é invalidada por mutações (só expira por TTL)
- [x] 3.4 Testes: segunda cotação do mesmo ticker dentro do TTL não chama o Yahoo

## 4. Aplicar cache de leitura nas rotas

- [x] 4.1 Balde volátil (TTL ~60s) em `GET /transactions`, `/allocation`, `/goals`, `/stocks`, `/stocks/:id`, `/stocks/:id/dividends`, `/dividends/monthly`
- [x] 4.2 Balde catálogo (TTL ~24h) em `GET /assets/:ticker` e `/assets/search`
- [x] 4.3 Balde externo (TTL curto) em `GET /search` (autocomplete Yahoo)
- [x] 4.4 Não cachear `GET /transactions/acoes|fiis|etfs` no nível de resposta (cotação ao vivo via cache de ticker)

## 5. Invalidação nas mutações

- [x] 5.1 Flush por usuário do balde volátil em `POST /transactions` e `POST /transactions/import`
- [x] 5.2 Flush por usuário em `PUT /transactions/:id`, `DELETE /transactions/:id` e `DELETE /transactions` (todos)
- [x] 5.3 Flush por usuário em `PUT /allocation` e nas escritas de `goals`
- [x] 5.4 Flush (global) nas escritas de `stocks` e `dividends` (chave não escopada por usuário)
- [x] 5.5 Garantir que nenhuma mutação de lançamento/importação toca o balde catálogo (`/assets/*`)
- [x] 5.6 Manter renovação do catálogo apenas em `POST /admin/catalog/refresh` e `StartCatalogSync`

## 6. Wiring e validação

- [x] 6.1 Instanciar os baldes no `SetupRouter` (volátil/catálogo/busca) e o cache de cotação no pacote handler — sem alterar a assinatura do router
- [x] 6.2 `go build ./...` sem erros
- [x] 6.3 `go vet ./...` sem apontamentos
- [x] 6.4 `go test ./...` sem regressões
- [x] 6.5 Validar manualmente: 2ª chamada de endpoint volátil é servida do cache; criar lançamento invalida; `/assets/:ticker` permanece cacheado após o lançamento

## 7. Commit e deploy

- [x] 7.1 Commit no padrão Conventional Commits e push para `main`
- [ ] 7.2 Confirmar deploy na Render e medir a redução de latência em `acoes/fiis/etfs`
