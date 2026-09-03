#!/bin/sh
# Proves: doctrine pages and skeletons exist, are
# cited, and carry no template delimiters.
set -eu
cd "$(git rev-parse --show-toplevel)"
lib=workflow/library

for p in promises elicit-then-prune question-files promise-adjacent-seams \
         proof-not-theater vertical-slices chapter-doc sprint-doc pyramid-index ledger; do
  f="$lib/doctrine/$p.md"
  test -f "$f" || { echo "missing $f"; exit 1; }
  tail -n 3 "$f" | grep -q '^Source:' || { echo "$f lacks a Source: line"; exit 1; }
done
for t in brief.md promises.md recommendation.md CHAPTER.md SPRINT.md ledger.md; do
  test -f "$lib/templates/$t" || { echo "missing $lib/templates/$t"; exit 1; }
done
if grep -rl '{{' "$lib/doctrine" "$lib/templates"; then
  echo "template delimiters found in content"; exit 1
fi
go test -run 'TestLibraryNoOrphans|TestLibraryRendersAll' ./workflow/... -count=1
echo "doctrine-pages.sh: ok"
