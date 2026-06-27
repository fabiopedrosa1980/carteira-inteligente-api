## Why

O frontend está sendo migrado para o domínio próprio `carteira-inteligente.com` (com `www.carteira-inteligente.com` redirecionando). Hoje o middleware de CORS só autoriza a origem `https://carteira-inteligente-eight.vercel.app`, então chamadas à API a partir do novo domínio serão bloqueadas pelo navegador. Precisamos liberar as novas origens para a aplicação continuar funcionando.

## What Changes

- Adicionar `https://carteira-inteligente.com` e `https://www.carteira-inteligente.com` à lista `AllowOrigins` do middleware de CORS (`pkg/middleware/cors.go`).
- Manter as origens já existentes (localhost de dev e o domínio `.vercel.app`) como fallback.

## Capabilities

### New Capabilities
- `cors-policy`: Define quais origens HTTP podem consumir a API, incluindo agora o domínio de produção customizado.

### Modified Capabilities
<!-- Nenhuma capability de domínio/negócio existente tem requisitos alterados. -->

## Impact

- **Código**: `pkg/middleware/cors.go` (lista `AllowOrigins`).
- **Deploy**: Render (`carteira-inteligente-api.onrender.com`) — requer novo deploy para entrar em vigor.
- **Frontend**: Desbloqueia chamadas de `https://carteira-inteligente.com` ao backend.
- **Segurança**: Apenas adiciona origens confiáveis e específicas; nenhuma origem curinga (`*`) é introduzida.
