#!/bin/sh
set -eu

revision=c55ee2a8152603f04a409163bd3edf79c425fbd7
root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
target="$root/oracle/upstream"

if [ "$(bun --version)" != "1.3.14" ]; then
  echo "OpenCode oracle requires Bun 1.3.14" >&2
  exit 1
fi
if [ ! -d "$target/.git" ]; then
  mkdir -p "$target"
  git -C "$target" init -q
  git -C "$target" remote add origin https://github.com/anomalyco/opencode.git
  git -C "$target" fetch --depth 1 origin "$revision"
  git -C "$target" checkout -q --detach FETCH_HEAD
fi
test "$(git -C "$target" rev-parse HEAD)" = "$revision"

verify() {
  path=$1 expected=$2
  actual=$(git -C "$target" hash-object "$path")
  if [ "$actual" != "$expected" ]; then
    echo "$path: expected $expected, got $actual" >&2
    exit 1
  fi
}
verify packages/client/src/solid/data.ts 83c6eeb32ecccb1e4e9fb589647ea8e33f96b71e
verify packages/schema/src/session-event.ts 6e43890ff8ae67d88f4990d710875f0d67b8db65
verify packages/schema/src/session-message.ts 7e8ce0482ae9db77bb1ddc8e9bb3fb0b4c41dcdf
verify packages/schema/src/session-inbox.ts baadcc947b18c267e10fee76cb648f140ccdffe6
verify LICENSE 6439474beed8e0271df9862eff97ffd70ec2464c
verify bun.lock 30b8f7f4d5e89f3360728495701a4bbeda8a3732

git -C "$target" diff --quiet
(cd "$target" && bun install --frozen-lockfile)
git -C "$target" diff --quiet
mkdir -p "$root/oracle/node_modules/@opencode"
link_dependency() {
  name=$1 source=$2
  if [ ! -e "$root/oracle/node_modules/$name" ]; then
    ln -s "$source" "$root/oracle/node_modules/$name"
  fi
}
link_dependency solid-js ../upstream/packages/client/node_modules/solid-js
link_dependency effect ../upstream/packages/client/node_modules/effect
link_dependency @opencode/schema ../../upstream/packages/schema
link_dependency @opencode/protocol ../../upstream/packages/protocol
