#!/bin/sh
# Proves: the sprint 3 doctrine pages and skeletons exist, each page
# points at its source, every skeleton is cited by a prompt, and every
# page renders and is cited. This is the sprint's demonstration, not
# P8's proof (validation/P8/design.md).
set -eu
cd "$(git rev-parse --show-toplevel)"
lib=workflow/library
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

content=ephemeral/projects/tractor/living-instructions/chapters/04-library/content
for p in promises elicit-then-prune question-files promise-adjacent-seams \
         proof-not-theater vertical-slices chapter-doc sprint-doc pyramid-index ledger; do
  f="$lib/doctrine/$p.md"
  test -f "$f" || { echo "missing $f"; exit 1; }
  cmp -s "$content/doctrine/$p.md" "$f" || { echo "$f differs from the planner's content/doctrine/$p.md"; exit 1; }
  grep -q '^Source:' "$f" || { echo "$f lacks a Source: line"; exit 1; }
  echo "doctrine: $p.md installed unchanged"
done
for t in brief.md promises.md recommendation.md CHAPTER.md SPRINT.md ledger.md; do
  test -f "$lib/templates/$t" || { echo "missing $lib/templates/$t"; exit 1; }
  cmp -s "$content/templates/$t" "$lib/templates/$t" || { echo "$lib/templates/$t differs from the planner's content/templates/$t"; exit 1; }
  grep -rq "$t" "$lib/prompts" || { echo "no prompt cites templates/$t"; exit 1; }
  echo "skeleton: $t installed unchanged and cited"
done

# Rendering and doctrine citation are the tests' job; they must run.
go test -v -run 'TestLibraryNoOrphans|TestLibraryRendersAll' ./workflow/ -count=1 > "$tmp/test.log" 2>&1 || { cat "$tmp/test.log"; exit 1; }
grep -q -- '--- PASS: TestLibraryRendersAll' "$tmp/test.log" || { echo "TestLibraryRendersAll did not run"; exit 1; }
grep -q -- '--- PASS: TestLibraryNoOrphans' "$tmp/test.log" || { echo "TestLibraryNoOrphans did not run"; exit 1; }
echo "doctrine-pages.sh: ok"
