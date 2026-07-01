#!/usr/bin/env bash
# Toolbox entrypoint. Best-effort wire-up so any command (status/logs/deploy…)
# run in a fresh container can reach an already-running Kind cluster:
#   - join the `kind` Docker network (created when the cluster is created)
#   - write an "internal" kubeconfig pointing at the control-plane container name
# On the very first `make _k8s-up` neither exists yet; _k8s-cluster does the same
# wire-up right after it creates the cluster.
set -e

CLUSTER="${K8S_CLUSTER:-auto-repair-shop}"

if docker network inspect kind >/dev/null 2>&1; then
  docker network connect kind "$(hostname)" 2>/dev/null || true
fi

if kind get clusters 2>/dev/null | grep -qx "${CLUSTER}"; then
  mkdir -p "${HOME}/.kube"
  kind get kubeconfig --internal --name "${CLUSTER}" > "${HOME}/.kube/config"
fi

exec "$@"
