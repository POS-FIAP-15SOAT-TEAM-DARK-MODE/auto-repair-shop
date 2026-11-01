#!/bin/sh
# Clones the service repositories into services/<repo> (HTTPS), or fast-forwards the ones already there.
#
#   sh scripts/clone-services.sh               clone or `git pull --ff-only` every repository
#   sh scripts/clone-services.sh --if-missing  clone only the repositories whose directory does not exist yet
#
# SERVICES_BRANCH selects the branch to check out (default: develop). GITHUB_ORG changes the owner.
set -eu

ORG="${GITHUB_ORG:-POS-FIAP-15SOAT-TEAM-DARK-MODE}"
BRANCH="${SERVICES_BRANCH:-develop}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$ROOT/services"

IF_MISSING=0
case "${1:-}" in
  --if-missing) IF_MISSING=1 ;;
  "") ;;
  *) echo "usage: $0 [--if-missing]" >&2; exit 2 ;;
esac

REPOS="
auto-repair-shop-customer-service
auto-repair-shop-catalog-service
auto-repair-shop-billing-service
auto-repair-shop-execution-service
auto-repair-shop-history-service
auto-repair-shop-service-order-service
auto-repair-shop-lambda-auth
"

mkdir -p "$DEST"

for repo in $REPOS; do
  dir="$DEST/$repo"
  if [ -d "$dir/.git" ]; then
    if [ "$IF_MISSING" -eq 1 ]; then
      echo "present: services/$repo"
      continue
    fi
    echo "update: services/$repo ($BRANCH)"
    git -C "$dir" fetch --quiet origin "$BRANCH"
    git -C "$dir" checkout --quiet "$BRANCH"
    git -C "$dir" pull --quiet --ff-only origin "$BRANCH"
  else
    echo "clone: $ORG/$repo ($BRANCH) -> services/$repo"
    git clone --quiet --branch "$BRANCH" "https://github.com/$ORG/$repo.git" "$dir"
  fi
done

echo "services ready in $DEST"
