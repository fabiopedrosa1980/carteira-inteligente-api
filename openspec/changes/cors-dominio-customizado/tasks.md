## 1. Alterar configuração de CORS

- [ ] 1.1 Em `pkg/middleware/cors.go`, adicionar `https://carteira-inteligente.com` e `https://www.carteira-inteligente.com` ao slice `AllowOrigins`, mantendo as origens existentes

## 2. Validar localmente

- [ ] 2.1 Rodar `go build ./...` sem erros
- [ ] 2.2 Rodar `go test ./...` sem regressões

## 3. Deploy e verificação

- [ ] 3.1 Commit e push das alterações
- [ ] 3.2 Confirmar deploy na Render (`carteira-inteligente-api.onrender.com`)
- [ ] 3.3 Validar preflight: `curl -i -X OPTIONS -H "Origin: https://carteira-inteligente.com" -H "Access-Control-Request-Method: GET" https://carteira-inteligente-api.onrender.com/api/v1/stocks` retorna `Access-Control-Allow-Origin: https://carteira-inteligente.com`
- [ ] 3.4 Repetir a validação para `https://www.carteira-inteligente.com`
