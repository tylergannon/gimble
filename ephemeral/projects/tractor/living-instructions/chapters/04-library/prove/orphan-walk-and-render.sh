#!/bin/sh
# Proves: every rendered prompt is library text plus data values; each
# node renders exactly its header file's include closure; every prompt
# an injected uncited page
# fails the orphan test by name; an injected broken action fails the
# render test by name; both tests run and pass in the real tree.
. "$(dirname "$0")/lib.sh"
build_copy

# ---- Build adds nothing to the header file's rendering -----------------
# `show --raw` (which equals Build, per show-equals-build.sh) must equal
# the check's own standalone workflow.Render of the node's header file.
for wf in $workflows; do
  built_ids "$wf" | while read -r node; do
    file="$(shown_file "$tmp/bin/tractor" "$wf" "$node")"
    [ -n "$file" ] || continue
    "$tmp/bin/builddump" render "$file" "$tmp/demo" "$tmp/bin/tractor" "$seed" > "$tmp/render.txt" 2>"$tmp/render.err" \
      || { echo "render: $file does not render standalone"; cat "$tmp/render.err"; exit 1; }
    show_raw "$tmp/bin/tractor" "$wf" "$node" > "$tmp/shown.txt"
    cmp -s "$tmp/render.txt" "$tmp/shown.txt" || { echo "render: $wf/$node differs from Render($file)"; diff "$tmp/render.txt" "$tmp/shown.txt" | head -20; exit 1; }
    echo "render: $wf/$node equals Render($file)"
  done
done
delim_open="$(sed -n 's/^[Dd]elimiters:[[:space:]]*\([^[:space:]]*\)[[:space:]]*\([^[:space:]]*\).*/\1/p' "$lib/README.md" | head -1)"
delim_close="$(sed -n 's/^[Dd]elimiters:[[:space:]]*\([^[:space:]]*\)[[:space:]]*\([^[:space:]]*\).*/\2/p' "$lib/README.md" | head -1)"
test -n "$delim_open" && test -n "$delim_close" || { echo "README does not state both delimiters"; exit 1; }

# ---- content: one sentinel per file, in the copy ------------------------
mlib="$tmp/src/workflow/library"
stamp="SENTINEL-$(date +%s)-$$"
find "$mlib/prompts" "$mlib/supervisors" "$mlib/passes" "$mlib/templates" -type f 2>/dev/null | while read -r f; do
  rel="${f#$mlib/}"
  printf '\n%s FILE %s\n' "$stamp" "$rel" >> "$f"
done
# When the library has no doctrine page yet (sprint 3 runs before sprint
# 4), plant one and cite it from one prompt file, so every probe below
# has a page to work on; otherwise pick an existing page at random.
if [ -z "$(find "$mlib/doctrine" -type f -name '*.md' 2>/dev/null)" ]; then
  planted="planted-$(od -An -N4 -tx1 /dev/urandom | tr -d ' \n')"
  printf '# planted\n\nA page the proof planted.\n' > "$mlib/doctrine/$planted.md"
  host="$(find "$mlib/prompts" -type f | sort | head -1)"
  printf '\n%s doctrine "%s" %s\n' "$delim_open" "$planted" "$delim_close" >> "$host"
  echo "planted: $planted.md cited from ${host#$mlib/}"
fi
page="$(find "$mlib/doctrine" -type f -name '*.md' 2>/dev/null | sort -R | head -1 || true)"
orphan="probe-$(od -An -N4 -tx1 /dev/urandom | tr -d ' \n')"
printf '# uncited\n\nNo prompt cites this page.\n' > "$mlib/doctrine/$orphan.md"
(cd "$tmp/src" && go build -o "$tmp/bin/mutated" ./cmd/tractor)
closure_lib="$mlib"
for wf in $workflows; do
  built_ids "$wf" | while read -r node; do
    file="$(shown_file "$tmp/bin/mutated" "$wf" "$node")"
    # A node whose payload is a prompt (Build prints one that is not a
    # command or checklist path) must name its library file.
    if "$tmp/bin/builddump" "$wf" "$node" "$tmp/demo" "$tmp/bin/tractor" "$([ "$wf" = plan ] && printf '%s' "$seed")" 2>/dev/null | grep -q '[[:space:]]'; then
      case "$(yaml_nodes "$wf" | awk -v n="$node" '$1==n{print $2}')" in
        tool|loop) ;;
        *) test -n "$file" || { echo "content: $wf/$node header names no library file"; exit 1; } ;;
      esac
    fi
    [ -n "$file" ] || continue
    show_raw "$tmp/bin/mutated" "$wf" "$node" > "$tmp/mut.txt"
    grep -q "^$stamp FILE $file\$" "$tmp/mut.txt" || { echo "content: $wf/$node does not render $file"; exit 1; }
    closure "$file" > "$tmp/closure.txt"
    grep "^$stamp FILE " "$tmp/mut.txt" | sed "s/^$stamp FILE //" | while read -r seen; do
      grep -qxF -- "$seen" "$tmp/closure.txt" || { echo "content: $wf/$node renders $seen, outside the include closure of $file"; exit 1; }
    done
    # A conditional include may not fire for the check's parameters, so a
    # closure file's sentinel is allowed to be absent; a file outside the
    # closure is not allowed to appear.
    echo "content: $wf/$node renders $file and nothing outside its closure"
  done
done

# ---- every doctrine page renders under some representative parameter set --
# The README lists representative parameter sets as lines of the form
# "- params: --project X --seed Y" (sprint 3 writes them for the render
# test). A doctrine page whose sentinel appears in no node's output under
# any listed set (nor under the check's own demo parameters) is an orphan,
# whatever text names it.
find "$mlib/doctrine" -type f -name '*.md' 2>/dev/null | while read -r d; do
  printf '\n%s DOCTRINE %s\n' "$stamp" "$(basename "$d" .md)" >> "$d"
done
(cd "$tmp/src" && go build -o "$tmp/bin/mutated" ./cmd/tractor)
: > "$tmp/seen-doctrine.txt"
sed -n 's/^- params: *//p' "$mlib/README.md" > "$tmp/param-sets.txt"
printf -- '--project demo --seed %s\n' "$seed" >> "$tmp/param-sets.txt"
while read -r pset; do
  for wf in $workflows; do
    built_ids "$wf" | while read -r node; do
      # shellcheck disable=SC2086
      "$tmp/bin/mutated" workflow show "$wf" $pset --workdir "$tmp/demo" --node "$node" --raw 2>/dev/null \
        | grep "^$stamp DOCTRINE " | sed "s/^$stamp DOCTRINE //"
    done
  done
done < "$tmp/param-sets.txt" | sort -u > "$tmp/seen-doctrine.txt"
find "$mlib/doctrine" -type f -name '*.md' 2>/dev/null | while read -r d; do
  b="$(basename "$d" .md)"
  grep -qxF -- "$b" "$tmp/seen-doctrine.txt" || { echo "doctrine: no node renders $b.md under any listed parameter set"; exit 1; }
  echo "doctrine: $b.md rendered under a listed parameter set"
done

# ---- every skeleton is included by some rendered prompt ------------------
: > "$tmp/seen-templates.txt"
for wf in $workflows; do
  built_ids "$wf" | while read -r node; do
    file="$(shown_file "$tmp/bin/mutated" "$wf" "$node")"
    [ -n "$file" ] || continue
    show_raw "$tmp/bin/mutated" "$wf" "$node" | grep "^$stamp FILE templates/" | sed "s/^$stamp FILE //"
  done
done | sort -u > "$tmp/seen-templates.txt"
# The README lists the planner's skeletons (sprint 4 writes the list as
# lines of the form "- templates/<name>"); each listed one must be rendered
# by some node. Files under templates/ not in the list may exist unused.
sed -n 's/^- \(templates\/[^ ]*\).*/\1/p' "$mlib/README.md" | sort -u > "$tmp/listed-templates.txt"
while read -r t; do
  grep -qxF -- "$t" "$tmp/seen-templates.txt" || { echo "templates: no node renders listed skeleton $t"; exit 1; }
  echo "templates: $t rendered by a node"
done < "$tmp/listed-templates.txt"

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

# ---- keep the probe outputs for the judge ------------------------------
logdir="$root/ephemeral/projects/tractor/living-instructions/chapters/04-library/prove/last-run"
mkdir -p "$logdir"
for l in orphan.log orphan2.log render.log; do [ -f "$tmp/$l" ] && cp "$tmp/$l" "$logdir/probe-$l"; done

# ---- the real tree's tests run and pass --------------------------------
go test -v -run 'TestLibraryRendersAll|TestLibraryNoOrphans' ./workflow/ -count=1 > "$tmp/test.log" 2>&1 || { cat "$tmp/test.log"; exit 1; }
grep -q -- '--- PASS: TestLibraryRendersAll' "$tmp/test.log" || { echo "TestLibraryRendersAll did not run"; exit 1; }
grep -q -- '--- PASS: TestLibraryNoOrphans' "$tmp/test.log" || { echo "TestLibraryNoOrphans did not run"; exit 1; }
cp "$tmp/test.log" "$logdir/probe-test.log"
echo "orphan-walk-and-render.sh: ok"
