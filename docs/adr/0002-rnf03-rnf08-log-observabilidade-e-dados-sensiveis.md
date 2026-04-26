# ADR 0002 — Avaliação de observabilidade e proteção de dados sensíveis

- Status: Accepted
- Date: 2026-04-26
- Deciders: Time Auto Repair Shop
- Tags: observabilidade, segurança, logs
- Issues relacionadas: #236, #223

## Contexto

Foi solicitada uma revisão do repositório para verificar o atendimento de dois requisitos não funcionais:

- #236 — logs estruturados (JSON) para criação de OS, transições de status e erros de validação, com timestamp, nível e contexto.
- #223 — garantir que CPF, CNPJ e senha não apareçam em logs da aplicação nem em mensagens de erro da API.

A base usa Go + Gin + Zap, com logger global em JSON e middleware HTTP para trilha de requisição.

## Estrutura atual relevante

- Logger central: `internal/pkg/logger/logger.go`
- Middleware de request logging: `internal/infra/http/middleware/logger.go`
- Fluxo principal de OS: `internal/services/service_order/service.go`
- Persistência de histórico de status da OS: `internal/infra/repository/service_order/postgres.go`
- Mapeamento de erros HTTP: `internal/pkg/web/errors.go`
- Handlers com debug de payload: `internal/infra/handler/*`
- Instruções de AI no repositório: `CLAUDE.md` e `.github/copilot-instructions.md`

## Avaliação dos cards

### Card #236 — Log estruturado em operações principais

Status atual: **atendido**.

Implementado:
- Eventos estruturados padronizados para OS no service layer (`internal/services/service_order/service.go`):
  - `service_order.created`
  - `service_order.status_transition`
  - `service_order.validation_failed`
- Log explícito de criação de OS com campos de contexto (`operation`, `entity`, `service_order_id`, `status`).
- Log explícito de transição de status para os fluxos:
  - `SendToCustomerApproval`
  - `Accept`
  - `Reject`
  - `Deliver`
  - `Cancel`
- Log de validação padronizado para cenários de entrada inválida nos fluxos principais de OS.

### Card #223 — Mascaramento de dados sensíveis

Status atual: **atendido**.

Implementado:
- Camada central de sanitização no logger (`internal/pkg/logger/logger.go`) aplicada em todos os níveis (`Info`, `Warn`, `Error`, `Fatal`, `Debug`).
- Redação por chave sensível (`password`, `senha`, `pwd`, `token`, `authorization`, `cpf`, `cnpj`, `document`, `documento`) com substituição por `[REDACTED]`.
- Redação de senha em DSN/connection string por padrão regex (`://user:pass@...` → `://user:[REDACTED]@...`).
- Remoção dos pontos pendentes com `zap.Any` em fluxos críticos mapeados:
  - `internal/infra/handler/user/handler.go`
  - `internal/infra/repository/user/postgres.go`
  - `internal/infra/handler/service_order/handler.go`

Sobre erro retornado pela API:
- Mantido o mapeamento de erros HTTP existente, sem exposição direta de senha/CPF/CNPJ nas mensagens padrão.

## Decisão

1. Considerar os gaps identificados nos cards #236 e #223 como implementados no código atual.
2. Formalizar os padrões adotados:
   - eventos estruturados de OS no service layer;
   - sanitização central de dados sensíveis no pacote de logger.
3. Manter a orientação de não registrar payload bruto com estruturas de domínio quando houver risco de dados sensíveis.

## Consequências

### Positivas

- Padronização de observabilidade no ciclo de vida de OS.
- Redução do risco de vazamento de dados sensíveis via logs.
- Maior rastreabilidade operacional para criação, validação e transição de status.

### Atenções contínuas

- Novos logs devem seguir o padrão de evento e campos já adotado.
- Mudanças futuras em handlers/repositories devem evitar exposição de payload bruto.

## Próximos passos recomendados

- Adicionar testes automatizados de logging/sanitização (asserções de não exposição de dados sensíveis).
- Expandir a revisão de `zap.Any` para módulos não críticos, quando aplicável, mantendo logs objetivos e seguros.
