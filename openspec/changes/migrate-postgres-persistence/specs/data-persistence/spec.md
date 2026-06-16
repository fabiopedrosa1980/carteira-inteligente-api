## ADDED Requirements

### Requirement: Persistência durável em PostgreSQL

O sistema SHALL persistir todas as entidades de domínio (ações, dividendos,
transações e metas) em um banco de dados PostgreSQL quando uma DSN de PostgreSQL
estiver configurada, de modo que os dados sobrevivam a reinícios do processo e a
novos deploys.

#### Scenario: Dados sobrevivem a reinício

- **WHEN** o serviço está configurado com `DATABASE_URL` apontando para o
  PostgreSQL e um registro é criado (por exemplo, uma transação) e o processo é
  reiniciado
- **THEN** o registro continua disponível através da API após o reinício

#### Scenario: Escrita confirmada no PostgreSQL

- **WHEN** uma operação de escrita (criação, atualização ou exclusão) é executada
  com o PostgreSQL configurado
- **THEN** a alteração é gravada no PostgreSQL e refletida em leituras
  subsequentes

### Requirement: Seleção de driver por configuração

O sistema SHALL selecionar o dialector do GORM na inicialização com base na
variável de ambiente `DATABASE_URL`: usar PostgreSQL quando ela estiver definida
e, caso contrário, recorrer ao SQLite em memória.

#### Scenario: DATABASE_URL definida

- **WHEN** a variável de ambiente `DATABASE_URL` está definida na inicialização
- **THEN** o sistema abre a conexão usando o driver PostgreSQL com essa DSN

#### Scenario: DATABASE_URL ausente

- **WHEN** a variável de ambiente `DATABASE_URL` não está definida
- **THEN** o sistema usa a DSN SQLite em memória como fallback (dev local e
  testes)

#### Scenario: Falha de conexão interrompe a inicialização

- **WHEN** `DATABASE_URL` está definida mas a conexão com o PostgreSQL falha
- **THEN** a inicialização do serviço falha com erro registrado, em vez de
  iniciar com persistência inválida

### Requirement: Migração automática de schema compatível com PostgreSQL

O sistema SHALL criar e manter o schema automaticamente na inicialização
(`AutoMigrate`) e as instruções de manutenção legadas (remoção de índice e de
colunas obsoletas em `goals`) SHALL ser compatíveis com PostgreSQL, sem impedir
o boot quando o objeto já não existir.

#### Scenario: Schema criado no primeiro boot

- **WHEN** o serviço inicia contra um banco PostgreSQL vazio
- **THEN** as tabelas de `Stock`, `Dividend`, `Transaction` e `Goal` são criadas
  pelo `AutoMigrate`

#### Scenario: Manutenção idempotente de schema

- **WHEN** as instruções de manutenção legadas (drop de índice/colunas) são
  executadas e o objeto-alvo não existe
- **THEN** a inicialização prossegue normalmente sem falhar (operação
  best-effort/idempotente)

### Requirement: Credenciais fornecidas por ambiente

O sistema SHALL obter as credenciais e a DSN do banco a partir de configuração de
ambiente, e essas credenciais NÃO SHALL ser fixadas (hardcoded) no código-fonte
nem commitadas no repositório.

#### Scenario: Conexão configurada via variável de ambiente

- **WHEN** o serviço é implantado no Render
- **THEN** a DSN do PostgreSQL é lida de `DATABASE_URL` definida no ambiente
  (dashboard/secret), e o repositório não contém a senha em texto claro
