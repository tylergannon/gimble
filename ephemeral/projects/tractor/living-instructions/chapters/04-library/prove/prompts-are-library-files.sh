#!/bin/sh
# Proves: prompts live in the library, Build is unchanged. This is the
# sprint's tripwire; the proof that Build renders the library files is
# show-equals-build.sh and orphan-walk-and-render.sh (the Build dumper
# and the sentinel mutations).
set -eu
cd "$(git rev-parse --show-toplevel)"

test -d workflow/library || { echo "no workflow/library"; exit 1; }
for f in workflows/plan.yaml workflows/medium.yaml workflows/large.yaml \
         prompts/plan/planner.md prompts/medium/implement.md \
         prompts/large/plan.md prompts/large/implement.md README.md; do
  test -f "workflow/library/$f" || { echo "missing workflow/library/$f"; exit 1; }
done
test ! -f workflow/plan.yaml || { echo "workflow/plan.yaml still present"; exit 1; }

# No prompt body remains in Go: the planner's opening sentence is gone.
if grep -n "built-in planning workflow" workflow/*.go; then
  echo "prompt text still in Go"; exit 1
fi
grep -q 'go:embed' workflow/library.go || { echo "no embed in workflow/library.go"; exit 1; }
grep -q 'text/template' workflow/*.go || { echo "text/template not used"; exit 1; }

# The immutable baseline: the commit recorded in prove/base-commit.txt,
# made before sprint 1 touched the workflow package. The check extracts
# that commit, builds a dumper against its Build, and compares every
# node's payload with the current show --raw for the same parameters.
# Snapshots and tests the coder owns are not the baseline; this is.
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
base="$(cat ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/base-commit.txt)"
mkdir -p "$tmp/base" "$tmp/bin" "$tmp/demo"
git archive "$base" | tar -x -C "$tmp/base"
mkdir -p "$tmp/base/cmd/builddump"
sed -n '/^  cat > "\$tmp\/src\/cmd\/builddump\/main.go" <<'"'"'EOF'"'"'$/,/^EOF$/p' ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/lib.sh \
  | sed '1d;$d' > "$tmp/base/cmd/builddump/main.go"
# The baseline package has no Render; keep only the Build path.
python3 - "$tmp/base/cmd/builddump/main.go" <<'PYEOF'
import re, sys
p = sys.argv[1]; s = open(p).read()
s = re.sub(r'\t// "render" mode.*?\n\t}\n', '', s, flags=re.S)
open(p, "w").write(s)
PYEOF
(cd "$tmp/base" && go build -o "$tmp/bin/basedump" ./cmd/builddump)
go build -o "$tmp/bin/tractor" ./cmd/tractor
printf 'Print a greeting.\n' > "$tmp/demo/seed.md"
for wf in plan medium large; do
  awk '/^  - id:/{print $3}' "$tmp/base/workflow/$wf.yaml" | while read -r node; do
    if [ "$wf" = plan ]; then s="$tmp/demo/seed.md"; else s=""; fi
    "$tmp/bin/basedump" "$wf" "$node" "$tmp/demo" "$tmp/bin/tractor" "$s" > "$tmp/want.txt" 2>/dev/null || continue
    if [ "$wf" = plan ]; then
      "$tmp/bin/tractor" workflow show "$wf" --project demo --seed "$s" --workdir "$tmp/demo" --node "$node" --raw > "$tmp/got.txt"
    else
      "$tmp/bin/tractor" workflow show "$wf" --project demo --workdir "$tmp/demo" --node "$node" --raw > "$tmp/got.txt"
    fi
    cmp -s "$tmp/want.txt" "$tmp/got.txt" || { echo "baseline: $wf/$node differs from the pre-migration Build at $base"; diff "$tmp/want.txt" "$tmp/got.txt" | head -20; exit 1; }
    echo "baseline: $wf/$node equals the pre-migration Build"
  done
done
# The coder's snapshot test is the tripwire for later content edits; it
# must run and pass too.
go test -v -run 'TestBuildMatchesSnapshot' ./workflow/ -count=1 2>&1 | tee /dev/stderr \
  | grep -q -- '--- PASS: TestBuildMatchesSnapshot' || { echo "TestBuildMatchesSnapshot did not pass"; exit 1; }
echo "prompts-are-library-files.sh: ok"
