## Contexto

A API não tem cache. Deploy é instância única no Render (`plan: free`), o que torna um cache **in-memory** suficiente e elimina a necessidade de Redis. O banco local usa SQLite (inclusive `file::memory:?cache=shared`), então leituras de DB já são baratas; o gargalo real é a camada de cotação externa (Yahoo) no fan-out de `acoes/fiis/etfs`.

## Decisões

### 1. Três baldes de cache com políticas distintas

Em vez de um cache único, separamos por ciclo de vida do dado:

```
┌────────────────────┐  ┌────────────────────┐  ┌────────────────────┐
│ VOLÁTIL (por user) │  │ COTAÇÃO (por tick) │  │ CATÁLOGO / EXTERNO │
│ TTL ~60s           │  │ TTL ~30–60s        │  │ assets: TTL ~24h   │
│                    │  │                    │  │ search: TTL curto  │
│ transactions(list) │  │ /quote/:ticker     │  │ /assets/:ticker    │
│ allocation, goals  │  │ enriquecimento de  │  │ /assets/search     │
│ stocks, monthly    │  │ acoes/fiis/etfs    │  │ /search (Yahoo)    │
│                    │  │                    │  │                    │
│ key inclui userID  │  │ key = ticker       │  │ key = ticker/query │
│ ⟵ FLUSH por-user   │  │ expira só por TTL  │  │ ♻️ só refresh/sync │
│    em TODA escrita │  │                    │  │    do catálogo     │
└────────────────────┘  └────────────────────┘  └────────────────────┘
        │                                                  ▲
        │  POST /transactions, /import, PUT/DELETE         │ imune a
        │  allocation, goals, stocks, dividends            │ lançamentos
        └──────────────────────────────────────────────────┘
```

**Por quê:** cada balde tem um motivo diferente de expirar. O volátil reflete dados do usuário e precisa morrer na escrita; a cotação só precisa de uma janela curta para deixar de ser "tempo real puro"; o catálogo é estável e caro de recompor, então vive muito e ignora lançamentos.

### 2. Cotação com TTL curto entra no escopo (a despeito de "tempo real")

A regra original excluía endpoints que dependem de cotação em tempo real. Decidimos incluir um cache de cotação com TTL de 30–60s porque:

- É onde mora a latência real (fan-out de N chamadas Yahoo na tela da carteira).
- Uma janela de poucos segundos continua sendo "tempo real" para fins de acompanhamento de carteira.
- Reduz drasticamente o risco de rate-limit do Yahoo sob recarregamentos repetidos.

O cache é por ticker (`key = TICKER`), compartilhado entre `quote_handler` e `respondPositions`. **Não** é invalidado por mutações (cotação não muda quando o usuário lança uma compra).

> Alternativa considerada: deixar cotação 100% ao vivo. Rejeitada — daria ganho de latência apenas nos endpoints de DB, que já são rápidos, deixando o endpoint mais lento sem melhora.

### 3. Invalidação em TODA mutação, escopada por usuário

O flush do balde volátil dispara em qualquer escrita que altere as views cacheadas — não só `add`/`import`:

| Operação | Flush volátil | Toca catálogo |
|---|---|---|
| POST /transactions, /import | ✅ (do usuário) | não |
| PUT/DELETE /transactions, DELETE all | ✅ (do usuário) | não |
| PUT /allocation | ✅ (do usuário) | não |
| POST/PUT/DELETE /goals | ✅ (do usuário) | não |
| POST/PUT/DELETE /stocks, dividends | ✅ | não |
| POST /admin/catalog/refresh | não | ♻️ |
| StartCatalogSync (24h) | não | ♻️ |

**Por quê toda mutação:** invalidar só em add/import deixaria `/transactions` e `/allocation` servindo dados velhos após um PUT/DELETE.

**Por quê por usuário:** as rotas autenticadas embutem `userID` na chave de cache; o flush remove apenas o prefixo daquele usuário, sem derrubar o cache dos demais. Endpoints não autenticados (`stocks`, `dividends/monthly`) usam chave global e fazem flush global no respectivo write.

### 4. Implementação como middleware de resposta

```
        WRITE (POST/PUT/DELETE)                READ (GET cacheável)
              │                                      │
              ▼                                ┌─────┴──────┐
   cache.Volatile.FlushUser(userID)           │ middleware │ key = método
              │                                │ CacheResp. │   +path+query
   (só o balde volátil; catálogo intacto)     └─────┬──────┘   +userID
                                                hit │ miss
                                          devolve ◀─┘  └─▶ executa handler,
                                          body cacheado     captura body, grava
```

O middleware captura o corpo via um `ResponseWriter` wrapper, guardando status + bytes serializados. Hit devolve direto sem tocar o handler.

## Riscos e limitações

- **Cache efêmero**: reciclagem/hibernação da instância no Render zera o cache. Aceitável — afeta latência, não correção.
- **Escala horizontal**: invalidação in-memory não propaga entre instâncias. Hoje N/A (instância única). Documentado como ponto de evolução (Redis) caso o plano mude.
- **Coerência cotação vs. posição**: `acoes/fiis/etfs` mistura posição (DB, no balde volátil/recalculada) com cotação (balde de ticker). Os dois TTLs são curtos e independentes; divergências ficam na ordem de segundos, aceitável.

## Pontos em aberto

- TTLs exatos (60s volátil / 30–60s cotação / 24h catálogo) são valores iniciais; ajustáveis após medição.
- Localização do pacote (`pkg/cache` vs `internal/infrastructure/cache`) — seguir a convenção do repositório na implementação.
