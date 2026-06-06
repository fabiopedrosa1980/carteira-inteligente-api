# Carteira Inteligente API

REST API for managing a stock portfolio with historical dividend data.

## Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| HTTP Framework | [Gin](https://github.com/gin-gonic/gin) v1.12 |
| ORM | [GORM](https://gorm.io) v1.31 |
| Database | SQLite (in-memory by default) |
| Architecture | Clean Architecture (domain / application / adapters / infrastructure) |
| AI Assistant | [Claude Sonnet 4.6](https://claude.ai/code) (Anthropic) — code generation, historical data seeding and documentation |

## Architecture

```
cmd/api/            → entrypoint
internal/
  domain/           → entities and repository interfaces
  application/      → use cases (services)
  adapters/http/    → handlers, DTOs, router (Gin)
  infrastructure/   → GORM persistence + seed
pkg/middleware/     → CORS and logger
scripts/            → helper scripts (seed via API)
```

## Running

```bash
go run ./cmd/api
```

The API starts on port `8080` by default. To change it, set the `PORT` environment variable.

```bash
PORT=9000 go run ./cmd/api
```

## Tests

```bash
go test ./...
```

---

## Endpoints

Base URL: `http://localhost:8080/api/v1`

---

### Stocks

#### `POST /stocks`

Creates a new stock.

**Body**
```json
{
  "ticker": "PETR4",
  "name": "Petrobras",
  "sector": "Petróleo & Gás",
  "score": 9,
  "current_price": 37.90,
  "daily_change": -0.5,
  "dy": 15.3
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `ticker` | string | Yes | Stock ticker symbol |
| `name` | string | Yes | Company name |
| `sector` | string | No | Market sector |
| `score` | float | No | Quality score (0–10) |
| `current_price` | float | Yes | Current price (> 0) |
| `daily_change` | float | No | Daily price change (%) |
| `dy` | float | No | Dividend yield (%) |

**Response** `201 Created`
```json
{
  "id": 1,
  "ticker": "PETR4",
  "name": "Petrobras",
  "sector": "Petróleo & Gás",
  "score": 9,
  "current_price": 37.90,
  "daily_change": -0.5,
  "dy": 15.3,
  "created_at": "2024-01-01T00:00:00-03:00",
  "updated_at": "2024-01-01T00:00:00-03:00"
}
```

---

#### `GET /stocks`

Lists all stocks with optional filters.

**Query params**

| Param | Type | Description |
|---|---|---|
| `sector` | string | Filter by sector (e.g. `Bancário`) |
| `sort` | string | Sort by `score`, `daily_change` or `dy` (descending) |

**Response** `200 OK` — array of stocks.

---

#### `GET /stocks/:id`

Returns a single stock by ID.

**Response** `200 OK` — stock object.

---

#### `PUT /stocks/:id`

Updates an existing stock.

**Body** — same fields as `POST /stocks` (`ticker`, `name`, `sector`, `score`, `current_price`, `daily_change`, `dy`).

**Response** `200 OK` — updated stock object.

---

#### `DELETE /stocks/:id`

Removes a stock.

**Response** `204 No Content`

---

### Dividends

#### `POST /stocks/:id/dividends`

Records a dividend payment for a stock.

**Body**
```json
{
  "amount": 1.29,
  "month": 5,
  "year": 2024,
  "type": "dividendo",
  "ex_date": "2024-05-10",
  "pay_date": "2024-05-25"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `amount` | float | Yes | Amount per share (> 0) |
| `month` | int | Yes | Payment month (1–12) |
| `year` | int | Yes | Payment year (≥ 2000) |
| `type` | string | Yes | `dividendo`, `jcp` or `rendimento` |
| `ex_date` | string | No | Ex-dividend date (`YYYY-MM-DD`) |
| `pay_date` | string | No | Payment date (`YYYY-MM-DD`) |

**Response** `201 Created`
```json
{
  "id": 42,
  "stock_id": 3,
  "amount": 1.29,
  "month": 5,
  "year": 2024,
  "type": "dividendo",
  "ex_date": "2024-05-10",
  "pay_date": "2024-05-25"
}
```

---

#### `GET /stocks/:id/dividends`

Lists all dividends for a stock.

**Query params**

| Param | Type | Description |
|---|---|---|
| `year` | int | Filter by year (e.g. `2023`) |

**Response** `200 OK` — array of dividends.

---

#### `GET /dividends/monthly?year=YYYY`

Returns a monthly aggregated summary of all dividends for a given year.

**Query params**

| Param | Type | Description |
|---|---|---|
| `year` | int | Year to query (defaults to current year) |

**Response** `200 OK`
```json
[
  {
    "month": 5,
    "month_name": "Maio",
    "tickers": ["PETR4", "CPFE3", "ITUB3"],
    "stock_count": 3,
    "avg_total": 0.82,
    "avg_yield": 2.1
  }
]
```

---

## Seed Data

On first startup with an empty database, the API automatically seeds **10 stocks** and their historical dividends from **2021 to 2025**:

| Ticker | Name | Sector |
|---|---|---|
| BBAS3 | Banco do Brasil | Bancário |
| BBSE3 | BB Seguridade | Seguros |
| PETR4 | Petrobras | Petróleo & Gás |
| ITUB3 | Itaú Unibanco | Bancário |
| BRAP4 | Bradespar | Mineração |
| CMIG4 | Cemig | Energia Elétrica |
| CPFE3 | CPFL Energia | Energia Elétrica |
| CSMG3 | Copasa | Saneamento |
| ISAE4 | Isa Cteep | Energia Elétrica |
| CXSE3 | Caixa Seguridade | Seguros |

Dividend amounts per year are derived from real historical data (source: [Investidor10](https://investidor10.com.br)).

To seed an already-running instance via the API, use the script:

```bash
./scripts/seed_dividends_2021_2024.sh http://localhost:8080
```

---

## Error Responses

| HTTP | Situation |
|---|---|
| `400` | Invalid body or parameter out of range |
| `404` | Stock not found |
| `409` | Duplicate ticker |
| `500` | Internal server error |

```json
{ "error": "descriptive message" }
```
