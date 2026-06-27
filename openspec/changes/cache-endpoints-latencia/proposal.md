## Why

A tela principal da carteira (`GET /transactions/acoes|fiis|etfs`) faz um fan-out de N requisições HTTP ao Yahoo Finance por requisição (uma por ticker, em goroutines, com timeout de ~10s cada). É o endpoint mais usado e mais lento da API, e cada recarregamento repete todas as chamadas externas — gerando latência alta e risco de rate-limit no Yahoo. Os endpoints de leitura puramente de banco (lista de lançamentos, allocation, dividendos mensais) são rápidos, mas também são consultados com frequência e podem ser servidos de cache para aliviar o banco.

Não há nenhuma camada de cache hoje. Queremos reduzir a latência percebida sem comprometer a atualidade dos dados: cotação permanece praticamente em tempo real (janela de poucos segundos) e o catálogo da B3 (resolução ticker→ativo), que é estável, fica cacheado de forma persistente na própria API Go.

## What Changes

- **Novo pacote de cache in-memory** (TTL + thread-safe) com flush por namespace e por prefixo de chave. Deploy é instância única no Render (`plan: free`), então cache em memória é suficiente — sem Redis.
- **Middleware Gin `CacheResponse`** que serializa a resposta de endpoints `GET` selecionados, com chave incluindo `userID` para rotas autenticadas.
- **Três baldes de cache** com políticas distintas:
  - **Volátil (por usuário, TTL ~60s)**: `GET /transactions` (lista), `/allocation`, `/goals`, `/stocks`, `/stocks/:id`, `/stocks/:id/dividends`, `/dividends/monthly`. Invalidado em **toda mutação**.
  - **Cotação (por ticker, TTL ~30–60s)**: camada Yahoo compartilhada entre `GET /quote/:ticker` e o enriquecimento de `acoes/fiis/etfs`. Expira só por TTL (continua ~tempo real). Não é invalidado por mutações.
  - **Catálogo / busca externa (TTL longo)**: `GET /assets/:ticker` e `/assets/search` (ticker→ativo, catálogo b3_assets) com TTL ~24h; `GET /search` (autocomplete Yahoo) com TTL curto. **Imune** a lançamento/importação; só renova por TTL ou refresh do catálogo.
- **Invalidação por usuário em toda escrita** que afeta as views voláteis: `POST /transactions`, `POST /transactions/import`, `PUT/DELETE /transactions`, `DELETE` (todos), `PUT /allocation`, e escritas de `goals`, `stocks` e `dividends`. A chave inclui `userID`, então o flush remove apenas as entradas do usuário que escreveu.
- O catálogo (`/assets/*`) **nunca** é invalidado por lançamento/importação — só pelo `StartCatalogSync` (24h) ou `POST /admin/catalog/refresh`.

## Capabilities

### New Capabilities
- `response-cache`: Define quais endpoints são cacheados, suas políticas de TTL e as regras de invalidação por mutação, preservando o catálogo ticker→ativo entre lançamentos.

### Modified Capabilities
<!-- Nenhuma capability de domínio/negócio existente tem requisitos de comportamento alterados; o cache é transparente para o contrato dos endpoints. -->

## Impact

- **Código novo**: pacote de cache (`internal/infrastructure/cache` ou `pkg/cache`), middleware `CacheResponse`, e refactor da camada de cotação (`fetchYahoo`/`fetchYahooQuote`) para passar pelo cache de ticker.
- **Código alterado**: `router.go` (aplica o middleware nas rotas de leitura), handlers de escrita (hooks de invalidação), `quote_handler.go` e `transaction_handler.go` (cotação cacheada).
- **Deploy**: Render instância única — cache efêmero por design (reciclagem da instância só recompõe o cache; não afeta correção).
- **Limitação conhecida**: escala horizontal quebraria a invalidação in-memory entre instâncias; nesse cenário futuro seria necessário um backend compartilhado (ex.: Redis). Fora do escopo atual.
