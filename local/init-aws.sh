#!/bin/sh
# Local messaging, storage and secret of the whole ecosystem on MiniStack. Idempotent: running it again
# reuses the existing resources. Runs once in the aws-init container (see docker-compose.yml).
#
# Creates exactly the resources of docs/fase4/contracts/{resources,messaging}.md, with the local prefix
# auto-repair-shop-local-:
#   4 topics:        saga-commands, so-events, billing-events, execution-events
#   10 queues:       billing-queue, execution-queue, catalog-queue, so-saga-replies, history-queue
#                    and one DLQ per queue (<queue>-dlq, maxReceiveCount=5)
#   8 subscriptions: saga-commands -> billing/execution/catalog-queue (filter on target),
#                    billing-events + execution-events -> so-saga-replies,
#                    so-events + billing-events + execution-events -> history-queue
#                    (all with RawMessageDelivery=true)
#   1 DynamoDB table: auto-repair-shop-local-history (PK, SK; on-demand)
#   1 secret:        auto-repair-shop-local/lambda-auth ({"POSTGRES_PASSWORD", "JWT_SECRET"}; read by the lambda)
#
# The values of JWT_SECRET and LAMBDA_DB_PASSWORD come from the environment of this container (local, fixed,
# development-only values; they are not real secrets).
set -eu

: "${JWT_SECRET:?JWT_SECRET is required}"
: "${LAMBDA_DB_PASSWORD:?LAMBDA_DB_PASSWORD is required}"

PREFIX="auto-repair-shop-local"
REGION="${AWS_REGION:-us-east-1}"
ENDPOINT="${AWS_ENDPOINT_URL:-http://ministack:4566}"

aws_local() {
  aws --endpoint-url "$ENDPOINT" --region "$REGION" "$@"
}

# json_attrs <name> <value>: {"name":"value"} with the value's quotes escaped (for --attributes).
json_attrs() {
  escaped=$(printf '%s' "$2" | sed 's/"/\\"/g')
  printf '{"%s":"%s"}' "$1" "$escaped"
}

queue_arn() {
  aws_local sqs get-queue-attributes --queue-url "$1" --attribute-names QueueArn \
    --query Attributes.QueueArn --output text
}

topic() {
  aws_local sns create-topic --name "${PREFIX}-$1" --query TopicArn --output text
}

# create_queue <logical name>: DLQ + queue with RedrivePolicy (maxReceiveCount=5). Prints the queue URL.
create_queue() {
  name="${PREFIX}-$1"
  dlq_url=$(aws_local sqs create-queue --queue-name "${name}-dlq" --query QueueUrl --output text)
  dlq_arn=$(queue_arn "$dlq_url")
  redrive=$(printf '{"deadLetterTargetArn":"%s","maxReceiveCount":"5"}' "$dlq_arn")
  url=$(aws_local sqs create-queue --queue-name "$name" --query QueueUrl --output text)
  aws_local sqs set-queue-attributes --queue-url "$url" --attributes "$(json_attrs RedrivePolicy "$redrive")" >/dev/null
  printf '%s' "$url"
}

# allow_topics <queue_url> <queue_arn> <topic_arn>...: queue policy letting those topics send to the queue.
allow_topics() {
  queue_url=$1
  q_arn=$2
  shift 2
  sources=""
  for t in "$@"; do
    sources="${sources}${sources:+,}\"${t}\""
  done
  policy="{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"sns.amazonaws.com\"},\"Action\":\"sqs:SendMessage\",\"Resource\":\"${q_arn}\",\"Condition\":{\"ArnEquals\":{\"aws:SourceArn\":[${sources}]}}}]}"
  aws_local sqs set-queue-attributes --queue-url "$queue_url" --attributes "$(json_attrs Policy "$policy")" >/dev/null
}

# subscribe <topic_arn> <queue_arn> [target]: raw delivery; FilterPolicy on the "target" attribute when given.
subscribe() {
  sub_arn=$(aws_local sns subscribe --topic-arn "$1" --protocol sqs --notification-endpoint "$2" \
    --query SubscriptionArn --output text)
  aws_local sns set-subscription-attributes --subscription-arn "$sub_arn" \
    --attribute-name RawMessageDelivery --attribute-value true >/dev/null
  if [ -n "${3:-}" ]; then
    aws_local sns set-subscription-attributes --subscription-arn "$sub_arn" \
      --attribute-name FilterPolicyScope --attribute-value MessageAttributes >/dev/null
    aws_local sns set-subscription-attributes --subscription-arn "$sub_arn" \
      --attribute-name FilterPolicy --attribute-value "{\"target\":[\"$3\"]}" >/dev/null
  fi
}

# Topics (4)
COMMANDS_ARN=$(topic saga-commands)
SO_EVENTS_ARN=$(topic so-events)
BILLING_EVENTS_ARN=$(topic billing-events)
EXECUTION_EVENTS_ARN=$(topic execution-events)

# Queues (5 main + 5 DLQs = 10)
BILLING_URL=$(create_queue billing-queue)
EXECUTION_URL=$(create_queue execution-queue)
CATALOG_URL=$(create_queue catalog-queue)
SO_REPLIES_URL=$(create_queue so-saga-replies)
HISTORY_URL=$(create_queue history-queue)

BILLING_ARN=$(queue_arn "$BILLING_URL")
EXECUTION_ARN=$(queue_arn "$EXECUTION_URL")
CATALOG_ARN=$(queue_arn "$CATALOG_URL")
SO_REPLIES_ARN=$(queue_arn "$SO_REPLIES_URL")
HISTORY_ARN=$(queue_arn "$HISTORY_URL")

# Queue policies: which topics may send to each queue
allow_topics "$BILLING_URL" "$BILLING_ARN" "$COMMANDS_ARN"
allow_topics "$EXECUTION_URL" "$EXECUTION_ARN" "$COMMANDS_ARN"
allow_topics "$CATALOG_URL" "$CATALOG_ARN" "$COMMANDS_ARN"
allow_topics "$SO_REPLIES_URL" "$SO_REPLIES_ARN" "$BILLING_EVENTS_ARN" "$EXECUTION_EVENTS_ARN"
allow_topics "$HISTORY_URL" "$HISTORY_ARN" "$SO_EVENTS_ARN" "$BILLING_EVENTS_ARN" "$EXECUTION_EVENTS_ARN"

# Subscriptions (8)
subscribe "$COMMANDS_ARN" "$BILLING_ARN" billing
subscribe "$COMMANDS_ARN" "$EXECUTION_ARN" execution
subscribe "$COMMANDS_ARN" "$CATALOG_ARN" catalog
subscribe "$BILLING_EVENTS_ARN" "$SO_REPLIES_ARN"
subscribe "$EXECUTION_EVENTS_ARN" "$SO_REPLIES_ARN"
subscribe "$SO_EVENTS_ARN" "$HISTORY_ARN"
subscribe "$BILLING_EVENTS_ARN" "$HISTORY_ARN"
subscribe "$EXECUTION_EVENTS_ARN" "$HISTORY_ARN"

# DynamoDB read model of the History service (1 table)
TABLE="${PREFIX}-history"
if ! aws_local dynamodb describe-table --table-name "$TABLE" >/dev/null 2>&1; then
  aws_local dynamodb create-table --table-name "$TABLE" \
    --attribute-definitions AttributeName=PK,AttributeType=S AttributeName=SK,AttributeType=S \
    --key-schema AttributeName=PK,KeyType=HASH AttributeName=SK,KeyType=RANGE \
    --billing-mode PAY_PER_REQUEST >/dev/null
fi

# Secret read by the login lambda (1 secret). Keys are the environment variable names, as in production.
SECRET_NAME="auto-repair-shop-local/lambda-auth"
SECRET_JSON=$(printf '{"POSTGRES_PASSWORD":"%s","JWT_SECRET":"%s"}' "$LAMBDA_DB_PASSWORD" "$JWT_SECRET")
if aws_local secretsmanager describe-secret --secret-id "$SECRET_NAME" >/dev/null 2>&1; then
  aws_local secretsmanager put-secret-value --secret-id "$SECRET_NAME" --secret-string "$SECRET_JSON" >/dev/null
else
  aws_local secretsmanager create-secret --name "$SECRET_NAME" --secret-string "$SECRET_JSON" >/dev/null
fi

echo "local messaging ready: 4 topics, 10 queues, 8 subscriptions, table $TABLE, secret $SECRET_NAME"
