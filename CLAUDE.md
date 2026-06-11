# carteira-inteligente-api

API REST em Go para gestão de carteira financeira pessoal. Utiliza Gin (HTTP), GORM (ORM), SQLite e MongoDB.

## Arquitetura

Clean Architecture com as seguintes camadas em `internal/`:

- `domain/` — entidades e interfaces de repositório
- `application/` — casos de uso
- `adapters/` — handlers HTTP (controllers) e presenters
- `infrastructure/` — implementações de repositório, banco de dados, roteamento

Ponto de entrada: `cmd/api/`

## Stack

- Go 1.25
- Gin (HTTP framework)
- GORM + SQLite (persistência local)
- MongoDB (persistência alternativa)
- Deploy: Render.com (`render.yaml`) e Docker (`Dockerfile`)

## Comandos úteis

```bash
go run ./cmd/api          # Inicia o servidor
go build ./...            # Compila o projeto
go test ./...             # Executa os testes
go vet ./...              # Verifica erros estáticos
```

## Regra obrigatória: commit após cada alteração

**Após qualquer alteração no código, faça commit e push imediatamente.**

### Fluxo obrigatório

1. Implemente a alteração
2. Verifique que o projeto compila: `go build ./...`
3. Execute `go vet ./...` para checar erros estáticos
4. Faça o commit com mensagem no padrão Conventional Commits
5. Faça push para o GitHub: `git push origin main`

### Padrão de mensagem de commit

Use o prefixo adequado:

| Prefixo | Quando usar |
|---------|-------------|
| `feat:` | Nova funcionalidade |
| `fix:` | Correção de bug |
| `refactor:` | Refatoração sem mudança de comportamento |
| `chore:` | Configuração, dependências, arquivos de infra |
| `docs:` | Documentação |
| `test:` | Testes |

Exemplo:
```
feat: adiciona endpoint de resumo mensal de transações
```

### Exemplo de sequência completa

```bash
go build ./...
go vet ./...
git add <arquivos alterados>
git commit -m "feat: descrição da alteração"
git push origin main
```

> Nunca use `git add -A` ou `git add .` para evitar commitar arquivos sensíveis ou binários acidentalmente. Adicione arquivos específicos pelo nome.
