#!/bin/sh
# End-to-end validation of the local ecosystem: starts it (idempotent), runs the BDD suite and
# prints a PASS/FAIL checklist. Exit code is the number of failed items (0 = everything passed).
# Needs a running Docker daemon. Usage: sh scripts/validate-local.sh   (or: make validate)
set -u

ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT" || exit 1
E2E_BASE_URL=${E2E_BASE_URL:-http://localhost:8080}
LOG=$(mktemp "${TMPDIR:-/tmp}/validate-local.XXXXXX")
trap 'rm -f "$LOG"' EXIT

FAILS=0
check() { # check <label> <ok: 0|1>
	if [ "$2" -eq 0 ]; then printf 'PASS  %s\n' "$1"; else printf 'FAIL  %s\n' "$1"; FAILS=$((FAILS + 1)); fi
}
scenario_passed() { grep -q -- "--- PASS: TestFeatures/$1 " "$LOG"; }

echo "==> starting the ecosystem (make start)"
if ! make start; then
	echo "FAIL  the ecosystem did not start"
	exit 1
fi

echo "==> running the BDD suite"
(cd e2e && E2E_BASE_URL="$E2E_BASE_URL" go test ./... -count=1 -v) >"$LOG" 2>&1
SUITE_RC=$?

echo
echo "==> checklist"
scenario_passed A_known_CPF_receives_a_token_that_the_services_accept; check "staff and customer login (customer by CPF through the lambda)" $?
scenario_passed Order_is_received,_approved,_paid,_executed,_delivered_and_recorded_in_the_history; check "order completes up to DELIVERED and is recorded in the history" $?
scenario_passed Attendant_rejects_the_payment_of_an_accepted_budget; check "payment rejected => REJECTED and the stock is returned" $?
scenario_passed Mechanic_cancels_all_works_after_the_payment_is_approved; check "execution failure => CANCELLED and refund" $?
scenario_passed Accept_and_confirm_sent_twice; check "duplicate message does not duplicate the effect" $?
scenario_passed Internal_endpoints_are_not_reachable_through_the_gateway; check "/internal/* is blocked by the gateway" $?
scenario_passed A_protected_route_without_a_token_is_refused_by_the_service; check "a protected route without a token answers 401" $?

METRICS_OK=0
for port in 8081 8082 8083 8084 8085 8086; do
	curl -fsS "http://localhost:$port/metrics" 2>/dev/null | grep -q '^# TYPE' || METRICS_OK=1
done
check "/metrics answers on the six services" $METRICS_OK

SAGA_OK=1
curl -fsS "http://localhost:8086/metrics" 2>/dev/null | grep -q '^saga_started_total [1-9]' && SAGA_OK=0
check "saga metrics counted at least one saga" $SAGA_OK

[ "$SUITE_RC" -eq 0 ] || { check "BDD suite exit code" 1; tail -30 "$LOG"; }

echo
if [ "$FAILS" -eq 0 ]; then echo "RESULT: 0 failures"; else echo "RESULT: $FAILS failure(s)"; fi
exit "$FAILS"
