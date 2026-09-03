#!/bin/sh
# Proves: every rendered prompt is library text plus data values; each
# node renders exactly its header file's include closure; every prompt
# and supervisor file is rendered by some node; an injected uncited page
# fails the orphan test by name; an injected broken action fails the
# render test by name; both tests run and pass in the real tree.
. "$(dirname "$0")/lib.sh"
build_copy

# ---- text from the closure only ----------------------------------------
# Remove every data value form (raw, Go-quoted, shell-quoted) from each
# rendered line and every template action or comment from each line of
# the node's closure; every rendered residue must equal some closure
# residue. Text parked in another file, in a comment, or supplied by Go
# fails; quoted values and repeated includes pass.
delim_open="$(sed -n 's/^[Dd]elimiters:[[:space:]]*\([^[:space:]]*\)[[:space:]]*\([^[:space:]]*\).*/\1/p' "$lib/README.md" | head -1)"
delim_close="$(sed -n 's/^[Dd]elimiters:[[:space:]]*\([^[:space:]]*\)[[:space:]]*\([^[:space:]]*\).*/\2/p' "$lib/README.md" | head -1)"
test -n "$delim_open" && test -n "$delim_close" || { echo "closure-text: README does not state both delimiters"; exit 1; }
projdir="$tmp/demo/ephemeral/projects/demo"
printf '%s\n' demo "$tmp/demo" "$tmp/bin/tractor" "$seed" "$projdir" "$projdir/brief.md" "$projdir/checklist.md" "$projdir/interview" "$tmp/bin/tractor ask" | sort -u > "$tmp/values-raw.txt"
python3 - "$tmp/values-raw.txt" > "$tmp/values-all.txt" <<'PYEOF'
import json, sys
vals = [l.rstrip("\n") for l in open(sys.argv[1]) if l.strip()]
out = set()
for v in vals:
    out.add(v)
    out.add(json.dumps(v))                       # Go quote for plain strings
    out.add("'" + v.replace("'", "'\\''") + "'")  # shell single-quoted
for v in sorted(out, key=len, reverse=True):
    print(v)
PYEOF
strip_values() {
  cmd="sed"
  while IFS= read -r v; do
    esc="$(printf '%s' "$v" | sed 's/[.[\*^$\/&|]/\\&/g')"
    cmd="$cmd -e 's|$esc||g'"
  done < "$tmp/values-all.txt"
  eval "$cmd"
}
strip_actions() { sed -e "s/$delim_open\/\*.*\*\/$delim_close//g" -e "s/$delim_open[^>]*$delim_close//g"; }
doc_closure() { # FILE... -> doctrine and template files named by any of them
  for f in "$@"; do
    for cand in $(cd "$lib" && find doctrine templates -type f 2>/dev/null); do
      base="$(basename "$cand" .md)"; base2="$(basename "$cand")"
      if grep -q "\"$base\"\|\"$base2\"\|\"$cand\"" "$lib/$f" 2>/dev/null; then printf '%s\n' "$cand"; fi
    done
  done | sort -u
}
for wf in $workflows; do
  yaml_nodes "$wf" | while read -r node kind; do
    file="$(shown_file "$tmp/bin/tractor" "$wf" "$node")"
    [ -n "$file" ] || continue
    closure "$file" > "$tmp/node-closure.txt"
    doc_closure $(cat "$tmp/node-closure.txt") >> "$tmp/node-closure.txt"
    : > "$tmp/closure-residue.txt"
    while read -r f; do strip_actions < "$lib/$f" >> "$tmp/closure-residue.txt"; done < "$tmp/node-closure.txt"
    awk 'length($0) > 0' "$tmp/closure-residue.txt" | sed 's/[[:space:]]*$//' | sort -u > "$tmp/closure-residue.txt.s"
    show_raw "$tmp/bin/tractor" "$wf" "$node" | strip_values | sed 's/[[:space:]]*$//' | awk 'length($0) > 0' | sort -u > "$tmp/rendered-residue.txt"
    while IFS= read -r line; do
      grep -qxF -- "$line" "$tmp/closure-residue.txt.s" \
        || { echo "closure-text: $wf/$node renders a line that is neither closure text nor data: $line"; exit 1; }
    done < "$tmp/rendered-residue.txt"
  done
done

# ---- content: one sentinel per file, in the copy ------------------------
mlib="$tmp/src/workflow/library"
stamp="SENTINEL-$(date +%s)-$$"
find "$mlib/prompts" "$mlib/supervisors" "$mlib/passes" -type f 2>/dev/null | while read -r f; do
  rel="${f#$mlib/}"
  printf '\n%s FILE %s\n' "$stamp" "$rel" >> "$f"
done
page="$(find "$mlib/doctrine" -type f -name '*.md' 2>/dev/null | sort -R | head -1 || true)"
if [ -n "$page" ]; then printf '\n%s DOCTRINE %s\n' "$stamp" "$(basename "$page" .md)" >> "$page"; fi
orphan="probe-$(od -An -N4 -tx1 /dev/urandom | tr -d ' \n')"
printf '# uncited\n\nNo prompt cites this page.\n' > "$mlib/doctrine/$orphan.md"
(cd "$tmp/src" && go build -o "$tmp/bin/mutated" ./cmd/tractor)
closure_lib="$mlib"
: > "$tmp/used.txt"
for wf in $workflows; do
  yaml_nodes "$wf" | while read -r node kind; do
    file="$(shown_file "$tmp/bin/mutated" "$wf" "$node")"
    case "$kind" in
      codergen|supervisor) test -n "$file" || { echo "content: $wf/$node header names no library file"; exit 1; } ;;
    esac
    [ -n "$file" ] || continue
    show_raw "$tmp/bin/mutated" "$wf" "$node" > "$tmp/mut.txt"
    grep -q "^$stamp FILE $file\$" "$tmp/mut.txt" || { echo "content: $wf/$node does not render $file"; exit 1; }
    closure "$file" > "$tmp/closure.txt"
    cat "$tmp/closure.txt" >> "$tmp/used.txt"
    grep "^$stamp FILE " "$tmp/mut.txt" | sed "s/^$stamp FILE //" | while read -r seen; do
      grep -qxF -- "$seen" "$tmp/closure.txt" || { echo "content: $wf/$node renders $seen, outside the include closure of $file"; exit 1; }
    done
    # A conditional include may not fire for the check's parameters, so a
    # closure file's sentinel is allowed to be absent; a file outside the
    # closure is not allowed to appear.
    if [ -n "$page" ] && grep -q "$(basename "$page" .md)" "$mlib/$file"; then
      grep -q "^$stamp DOCTRINE " "$tmp/mut.txt" || { echo "content: $wf/$node names $(basename "$page") but does not render it"; exit 1; }
    fi
    echo "content: $wf/$node renders $file and nothing outside its closure"
  done
done

# ---- no unused prompt file --------------------------------------------
sort -u "$tmp/used.txt" -o "$tmp/used.txt"
(cd "$mlib" && find prompts supervisors -type f | sort) | while read -r f; do
  grep -qxF -- "$f" "$tmp/used.txt" || { echo "unused: no node renders $f"; exit 1; }
done

# ---- orphans: the injected page must be reported by name ---------------
(cd "$tmp/src" && go test -run 'TestLibraryNoOrphans' ./workflow/ -count=1 > "$tmp/orphan.log" 2>&1) \
  && { echo "orphans: injected $orphan.md was not reported"; cat "$tmp/orphan.log"; exit 1; }
grep -q "$orphan" "$tmp/orphan.log" || { echo "orphans: test failed but did not name $orphan.md"; cat "$tmp/orphan.log"; exit 1; }
echo "orphans: $orphan.md reported"
# A second probe a name-recognising test cannot pass: remove every citation
# of one existing page (chosen at random) from the prompt files and require
# the test to fail naming that page; then put the files back.
rm -f "$mlib/doctrine/$orphan.md"
victim="$(find "$mlib/doctrine" -type f -name '*.md' 2>/dev/null | sort -R | head -1 || true)"
if [ -n "$victim" ]; then
  vbase="$(basename "$victim" .md)"
  mkdir -p "$tmp/keep"
  for f in $(grep -rl "\"$vbase\"" "$mlib/prompts" "$mlib/supervisors" "$mlib/passes" 2>/dev/null); do
    cp "$f" "$tmp/keep/$(echo "$f" | tr '/' '_')"
    sed -i '' -e "/\"$vbase\"/d" "$f"
  done
  (cd "$tmp/src" && go test -run 'TestLibraryNoOrphans' ./workflow/ -count=1 > "$tmp/orphan2.log" 2>&1) \
    && { echo "orphans: uncited existing page $vbase.md was not reported"; cat "$tmp/orphan2.log"; exit 1; }
  grep -q "$vbase" "$tmp/orphan2.log" || { echo "orphans: test failed but did not name $vbase.md"; cat "$tmp/orphan2.log"; exit 1; }
  echo "orphans: uncited existing page $vbase.md reported"
  for k in "$tmp/keep"/*; do [ -f "$k" ] || continue; cp "$k" "$(basename "$k" | tr '_' '/')"; done
fi

# ---- rendering: a broken action must be reported by name ---------------
if [ -n "$page" ]; then
  printf '\n%s broken-%s\n' "$delim_open" "$(od -An -N4 -tx1 /dev/urandom | tr -d ' \n')" >> "$page"
  (cd "$tmp/src" && go test -run 'TestLibraryRendersAll' ./workflow/ -count=1 > "$tmp/render.log" 2>&1) \
    && { echo "rendering: broken page was not reported"; cat "$tmp/render.log"; exit 1; }
  grep -q "$(basename "$page")" "$tmp/render.log" || { echo "rendering: test failed but did not name $(basename "$page")"; cat "$tmp/render.log"; exit 1; }
  echo "rendering: broken action in $(basename "$page") reported"
fi

# ---- the real tree's tests run and pass --------------------------------
go test -v -run 'TestLibraryRendersAll|TestLibraryNoOrphans' ./workflow/ -count=1 > "$tmp/test.log" 2>&1 || { cat "$tmp/test.log"; exit 1; }
grep -q -- '--- PASS: TestLibraryRendersAll' "$tmp/test.log" || { echo "TestLibraryRendersAll did not run"; exit 1; }
grep -q -- '--- PASS: TestLibraryNoOrphans' "$tmp/test.log" || { echo "TestLibraryNoOrphans did not run"; exit 1; }
echo "orphan-walk-and-render.sh: ok"
