#!/bin/bash
set -euo pipefail

TMP_DIR=$(mktemp -d)

function cleanup() {
    return_code=$?
    rm -rf "${TMP_DIR}"
    exit "${return_code}"
}
trap "cleanup" EXIT

OUTDIR=${TMP_DIR} ./hack/update-generated-bindata.sh

diff -Naup {.,${TMP_DIR}}/pkg/manifests/bindata.go
