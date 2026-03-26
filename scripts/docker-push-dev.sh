#!/usr/bin/env bash
# Build and push litmus-grafana to the dev registry.
# Mirrors the CI docker job but targets litmus-solutions-dev.
#
# Usage: docker-push-dev.sh [--version <tag>] [--grafana-version <version>] [--skip-push]
#   --version         Image tag to push (default: <package.json version>-<timestamp>)
#   --grafana-version Grafana base image version (default: Dockerfile ARG default)
#   --skip-push       Build the image but do not push (useful for testing the build)
set -euo pipefail

# --- Parse args ---
IMAGE_TAG=""
GRAFANA_VERSION=""
SKIP_PUSH=false
while [[ $# -gt 0 ]]; do
  case $1 in
    --version)         IMAGE_TAG="$2";        shift 2 ;;
    --grafana-version) GRAFANA_VERSION="$2";  shift 2 ;;
    --skip-push)       SKIP_PUSH=true;        shift ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# Load .env
if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

REGISTRY="us-docker.pkg.dev/litmus-customer-facing/litmus-solutions-dev"
IMAGE_NAME="litmus-grafana"
IMAGE_TAG="${IMAGE_TAG:-$(jq -r .version package.json)-$(date +%Y%m%d%H%M%S)}"

echo "==> Building frontend..."
npm run build -- --stats errors-only

echo "==> Building backend..."
mage build:linux
chmod 0755 dist/gpx_edge_linux_amd64

echo "==> Signing plugin..."
npx --yes @grafana/sign-plugin@latest --rootUrls "$ROOT_URLS"

echo "==> Setting up Docker auth..."
DOCKER_CONFIG_DIR=$(mktemp -d)
trap 'rm -rf "$DOCKER_CONFIG_DIR"' EXIT
echo "$DOCKER_AUTH_CONFIG_DEV" > "$DOCKER_CONFIG_DIR/config.json"

echo "==> Verifying registry connection..."
AUTH_HEADER=$(jq -r '.auths."us-docker.pkg.dev".auth' "$DOCKER_CONFIG_DIR/config.json")
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Basic $AUTH_HEADER" \
  "https://us-docker.pkg.dev/v2/")
if [ "$HTTP_STATUS" != "200" ]; then
  echo "ERROR: Registry auth failed (HTTP $HTTP_STATUS) — check DOCKER_AUTH_CONFIG_DEV"
  exit 1
fi
echo "Registry connection OK (HTTP $HTTP_STATUS)"

echo "==> Building Docker image..."
BUILD_ARGS=()
[ -n "$GRAFANA_VERSION" ] && BUILD_ARGS+=(--build-arg "GRAFANA_VERSION=$GRAFANA_VERSION")

# Build to a unique temp tag so previously-pushed tags don't shadow the new image.
BUILD_TAG="${REGISTRY}/${IMAGE_NAME}:build-$(date +%s)"
docker --config "$DOCKER_CONFIG_DIR" build \
  --platform linux/amd64 \
  "${BUILD_ARGS[@]}" \
  -f litmus-grafana/Dockerfile \
  -t "$BUILD_TAG" \
  .

# Explicitly retag from the fresh build so both tags always match what was just built.
docker tag "$BUILD_TAG" "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
docker tag "$BUILD_TAG" "${REGISTRY}/${IMAGE_NAME}:latest"
docker rmi "$BUILD_TAG" > /dev/null

if [ "$SKIP_PUSH" = true ]; then
  echo ""
  echo "Skipping push (--skip-push). Image built locally:"
  echo "  ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
  echo "  ${REGISTRY}/${IMAGE_NAME}:latest"
  exit 0
fi

echo "==> Pushing Docker image..."
docker --config "$DOCKER_CONFIG_DIR" push "${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
docker --config "$DOCKER_CONFIG_DIR" push "${REGISTRY}/${IMAGE_NAME}:latest"

echo ""
echo "Pushed:"
echo "  ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
echo "  ${REGISTRY}/${IMAGE_NAME}:latest"
