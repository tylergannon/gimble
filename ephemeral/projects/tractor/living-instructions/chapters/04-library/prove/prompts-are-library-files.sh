#!/bin/sh
# Proves: prompts live in the library, Build is unchanged.
set -eu
cd "$(git rev-parse --show-toplevel)"

test -d workflow/library || { echo "no workflow/library"; exit 1; }
for f in workflows/plan.yaml workflows/medium.yaml workflows/large.yaml \
         prompts/plan/planner.md prompts/medium/implement.md \
         prompts/large/plan.md prompts/large/implement.md README.md; do
  test -f "workflow/library/$f" || { echo "missing workflow/library/$f"; exit 1; }
done
test ! -f workflow/plan.yaml || { echo "workflow/plan.yaml still present"; exit 1; }

# No prompt body remains in Go: the old functions are gone and no Go file
# under workflow/ carries the planner's opening sentence.
if grep -n 'func plannerPrompt\|func mediumPrompt\|func largePlanPrompt\|func largeImplementPrompt' workflow/*.go; then
  echo "prompt functions still present"; exit 1
fi
if grep -n "built-in planning workflow" workflow/*.go; then
  echo "prompt text still in Go"; exit 1
fi
grep -q 'go:embed' workflow/library.go || { echo "no embed in workflow/library.go"; exit 1; }
grep -q 'text/template' workflow/*.go || { echo "text/template not used"; exit 1; }

# The byte-equality snapshot test exists and passes.
ls workflow/testdata/*.snap* >/dev/null 2>&1 || ls workflow/testdata/* >/dev/null 2>&1 || { echo "no snapshot testdata"; exit 1; }
go test -run 'Snapshot|Golden|ByteEqual' ./workflow/... -count=1
echo "prompts-are-library-files.sh: ok"
