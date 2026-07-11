#!/usr/bin/env bash
# Runs INSIDE the toolbox (workdir = repo root). Full local pipeline / teardown.
set -euo pipefail

CLUSTER=auto-repair-shop
NAMESPACE=auto-repair-shop
IMAGE=auto-repair-shop:dev
KIND_CONFIG=k8s/kind/kind-config.yaml
OVERLAY=k8s/manifests/overlays/local
KUBECTL="kubectl --context kind-${CLUSTER}"
SMOKE_URL="http://${CLUSTER}-control-plane:30080/ping"

up() {
  # 1. Kind cluster — reach it from the toolbox via the kind network + internal kubeconfig.
  kind get clusters | grep -qx "${CLUSTER}" || kind create cluster --name "${CLUSTER}" --config "${KIND_CONFIG}"
  docker network connect kind "$(hostname)" 2>/dev/null || true
  mkdir -p "${HOME}/.kube"
  kind get kubeconfig --internal --name "${CLUSTER}" > "${HOME}/.kube/config"
  ${KUBECTL} wait --for=condition=Ready nodes --all --timeout=120s

  # 2. Build the app image and load it into the cluster.
  docker build -t "${IMAGE}" -f Dockerfile .
  kind load docker-image "${IMAGE}" --name "${CLUSTER}"

  # 3. Migrations ConfigMap (single source of truth: the repo migrations/ dir).
  ${KUBECTL} create namespace "${NAMESPACE}" --dry-run=client -o yaml | ${KUBECTL} apply -f -
  ${KUBECTL} create configmap db-migrations --from-file=migrations \
    -n "${NAMESPACE}" --dry-run=client -o yaml | ${KUBECTL} apply -f -
  ${KUBECTL} -n "${NAMESPACE}" delete job db-migrate --ignore-not-found

  # 4. Deploy the workload (app + Service + ConfigMap + Secret + HPA + Postgres + migrate Job).
  ${KUBECTL} apply -k "${OVERLAY}"

  # 5. Smoke test (authoritative).
  echo "waiting for app rollout..."
  ${KUBECTL} -n "${NAMESPACE}" rollout status deploy/auto-repair-shop --timeout=120s
  echo "GET ${SMOKE_URL}"
  curl -fsS "${SMOKE_URL}" && echo " OK"

  # 6. metrics-server so the HPA can read CPU/memory (best-effort; Kind ships none).
  echo "installing metrics-server for the HPA..."
  ${KUBECTL} apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml || true
  ${KUBECTL} -n kube-system patch deployment metrics-server --type=json \
    -p '[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]' 2>/dev/null || true

  echo "done. App reachable from the host at http://localhost:8080"
}

down() {
  kind delete cluster --name "${CLUSTER}"
}

"${1:-up}"
