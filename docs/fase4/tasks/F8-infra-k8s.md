# Ficha F8 — `auto-repair-shop-infra-k8s` (mensageria, registry, nós)

Repo existente; fluxo do ADR 0001: `feature → develop → main` (**base dos PRs = `develop`**, que já existe; o default do GitHub é `main`). Ler `contracts/{resources,messaging}.md`, `terraform/shared/{ecr,variables}.tf`, `terraform/aws/eks.tf`, `.github/workflows/infra.yml`. Regra de ouro: **plano só adiciona**, exceto o ajuste de nós (T4), que deve ser inspecionado antes de qualquer apply.

## T1 — State `terraform/messaging/` (novo, sem remote state)
- `versions.tf`: backend `s3` `key = "messaging/terraform.tfstate"` (mesmo padrão parcial dos outros states), `aws ~> 6.0`.
- `variables.tf`: `region`, `project` (default `auto-repair-shop`), `aws_endpoint` (default `null`), `state_bucket` (**declarar** — o workflow repassa) e **`manage_iam` (declarar, sem uso)**: o workflow passa `-var manage_iam=…` a *todo* layer, e variável não declarada faz o Terraform falhar.
- `provider "aws"` com bloco `endpoints { sns, sqs }` + `skip_*` quando `aws_endpoint != null` (igual à ficha F7/T4).
- `locals.tf`: `environment` = workspace (`default`→`stg`); prefixo `"${project}-${environment}-"`; mapas:
  - `topics = ["saga-commands","so-events","billing-events","execution-events"]`
  - `queues = { billing-queue = {topic="saga-commands", filter={target=["billing"]}}, execution-queue = {…["execution"]}, catalog-queue = {…["catalog"]}, so-saga-replies = {topics=["billing-events","execution-events"]}, history-queue = {topics=["so-events","billing-events","execution-events"]} }` (modelar como lista de assinaturas `{queue, topic, filter?}`).
- `main.tf`:
  - `aws_sns_topic` por tópico; `aws_sqs_queue` por fila e `…-dlq`; fila principal com `redrive_policy` (`maxReceiveCount=5`), `visibility_timeout_seconds=60`, `message_retention_seconds=345600`, `receive_wait_time_seconds=20`; DLQ com retenção 14 dias.
  - `aws_sqs_queue_policy` por fila permitindo `sqs:SendMessage` ao(s) tópico(s) que a alimentam (`Condition ArnEquals aws:SourceArn`).
  - `aws_sns_topic_subscription` por assinatura: `protocol="sqs"`, `raw_message_delivery=true`, e para `saga-commands` `filter_policy = jsonencode({target=[…]})` com `filter_policy_scope = "MessageAttributes"`.
- `outputs.tf`: `topic_arns` (nome lógico→ARN), `queue_urls`, `queue_arns`, `dlq_arns`.
- Nomes físicos exatamente `"${project}-${environment}-<nome lógico>"` (contrato `resources.md`; os `docker.yml` dos serviços derivam URLs/ARNs por convenção).

## T2 — ECR por serviço (`terraform/shared/`)
- Em `variables.tf`: `service_repositories` (`list(string)`, default os 6: `auto-repair-shop-customer-service`, `…-catalog-service`, `…-billing-service`, `…-execution-service`, `…-history-service`, `…-service-order-service`).
- Em `ecr.tf`: `aws_ecr_repository.services` (`for_each = toset(var.service_repositories)`, `IMMUTABLE`, `scan_on_push`) + `aws_ecr_lifecycle_policy.services` (últimas 10 imagens), **copiando** as configurações do repo atual. **Não mexer** em `aws_ecr_repository.app` (gateway). Output `service_repository_urls`.
- Lab (`manage_iam=false`): só ECR é criado no `shared`; credenciais estáticas dos repos de serviço já fazem push (nada de OIDC/IAM a alterar). Registrar no README que em conta com IAM livre é preciso ampliar `github_repo`/trust das roles para os novos repos.

## T3 — Workflow (`.github/workflows/infra.yml`)
- `inputs.layer.options`: adicionar `messaging`.
- Job `validate` (PR): incluir `messaging` no `for d in …`.
- No passo Terraform do job `run`: `case` — `messaging) VARS+=(-var "state_bucket=$STATE_BUCKET") ;;` (o `manage_iam` já é adicionado para todos).
- **Novo job** `messaging-ministack` (PR, sem credenciais reais): `services: ministack` (porta 4566); `terraform -chdir=terraform/messaging init -backend=false`; `apply -auto-approve -var aws_endpoint=http://localhost:4566 -var manage_iam=false -var state_bucket=x`; conferir com `aws --endpoint-url … sns list-topics`/`sqs list-queues` que existem **4 tópicos e 10 filas (5 + 5 DLQs)**, e que `sns list-subscriptions` tem as 8 assinaturas esperadas (3 de `saga-commands`, 2 de `so-saga-replies`, 3 de `history-queue`).

## T4 — Nós com IMDS hop limit 2 (pods leem credenciais da `LabRole`)
`terraform/aws/eks.tf`, ramo Lab (`aws_eks_node_group.lab`): criar `aws_launch_template.lab_nodes` (`count = var.manage_iam ? 0 : 1`) com `metadata_options { http_endpoint = "enabled", http_tokens = "required", http_put_response_hop_limit = 2 }` e referenciar no node group (`launch_template { id = …, version = aws_launch_template.lab_nodes[0].latest_version }`; manter `instance_types` no node group). Ramo com módulo: `eks_managed_node_groups.default.metadata_options = { … }`.
⚠️ **Antes de qualquer apply**: `terraform plan` e inspecionar. Se o plano indicar **replace** do node group (destroy/create), **parar e avisar o usuário** — isso derruba os pods atuais do STG. Só seguir com aprovação explícita.

## T5 — Documentação
README (novo state `messaging`, ECR por serviço, hop limit, como validar no MiniStack); ADRs `0005-messaging-with-sns-sqs.md` (SNS fan-out + SQS por consumidor, DLQ, raw delivery, filter policy por atributo, Standard vs FIFO) e `0006-ecr-per-service.md`; `docs/` do gateway **sem mudanças** (API Gateway AWS segue roteando `/auth/customer-login` → lambda e o resto → host do `auto-repair-shop` no SSM).

## T6 — Entrega
Branch `feat/fase4-messaging`, commits Conventional Commits, PR para `develop`; aguardar `validate` e `messaging-ministack`; **não** rodar `apply` em AWS (decisão do usuário).

## Aceite
- [ ] `fmt`/`validate` de todos os layers ok (incl. `messaging`)
- [ ] Job `messaging-ministack` verde com as contagens acima
- [ ] `terraform plan` de `shared` e `messaging` (quando houver credenciais): só adições
- [ ] Plano de `aws` inspecionado e **resultado reportado** (in-place ou replace) antes de qualquer apply
- [ ] Docs/ADRs atualizados; PR aberto
