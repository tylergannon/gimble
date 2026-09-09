"""Audit this research index and copied-source integrity; never execute the corpus."""
from pathlib import Path
import hashlib
import json
import re
from datetime import datetime, timezone

root = Path(__file__).resolve().parents[1]
index, corpus = root / "index", root / "corpus"
state_dir = index / ".semantic-index"
state_dir.mkdir(exist_ok=True)
issues = []
files = sorted(p for p in corpus.rglob("*") if p.is_file())
hashes = {str(p.relative_to(corpus)): hashlib.sha256(p.read_bytes()).hexdigest() for p in files}
fingerprint = hashlib.sha256(json.dumps(hashes, sort_keys=True).encode()).hexdigest()

def check_entries(value, family):
    if isinstance(value, dict):
        raw = value.get("local_path", value.get("path"))
        if raw and "sha256" in value:
            candidates = [corpus / raw, corpus / family / raw]
            found = next((p for p in candidates if p.is_file()), None)
            if found is None:
                issues.append(f"missing manifest file: {family}/{raw}")
            elif hashlib.sha256(found.read_bytes()).hexdigest() != value["sha256"]:
                issues.append(f"hash mismatch: {found.relative_to(corpus)}")
        for child in value.values():
            check_entries(child, family)
    elif isinstance(value, list):
        for child in value:
            check_entries(child, family)

for manifest in corpus.glob("*/manifest.json"):
    check_entries(json.loads(manifest.read_text()), manifest.parent.name)
durable_manifest = corpus / "durable/MANIFEST.md"
if durable_manifest.exists():
    family_hashes = {v for k, v in hashes.items() if k.startswith("durable/")}
    for expected in re.findall(r"`([0-9a-f]{64})`", durable_manifest.read_text()):
        if expected not in family_hashes:
            issues.append(f"durable manifest hash absent: {expected}")

leaves = sorted((index / "leaves").glob("*.md"))
routes = sorted((index / "routes").glob("*.md"))
link_graph = {}
for doc in [index / "README.md", *routes, *leaves]:
    links = []
    for target in re.findall(r"\]\(([^)]+)\)", doc.read_text()):
        if "://" in target or target.startswith("#"):
            continue
        dest = (doc.parent / target.split("#", 1)[0]).resolve()
        # Generated state and benchmark outputs are checked after writing.
        if not dest.exists() and ".semantic-index/" not in str(dest):
            issues.append(f"broken link: {doc.relative_to(root)} -> {target}")
        links.append(dest)
    link_graph[doc.resolve()] = links

seen, pending = set(), [index / "README.md"]
while pending:
    node = pending.pop().resolve()
    if node in seen:
        continue
    seen.add(node)
    pending.extend(link_graph.get(node, []))
orphans = [str(p.relative_to(index)) for p in leaves if p.resolve() not in seen]
issues.extend("orphan: " + p for p in orphans)

citation_count = 0
pattern = r"(?:corpus/)?((?:concepts|durable|imperative|static|method)/[\w./-]+):(?:L)?(\d+)(?:-(?:L)?(\d+))?"
for leaf in leaves:
    for rel, start, end in re.findall(pattern, leaf.read_text()):
        citation_count += 1
        target = corpus / rel
        if not target.is_file():
            issues.append(f"missing citation: {leaf.name} -> {rel}:{start}")
        else:
            count = len(target.read_text().splitlines())
            if int(start) < 1 or int(end or start) > count:
                issues.append(f"out-of-range citation: {leaf.name} -> {rel}:{start}-{end}; lines={count}")

evals = [json.loads(line) for line in (state_dir / "evals.jsonl").read_text().splitlines() if line]
structural = []
for item in evals:
    wanted = item["expected_routes"] + item["expected_citations"]
    misses = [p for p in wanted if (index / p).resolve() not in seen]
    structural.append({"id": item["id"], "reachable": not misses, "missing": misses})
issues.extend(f"eval target missing: {x['id']}: {x['missing']}" for x in structural if x['missing'])

if list(corpus.rglob("*.go")):
    issues.append("raw .go files could affect parent Go package discovery")
state = {
    "built_at": datetime.now(timezone.utc).isoformat(), "mode": "scratch-build",
    "corpus": "../corpus", "corpus_file_count": len(files),
    "corpus_bytes": sum(p.stat().st_size for p in files), "corpus_sha256": fingerprint,
    "leaf_count": len(leaves), "route_count": len(routes), "citation_count": citation_count,
    "orphan_leaf_count": len(orphans), "index_leaf_depth": 2,
    "structural_evals": structural,
    "semantic_benchmark": "See retrieval-review.md; reachability here is not relevance or a latency benchmark.",
    "issues": issues,
}
(state_dir / "state.json").write_text(json.dumps(state, indent=2) + "\n")
(state_dir / "corpus-sha256.json").write_text(json.dumps(hashes, indent=2) + "\n")
print(json.dumps(state, indent=2))
raise SystemExit(bool(issues))
