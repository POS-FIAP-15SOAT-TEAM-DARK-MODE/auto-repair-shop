#!/usr/bin/env bash
# Docker-only launcher for the k8s toolbox. The host needs ONLY Docker.
# Invoked by the Makefile: `make k8s-up` / `make k8s-down` / `make k8s-shell`.
set -euo pipefail

command -v docker >/dev/null 2>&1 || { echo "Docker is required on the host."; exit 1; }
docker info >/dev/null 2>&1 || { echo "Docker daemon not reachable."; exit 1; }

IMAGE="auto-repair-shop-toolbox"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

case "${1:-shell}" in
  up)    CMD=(bash k8s/toolbox/pipeline.sh up) ;;
  down)  CMD=(bash k8s/toolbox/pipeline.sh down) ;;
  shell) CMD=(bash) ;;
  *)     CMD=("$@") ;;   # passthrough (advanced)
esac

docker build -t "${IMAGE}" "${SCRIPT_DIR}"

exec docker run --rm -it \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v "${REPO_ROOT}:/workspace" \
  -w /workspace \
  "${IMAGE}" \
  "${CMD[@]}"
