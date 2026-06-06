# Carteira Inteligente API

API REST para gerenciamento de carteira de ações com histórico de dividendos.

## Stack

| Camada | Tecnologia |
|---|---|
| Linguagem | Go 1.25 |
| Framework HTTP | [Gin](https://github.com/gin-gonic/gin) v1.12 |
| ORM | [GORM](https://gorm.io) v1.31 |
| Banco de dados | SQLite (in-memory por padrão) |
| Arquitetura | Clean Architecture (domain / application / adapters / infrastructure) |
| IA assistente | [Claude Sonnet 4.6](https://claude.ai/code) (Anthropic) — geração de código, seed de dados históricos e documentação |

## Arquitetura

```
cmd/api/            → entrypoint
internal/
  domain/           → entidades e interfaces de repositório
  application/      → casos de uso (services)
  adapters/http/    → handlers, DTOs, router (Gin)
  infrastructure/   → persistência GORM + seed
pkg/middleware/     → CORS e logger
scripts/            → scripts auxiliares (seed via API)
```

## Executando

```bash
go run ./cmd/api
```

A API sobe na porta `8080` por padrão. Para alterar, defina a variável de ambiente `PORT`.

---

## Endpoints

Base URL: `http://localhost:8080/api/v1`

---

### Ações

#### `POST /stocks`

Cria uma nova ação.

**Body**
```json
{
  "ticker": "PETR4",
  "nome": "Petrobras",
  "setor": "Petróleo & Gás",
  "nota": 9,
  "preco_atual": 37.90,
  "variacao_hoje": -0.5,
  "dy": 15.3
}
```

**Resposta** `201 Created`
```json
{
  "id": 1,
  "ticker": "PETR4",
  "nome": "Petrobras",
  "setor": "Petróleo & Gás",
  "nota": 9,
  "preco_atual": 37.90,
  "variacao_hoje": -0.5,
  "dy": 15.3,
  "created_at": "2024-01-01T00:00:00-03:00",
  "updated_at": "2024-01-01T00:00:00-03:00"
}
```

---

#### `GET /stocks`

Lista todas as ações com filtros opcionais.

**Query params**

| Param | Tipo | Descrição |
|---|---|---|
| `setor` | string | Filtra por setor (ex: `Bancário`) |
| `sort` | string | Ordena por `nota`, `variacao` ou `dy` |

**Resposta** `200 OK` — array de ações.

---

#### `GET /stocks/:id`

Retorna uma ação pelo ID.

**Resposta** `200 OK` — objeto da ação.

---

#### `PUT /stocks/:id`

Atualiza uma ação existente.

**Body** — mesmos campos do `POST /stocks`.

**Resposta** `200 OK` — objeto atualizado.

---

#### `DELETE /stocks/:id`

Remove uma ação.

**Resposta** `204 No Content`

---

### Dividendos

#### `POST /stocks/:id/dividends`

Registra um dividendo para a ação.

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

| Campo | Tipo | Obrigatório | Descrição |
|---|---|---|---|
| `amount` | float | Sim | Valor por ação (> 0) |
| `month` | int | Sim | Mês do pagamento (1–12) |
| `year` | int | Sim | Ano do pagamento (≥ 2000) |
| `type` | string | Sim | `dividendo`, `jcp` ou `rendimento` |
| `ex_date` | string | Não | Data ex-dividendo (`YYYY-MM-DD`) |
| `pay_date` | string | Não | Data de pagamento (`YYYY-MM-DD`) |

**Resposta** `201 Created`
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

Lista todos os dividendos de uma ação.

**Query params**

| Param | Tipo | Descrição |
|---|---|---|
| `year` | int | Filtra por ano (ex: `2023`) |

**Resposta** `200 OK` — array de dividendos.

---

#### `GET /dividends/monthly?year=YYYY`

Retorna um resumo mensal agregado de todos os dividendos de um ano.

**Query params**

| Param | Tipo | Descrição |
|---|---|---|
| `year` | int | Ano consultado (padrão: ano atual) |

**Resposta** `200 OK`
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

## Dados de seed

Ao iniciar com banco vazio, a API popula automaticamente **10 ações** e seus dividendos históricos de **2021 a 2025**:

| Ticker | Nome | Setor |
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

Os valores de dividendo por ano foram derivados do histórico real de cada ação (fonte: [Investidor10](https://investidor10.com.br)).

Para popular um banco já existente via API, use o script:

```bash
./scripts/seed_dividends_2021_2024.sh http://localhost:8080
```

---

## Erros

| HTTP | Situação |
|---|---|
| `400` | Body inválido ou parâmetro fora do range |
| `404` | Ação não encontrada |
| `409` | Ticker duplicado |
| `500` | Erro interno |

```json
{ "error": "mensagem descritiva" }
```
