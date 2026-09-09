#!/bin/sh
set -eu

module=github.com/tylergannon/gimble/cmd/gimble
version=${GIMBLE_VERSION:-latest}

command -v go >/dev/null 2>&1 || {
  echo "gimble: Go is required (https://go.dev/dl/)" >&2
  exit 1
}
command -v codex >/dev/null 2>&1 || {
  echo "gimble: codex must be installed and available on PATH" >&2
  exit 1
}

go install "${module}@${version}"

gobin=$(go env GOBIN)
if [ -z "$gobin" ]; then
  gobin=$(go env GOPATH)/bin
fi
"${gobin}/gimble" plugin install
