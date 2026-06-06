## Context

Projeto greenfield para uma API de portfólio de ações. Não há código legado. A solução deve ser construída com Go, usando Gin como framework HTTP, GORM como ORM e SQLite in-memory como banco de dados. A arquitetura deve seguir o padrão hexagonal (ports & adapters), com estrutura de pacotes enterprise-ready, facilitando substituição de dependências externas sem afetar o domínio.

## Goals / Non-Goals

**Goals:**
- API REST CRUD para gerenciamento de ações no portfólio
- Arquitetura hexagonal com separação clara entre domínio, aplicação, infraestrutura e adaptadores
- Banco de dados em memória — sem necessidade de setup externo para rodar o projeto
- Estrutura de pacotes profissional e escalável
- Validação de entrada nos handlers HTTP

**Non-Goals:**
- Autenticação e autorização (fora do escopo inicial)
- Integração com fontes externas de preços em tempo real
- Persistência durável em banco relacional ou NoSQL
- Paginação avançada ou full-text search
- Containerização ou deploy configuration

## Decisions

### 1. Arquitetura Hexagonal

**Decisão**: Usar ports & adapters com as seguintes camadas:
- `internal/domain` — entidades, value objects e interfaces (ports)
- `internal/application` — use cases, serviços de aplicação
- `internal/infrastructure` — implementações concretas (GORM repository)
- `internal/adapters/http` — handlers Gin, DTOs, routers

**Alternativas consideradas**:
- MVC simples: mais rápido de bootstrapar, mas mistura responsabilidades e dificulta testes unitários do domínio
- Clean Architecture (Uncle Bob): mais verbosa; hexagonal é suficiente para este escopo

**Rationale**: A separação de ports/adapters permite trocar o banco de dados por um real sem tocar no domínio, e testar use cases sem subir servidor HTTP.

### 2. Banco de Dados In-Memory com GORM + SQLite

**Decisão**: `gorm.io/driver/sqlite` com DSN `file::memory:?cache=shared`

**Alternativas consideradas**:
- Map em memória puro (sem GORM): mais simples, mas perde auto-migrate e query builder
- PostgreSQL local: requer setup externo, contradiz o requisito de in-memory
- `gorm.io/driver/sqlite` com arquivo: persiste dados mas complica reset em testes

**Rationale**: SQLite in-memory mantém a interface GORM (útil para futuras migrações de banco), sem dependência externa.

### 3. Estrutura de Pacotes

```
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── stock.go          # Entidade Stock
│   │   └── stock_repository.go # Port (interface)
│   ├── application/
│   │   └── stock_service.go  # Use cases
│   ├── infrastructure/
│   │   └── persistence/
│   │       └── gorm_stock_repository.go # Adapter (implementação GORM)
│   └── adapters/
│       └── http/
│           ├── handler/
│           │   └── stock_handler.go
│           ├── dto/
│           │   └── stock_dto.go
│           └── router/
│               └── router.go
├── go.mod
└── go.sum
```

### 4. Modelo de Domínio

Campo `Nota` (rating) como `float64` com range 0–10. `Variacao` (today's change) como `float64` representando percentual (ex: -1.5 = -1.5%). `Setor` como string livre (sem enum) para flexibilidade.

### 5. Convenções de API

- Base path: `/api/v1/stocks`
- Formato: JSON em todas as requests/responses
- IDs: autoincrement uint gerado pelo GORM
- Erros: `{"error": "<mensagem>"}` com HTTP status adequado

## Risks / Trade-offs

- **Dados perdidos ao reiniciar** → Aceitável para este escopo; in-memory é requisito explícito. Mitigação futura: trocar driver SQLite por PostgreSQL sem alterar a camada de domínio.
- **SQLite thread safety in-memory** → Usar `cache=shared` e uma única instância de `*gorm.DB` compartilhada via injeção de dependência. Não abrir múltiplas conexões independentes.
- **Sem validação de ticker duplicado** → A constraint unique no banco rejeita duplicatas; o handler deve capturar o erro GORM e retornar 409 Conflict.
