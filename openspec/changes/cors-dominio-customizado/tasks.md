## 1. Alterar configuração de CORS

- [x] 1.1 Em `pkg/middleware/cors.go`, adicionar `https://carteira-inteligente.com` e `https://www.carteira-inteligente.com` ao slice `AllowOrigins`, mantendo as origens existentes

## 2. Validar localmente

- [x] 2.1 Rodar `go build ./...` sem erros
- [x] 2.2 Rodar `go test ./...` sem regressões (49 testes, 9 pacotes)

## 3. Deploy e verificação

- [x] 3.1 Commit e push das alterações (commit ff40f56)
- [x] 3.2 Confirmar deploy na Render (`carteira-inteligente-api.onrender.com`)
- [x] 3.3 Validar preflight: apex retorna `Access-Control-Allow-Origin: https://carteira-inteligente.com` (HTTP 204)
- [x] 3.4 Repetir a validação para `https://www.carteira-inteligente.com` (HTTP 204; origem desconhecida → 403)
