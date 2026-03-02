#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

echo "> Unit Tests"

test_flags=()
# If running in Prow, we want to generate a machine-readable output file under the location specified via $ARTIFACTS.
if [ -n "${CI:-}" ] && [ -n "${ARTIFACTS:-}" ]; then
  mkdir -p "$ARTIFACTS/junit"
  test_flags+=(--ginkgo.junit-report=junit.xml)
  test_flags+=(--ginkgo.timeout=2m)
else
  timeout_flag="-timeout=2m"
fi

go test ${timeout_flag:+"$timeout_flag"} "$@" ${test_flags:+"${test_flags[@]}"}
