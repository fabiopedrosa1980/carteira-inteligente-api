## ADDED Requirements

### Requirement: Configuração de alocação por usuário

A API SHALL persistir a configuração de alocação **por usuário** (escopada pelo `userID` do token autenticado): alvos por classe (Ações, FIIs, ETFs) e limite de concentração por ativo. Cada usuário MUST ter no máximo uma configuração (singleton). Os endpoints MUST exigir autenticação (`AuthRequired`).

#### Scenario: Usuário sem configuração recebe defaults

- **WHEN** um usuário autenticado faz `GET /api/v1/allocation` e ainda não salvou configuração
- **THEN** a API responde 200 com os defaults (Ações 50, FIIs 40, ETFs 10, limite 20)

#### Scenario: Gravar e ler de volta

- **WHEN** o usuário faz `PUT /api/v1/allocation` com alvos e limite válidos
- **THEN** a API persiste (upsert) e responde com a configuração salva
- **AND** um `GET` subsequente do mesmo usuário retorna os valores gravados

#### Scenario: Isolamento entre usuários

- **WHEN** dois usuários distintos gravam configurações diferentes
- **THEN** cada `GET` retorna apenas a configuração do próprio usuário

#### Scenario: Requer autenticação

- **WHEN** uma requisição a `/api/v1/allocation` chega sem `Authorization: Bearer <token>` válido
- **THEN** a API responde 401

### Requirement: Contrato do endpoint de alocação

O corpo de request/response SHALL usar o shape `{ "targets": { "Acoes": number, "FIIs": number, "ETFs": number }, "concentrationLimit": number }`, compatível com o `ApiAllocation` do frontend. Valores MUST ser percentuais numéricos.

#### Scenario: Shape do PUT

- **WHEN** o cliente envia `PUT /api/v1/allocation` com `targets` e `concentrationLimit`
- **THEN** a API aceita o corpo e retorna o mesmo shape com os valores salvos
