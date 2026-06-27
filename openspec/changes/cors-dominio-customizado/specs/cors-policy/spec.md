## ADDED Requirements

### Requirement: Origens de produção autorizadas no CORS

O middleware de CORS da API SHALL autorizar as origens do domínio de produção customizado `https://carteira-inteligente.com` e `https://www.carteira-inteligente.com`, além das origens já existentes de desenvolvimento e do domínio `.vercel.app`.

#### Scenario: Requisição a partir do domínio apex

- **WHEN** o navegador envia uma requisição preflight `OPTIONS` com `Origin: https://carteira-inteligente.com`
- **THEN** a resposta inclui o header `Access-Control-Allow-Origin: https://carteira-inteligente.com`

#### Scenario: Requisição a partir do subdomínio www

- **WHEN** o navegador envia uma requisição com `Origin: https://www.carteira-inteligente.com`
- **THEN** a resposta inclui o header `Access-Control-Allow-Origin: https://www.carteira-inteligente.com`

#### Scenario: Origens existentes preservadas

- **WHEN** o navegador envia uma requisição com `Origin: https://carteira-inteligente-eight.vercel.app`
- **THEN** a resposta continua incluindo o header `Access-Control-Allow-Origin` correspondente

#### Scenario: Origem não autorizada é rejeitada

- **WHEN** o navegador envia uma requisição com uma origem não listada (ex.: `https://exemplo-malicioso.com`)
- **THEN** a resposta NÃO inclui essa origem em `Access-Control-Allow-Origin`
