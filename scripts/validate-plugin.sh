#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
PLUGIN_ID=$(node -p "require('${ROOT_DIR}/src/plugin.json').id")
PLUGIN_VERSION=$(node -p "require('${ROOT_DIR}/package.json').version")

usage() {
  cat <<'EOF'
Usage:
  npm run validate:plugin
  npm run validate:plugin -- <archive-url-or-path>

Behavior:
  - No argument: run `npm run build`, package the fresh dist/ output, and validate it locally.
  - One argument: validate a remote archive URL or a local zip file path.

Optional environment variables:
  VALIDATOR_FLAGS            Extra flags passed to grafana/plugin-validator-cli
  VALIDATOR_SOURCE_CODE_URI  Override -sourceCodeUri

Examples:
  npm run validate:plugin
  npm run validate:plugin -- https://github.com/org/repo/releases/download/v1.0.0/litmus-edge-datasource-1.0.0.zip
  VALIDATOR_FLAGS='-strict' npm run validate:plugin
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

validator_flags=()
# shellcheck disable=SC2206
validator_flags=(${VALIDATOR_FLAGS:-})

prepare_source_snapshot() {
  local destination=$1

  mkdir -p "$destination"
  tar -C "$ROOT_DIR" \
    --exclude='.git' \
    --exclude='node_modules' \
    --exclude='dist' \
    --exclude='.env' \
    --exclude='coverage' \
    --exclude='playwright-report' \
    --exclude='test-results' \
    -cf - . | tar -C "$destination" -xf -
}

if [[ $# -eq 0 ]]; then
  printf 'Building plugin with npm run build\n'
  (
    cd "$ROOT_DIR"
	    npm run build
  )

  tmp_dir=$(mktemp -d)
  archive_dir="${tmp_dir}/${PLUGIN_ID}"
  archive_path="${tmp_dir}/${PLUGIN_ID}-${PLUGIN_VERSION}.zip"
  source_dir="${tmp_dir}/source_code"
  source_code_uri=${VALIDATOR_SOURCE_CODE_URI:-file:///source_code}

  trap 'rm -rf "${tmp_dir}"' EXIT

  prepare_source_snapshot "$source_dir"
  cp -R "${ROOT_DIR}/dist" "$archive_dir"
  (
    cd "$tmp_dir"
    zip -qr "$archive_path" "$PLUGIN_ID"
  )

  printf 'Validating local archive %s\n' "$archive_path"
  printf 'Validator may take a minute on first run while Docker layers and analyzers initialize.\n'
  docker run --pull=always \
    --rm \
    -v "${archive_path}:/archive.zip" \
    -v "${source_dir}:/source_code" \
    grafana/plugin-validator-cli \
    "${validator_flags[@]}" \
    -sourceCodeUri "$source_code_uri" \
    /archive.zip
  exit 0
fi

if [[ $# -gt 1 ]]; then
  printf 'Expected at most one archive URL or path.\n\n' >&2
  usage >&2
  exit 1
fi

target=$1

case "$target" in
  http://*|https://*)
    if [[ -n "${VALIDATOR_SOURCE_CODE_URI:-}" ]]; then
      docker run --pull=always --rm grafana/plugin-validator-cli \
        "${validator_flags[@]}" \
        -sourceCodeUri "$VALIDATOR_SOURCE_CODE_URI" \
        "$target"
    else
      docker run --pull=always --rm grafana/plugin-validator-cli \
        "${validator_flags[@]}" \
        "$target"
    fi
    ;;
  *)
    if [[ ! -f "$target" ]]; then
      printf 'Archive not found: %s\n' "$target" >&2
      exit 1
    fi

    archive_path=$(python3 -c 'import os,sys; print(os.path.abspath(sys.argv[1]))' "$target")
    tmp_dir=$(mktemp -d)
    source_dir="${tmp_dir}/source_code"
    source_code_uri=${VALIDATOR_SOURCE_CODE_URI:-file:///source_code}
    trap 'rm -rf "${tmp_dir}"' EXIT
    prepare_source_snapshot "$source_dir"

    printf 'Validating local archive %s\n' "$archive_path"
    printf 'Validator may take a minute on first run while Docker layers and analyzers initialize.\n'
    docker run --pull=always \
      --rm \
      -v "${archive_path}:/archive.zip" \
      -v "${source_dir}:/source_code" \
      grafana/plugin-validator-cli \
      "${validator_flags[@]}" \
      -sourceCodeUri "$source_code_uri" \
      /archive.zip
    ;;
esac
