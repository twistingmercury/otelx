#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPOSITORY_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
IS_LOCAL="${IS_LOCAL:-0}"
IMAGE="otelx-ci:local"

cleanup_local_image() {
    local exit_status=$?

    if [[ "${IS_LOCAL}" == "1" ]]; then
        docker rmi "${IMAGE}" || true
    fi

    exit "${exit_status}"
}

trap cleanup_local_image EXIT

cd "${REPOSITORY_ROOT}"

docker build --file "${SCRIPT_DIR}/Dockerfile" --target verify --tag "${IMAGE}" .
