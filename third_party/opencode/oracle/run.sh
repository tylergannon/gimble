#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
test -f "$root/upstream/packages/client/src/solid/data.ts" || {
  echo "run ../prepare.sh first" >&2
  exit 1
}
exec bun run "$root/session-prefixes.ts" "$@"
