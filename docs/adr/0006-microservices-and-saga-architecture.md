# ADR 0006 — Arquitetura de microsserviços com saga orquestrada

- Status: Accepted
- Contexto: Tech Challenge Fase 4 (SOAT)
- Relacionados: [ADR 0005](0005-monolith-becomes-gateway.md), [registro de decisões](../decisions.md), [spec](../fase4/SPEC.md)

## Contexto

A oficina passou a operar em várias filiais. O monólito Go concentrava ordens de serviço, clientes, estoque e autenticação num só processo e num só banco. O desafio pede serviços independentes, cada um com repositório, infraestrutura e banco próprios, e consistência entre as transações críticas (criação da OS, aprovação e pagamento) com rollback seguro em caso de falha.

## Decisão

1. **Separar em 6 serviços** por bounded context: Customer & Identity, Catalog, Billing, Execution, History & Reporting e Service Order. Cada um em repositório próprio, com banco próprio e pipeline próprio; nenhum serviço acessa o banco de outro.
2. **Coordenar o fluxo de aprovação com uma saga orquestrada**, cujo orquestrador mora no Service Order: orçamento → decisão do cliente → pagamento → execução, com compensações (cancelar orçamento, estornar pagamento, liberar estoque) na ordem inversa. A escolha foi por orquestração porque o fluxo é linear e de ordem fixa, o estado fica num lugar só (`saga_instance`) e a máquina de estados da OS já existia.
3. **Comunicar por mensageria (SNS + SQS)** para a saga e por REST interno apenas onde a resposta é imediata. Cada serviço publica por *outbox* transacional e consome com deduplicação por `message_id`.
4. **Usar o monólito como gateway** (reverse proxy): único ponto público, sem lógica de negócio e sem validar tokens.

## Consequências

- Consistência eventual: depois do `accept` a OS está `IN_PROGRESS` enquanto o pagamento é processado; falhas posteriores compensam e levam a OS a `REJECTED` ou `CANCELLED`.
- O Service Order guarda snapshots de dados de outros serviços para responder sem chamadas cruzadas.
- Operar seis serviços e uma saga exige correlação nos logs (`request_id`, `saga_id`) e métricas da saga.
- A lambda de login por CPF permanece, lendo o database do serviço Customer.

## Alternativas consideradas

- **Saga coreografada:** sem ponto central, mas o fluxo ficaria espalhado em seis serviços e as compensações seriam difíceis de acompanhar.
- **Menos serviços** (por exemplo, juntar Catalog ao Service Order): menos repositórios, mas mistura estoque com o ciclo da OS.
- **Outros brokers (RabbitMQ):** descartado por exigência do time (SNS/SQS) e porque o Lab não permite IAM por pod.
