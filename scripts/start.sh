#!/bin/sh
# Starts the local ecosystem of the Auto Repair Shop and checks it. One entry point, four modes:
#
#   sh scripts/start.sh          start (default): checks the prerequisites, clones the missing service
#                                repositories, pulls the base images, builds one image at a time, starts the
#                                stack and waits for /ping and /ready. Running it again is safe.
#   sh scripts/start.sh status   state of the containers, /ping of every service and /ready of the gateway
#   sh scripts/start.sh smoke    one service order through the gateway, from login to the execution queue
#   sh scripts/start.sh reset    removes the containers and their volumes, then starts again
#
# Environment (all optional):
#   SERVICES_DIR     clones of the service repositories (absolute, or relative to the repository; default services/)
#   GATEWAY_URL      gateway address for status and smoke (default http://localhost:8080)
#   SERVICES_BRANCH  branch cloned by clone-services.sh (default develop)
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"
MIN_MEMORY_BYTES=4000000000

# Built one at a time, in this order (the same order as make build-all): parallel builds exhaust a small Docker VM.
BUILD_ORDER="customer-service catalog-service billing-service execution-service history-service service-order-service lambda lambda-adapter gateway"
# Pulled before the build, so an old copy in the local cache (for example golang:1.26) is not used silently.
BASE_IMAGES="golang:1.27-alpine postgres:15-alpine migrate/migrate:v4.17.1 ministackorg/ministack amazon/aws-cli public.ecr.aws/lambda/provided:al2023"

# The lambda image is built from a Dockerfile of this repository, which is outside its build context.
LOCAL_LAMBDA_DOCKERFILE="${LOCAL_LAMBDA_DOCKERFILE:-$ROOT/local/lambda/Dockerfile}"
export LOCAL_LAMBDA_DOCKERFILE

# Seed data (docs/fase4/contracts/resources.md). The customer CPF is the one of e2e/steps_test.go; it is never printed.
ATTENDANT_EMAIL="attendant@autorepairshop.com"
ATTENDANT_PASSWORD="attendant123"
MECHANIC_EMAIL="mechanic@autorepairshop.com"
MECHANIC_PASSWORD="mechanic123"
CUSTOMER_A_CPF="52998224725"
CUSTOMER_A_ID="44444444-4444-4444-8444-000000000002"
VEHICLE_A_ID="44444444-4444-4444-8444-000000000003"
WORK_OIL_CHANGE_ID="11111111-1111-4111-8111-000000000001"

say() { printf '%s\n' "$*"; }
warn() { printf 'warning: %s\n' "$*" >&2; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

check_tools() {
  missing=""
  for tool in docker git make curl; do
    command -v "$tool" >/dev/null 2>&1 || missing="$missing $tool"
  done
  [ -z "$missing" ] || die "missing tools:$missing. Install them and run again (README, pre-requisitos)."
  docker compose version >/dev/null 2>&1 || die "Docker Compose v2 is missing: 'docker compose version' fails."
}

require_daemon() {
  docker info >/dev/null 2>&1 || die "the Docker daemon is not running. Start Docker (Docker Desktop) and run this again."
}

check_memory() {
  memory=$(docker info --format '{{.MemTotal}}' 2>/dev/null || echo 0)
  if [ "$memory" -lt "$MIN_MEMORY_BYTES" ]; then
    warn "Docker has $((memory / 1048576)) MiB of memory; the stack runs better with 4 GB or more (Docker Desktop: Settings > Resources). Continuing."
  fi
}

pull_base_images() {
  for image in $BASE_IMAGES; do
    if docker pull --quiet "$image" >/dev/null 2>&1; then
      say "  pulled $image"
    else
      warn "could not pull $image; the local copy is used if there is one"
    fi
  done
}

summary() {
  cat <<EOF

Auto Repair Shop is up.

  Gateway (the only public entry point)  http://localhost:8080
  Swagger of the documented services     http://localhost:8080/swagger/index.html
  Services (direct, for debugging)       customer :8081  catalog :8082  billing :8083
                                         execution :8084  history :8085  service-order :8086
  Login lambda (through the gateway)     POST http://localhost:8080/auth/customer-login
  MiniStack (AWS emulator)               http://localhost:4566

Seed users (passwords in README, section "Usuarios de seed"):
  staff: $ATTENDANT_EMAIL (attendant), $MECHANIC_EMAIL (mechanic)
  customer: login by CPF with POST /auth/customer-login (the seed customer A)

Useful commands:
  make status    state of the containers and of every /ping and /ready
  make smoke     one service order through the gateway, about one minute
  make e2e       the BDD scenarios against the gateway
  make logs      follow the logs
  make down-all  stop the stack and remove its volumes
  make reset     remove the volumes and start again
EOF
}

start_stack() {
  check_tools
  require_daemon
  check_memory

  say "1/6 service repositories (missing ones are cloned into ${SERVICES_DIR:-services/})"
  sh "$ROOT/scripts/clone-services.sh" --if-missing

  say "2/6 base images"
  pull_base_images

  say "3/6 images, one at a time"
  for service in $BUILD_ORDER; do
    docker compose build "$service" || die "the build of $service failed (the output above shows why)"
  done

  say "4/6 containers"
  docker compose up -d

  say "5/6 waiting for /ping of the services and /ready of the gateway (up to 3 minutes)"
  sh "$ROOT/scripts/wait-ready.sh" 180 || die "the stack did not become ready. 'docker compose logs <service>' shows why; 'make status' lists the state."

  say "6/6 summary"
  summary
}

probe() {
  if curl -fsS -m 3 -o /dev/null "$2" 2>/dev/null; then
    say "  ok    $1"
  else
    say "  FAIL  $1 ($2)"
    failures=$((failures + 1))
  fi
}

show_status() {
  check_tools
  require_daemon
  docker compose ps --format 'table {{.Service}}\t{{.Status}}'
  echo
  failures=0
  probe "gateway /ping" "$GATEWAY_URL/ping"
  probe "gateway /ready" "$GATEWAY_URL/ready"
  probe "customer-service /ping" "http://localhost:8081/ping"
  probe "catalog-service /ping" "http://localhost:8082/ping"
  probe "billing-service /ping" "http://localhost:8083/ping"
  probe "execution-service /ping" "http://localhost:8084/ping"
  probe "history-service /ping" "http://localhost:8085/ping"
  probe "service-order-service /ping" "http://localhost:8086/ping"
  [ "$failures" -eq 0 ] || die "$failures check(s) failed"
  say "all checks passed"
}

# call METHOD PATH TOKEN WANTED_STATUS [JSON]: one request through the gateway. It sets LAST_STATUS and
# LAST_BODY, and returns 1 when the status differs from WANTED_STATUS. Bodies are never printed (tokens, CPF).
call() {
  method=$1 path=$2 token=$3 want=$4 data=${5:-}
  set -- -sS -m 30 -w '\n%{http_code}' -X "$method"
  if [ -n "$token" ]; then set -- "$@" -H "Authorization: Bearer $token"; fi
  if [ -n "$data" ]; then set -- "$@" -H 'Content-Type: application/json' --data "$data"; fi
  reply=$(curl "$@" "$GATEWAY_URL$path") || reply=$(printf '\n000')
  LAST_BODY=$(printf '%s' "$reply" | sed '$d')
  LAST_STATUS=$(printf '%s' "$reply" | tail -n 1)
  [ "$LAST_STATUS" = "$want" ]
}

# expect is call, but a different status ends the smoke run.
expect() {
  call "$@" || die "$STEP: HTTP $LAST_STATUS, expected $4"
}

# json_field NAME: the first string value of NAME in LAST_BODY (the services answer compact JSON).
json_field() {
  printf '%s' "$LAST_BODY" | grep -o "\"$1\":\"[^\"]*\"" | head -n 1 | cut -d'"' -f4
}

# until_true SECONDS COMMAND...: runs COMMAND every second until it succeeds, or ends the smoke run.
until_true() {
  limit=$1
  shift
  i=0
  until "$@"; do
    i=$((i + 1))
    [ "$i" -lt "$limit" ] || die "$STEP: not reached within $limit s"
    sleep 1
  done
}

login() {
  STEP="login as $1"
  expect POST /v1/auth/login "" 200 "{\"email\":\"$1\",\"password\":\"$2\"}"
  json_field token
}

budget_generated() {
  call GET "/v1/budgets/$SO_ID" "$ATTENDANT" 200 && [ "$(json_field status)" = GENERATED ]
}

order_status_is() {
  call GET "/v1/service-order/$SO_ID/status" "" 200 && [ "$(json_field status)" = "$1" ]
}

payment_pending() {
  call GET "/v1/payments/$SO_ID" "$ATTENDANT" 200 && [ "$(json_field status)" = PENDING ]
}

in_execution_queue() {
  call GET /v1/execution/queue "$MECHANIC" 200 && printf '%s' "$LAST_BODY" | grep -q "\"serviceOrderId\":\"$SO_ID\""
}

smoke() {
  check_tools
  require_daemon

  ATTENDANT=$(login "$ATTENDANT_EMAIL" "$ATTENDANT_PASSWORD")
  MECHANIC=$(login "$MECHANIC_EMAIL" "$MECHANIC_PASSWORD")

  STEP="create the service order"
  expect POST /v1/service-order "$ATTENDANT" 201 "{\"client\":\"$CUSTOMER_A_ID\",\"vehicle\":\"$VEHICLE_A_ID\",\"works\":[\"$WORK_OIL_CHANGE_ID\"]}"
  SO_ID=$(json_field id)
  say "smoke: service order $SO_ID created"

  STEP="receive the order"
  expect PUT "/v1/service-order/$SO_ID/received" "$ATTENDANT" 204
  STEP="start the diagnosis"
  expect PUT "/v1/service-order/$SO_ID/start-diagnosis" "$MECHANIC" 204
  STEP="send the budget"
  expect PUT "/v1/service-order/$SO_ID/send" "$MECHANIC" 204

  STEP="budget generated"
  until_true 30 budget_generated
  STEP="order waiting for the customer"
  until_true 30 order_status_is AWAITING_APPROVAL
  say "smoke: budget generated, order AWAITING_APPROVAL"

  STEP="customer login"
  expect POST /auth/customer-login "" 200 "{\"cpf\":\"$CUSTOMER_A_CPF\"}"
  CUSTOMER=$(json_field token)
  STEP="customer accepts the budget"
  expect PUT "/v1/service-order/$SO_ID/accept" "$CUSTOMER" 204
  say "smoke: budget accepted by the customer"

  STEP="payment pending"
  until_true 30 payment_pending
  PAYMENT_ID=$(json_field id)
  STEP="confirm the payment"
  expect POST "/v1/payments/$PAYMENT_ID/confirm" "$ATTENDANT" 204
  say "smoke: payment confirmed by the attendant"

  STEP="execution queue"
  until_true 30 in_execution_queue
  say "smoke: service order $SO_ID is in the execution queue; the flow works end to end"
}

reset_stack() {
  check_tools
  require_daemon
  docker compose down -v
  start_stack
}

case "${1:-start}" in
  start) start_stack ;;
  status) show_status ;;
  smoke) smoke ;;
  reset) reset_stack ;;
  *) die "usage: sh scripts/start.sh [start|status|smoke|reset]" ;;
esac
