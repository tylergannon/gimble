#!/usr/bin/env python3
"""
Run retrieval benchmarks against a semantic index eval set.

Reads .semantic-index/evals.jsonl (or --evals <path>) and simulates agent
routing through the index to score retrieval quality.

Eval record format (one JSON object per line):
  {
    "id": "q001",
    "query": "how to handle AcroForm appearance streams?",
    "expected_routes": ["topics.md", "themes.md"],      // routing files expected to route here
    "expected_citations": ["repos/apache-pdfbox.md"],   // leaf files or citation targets expected
    "acceptable_families": ["repos/"],                  // path prefixes that count as a hit
    "notes": "optional free-text"
  }

Outputs aggregate metrics and per-query results to stdout (JSON) and writes
a timestamped run file to .semantic-index/benchmarks/.
"""

from __future__ import annotations

import argparse
import json
import re
from collections import deque
from datetime import datetime, timezone
from pathlib import Path


TEXT_EXTENSIONS = {".md", ".yaml", ".yml", ".json", ".jsonl", ".txt", ".rst", ".toml", ".xml"}
ROUTING_NAMES = {
    "readme.md", "index.md", "taxonomy.md", "topics.md", "themes.md",
    "routes.md", "routes.yaml", "routes.yml", "manifest.json",
    "semantic_index.md", "semantic-index.md",
}
LINK_RE = re.compile(r'\[.*?\]\(([^)#?]+)\)')


def utc_now() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat()


def safe_read(path: Path, limit: int = 500_000) -> str:
    try:
        if path.stat().st_size > limit:
            return path.read_text(errors="ignore")[:limit]
        return path.read_text(errors="ignore")
    except OSError:
        return ""


def find_entrypoint(index: Path) -> Path | None:
    for name in ("README.md", "readme.md", "INDEX.md", "index.md", "TAXONOMY.md"):
        candidate = index / name
        if candidate.exists():
            return candidate
    return None


def resolve_link(raw: str, from_file: Path, index: Path) -> Path | None:
    raw = raw.strip()
    if raw.startswith("http"):
        return None
    resolved = (from_file.parent / raw).resolve()
    if resolved.is_file() and resolved.suffix in TEXT_EXTENSIONS:
        try:
            resolved.relative_to(index)
            return resolved
        except ValueError:
            pass
    return None


def bfs_route(
    start: Path,
    index: Path,
    query_terms: set[str],
    max_hops: int = 6,
    max_files: int = 30,
) -> dict[str, object]:
    """
    BFS from entrypoint through linked files.

    Returns:
      route_path: ordered list of (rel_path, term_score) visited
      files_read: count of files opened
      term_hits_by_file: {rel_path: score}
    """
    visited: dict[str, int] = {}  # rel_path → hop_depth
    queue: deque[tuple[Path, int]] = deque()
    queue.append((start, 0))
    route_path: list[dict] = []
    files_read = 0

    while queue and files_read < max_files:
        path, depth = queue.popleft()
        rel = path.relative_to(index).as_posix()
        if rel in visited:
            continue
        visited[rel] = depth
        files_read += 1

        text = safe_read(path)
        term_score = sum(1 for t in query_terms if t in text.lower()) if text else 0
        route_path.append({"file": rel, "depth": depth, "term_score": term_score})

        if depth < max_hops and text:
            for raw in LINK_RE.findall(text):
                linked = resolve_link(raw, path, index)
                if linked and linked.relative_to(index).as_posix() not in visited:
                    queue.append((linked, depth + 1))

    return {
        "route_path": route_path,
        "files_read": files_read,
        "term_hits_by_file": {r["file"]: r["term_score"] for r in route_path},
    }


def score_query(entry: dict, index: Path) -> dict:
    query: str = entry.get("query", "")
    expected_routes: list[str] = entry.get("expected_routes", [])
    expected_citations: list[str] = entry.get("expected_citations", [])
    acceptable_families: list[str] = entry.get("acceptable_families", [])

    query_terms = {t.lower() for t in re.findall(r'\w{4,}', query)}

    entrypoint = find_entrypoint(index)
    if entrypoint is None:
        return {
            "id": entry.get("id", "?"),
            "query": query,
            "error": "no_entrypoint",
            "success_at_1": False,
            "success_at_3": False,
            "citation_precision": 0.0,
            "route_found": False,
            "files_read": 0,
        }

    traversal = bfs_route(entrypoint, index, query_terms)
    visited_files = set(traversal["term_hits_by_file"].keys())
    files_read = traversal["files_read"]

    # Route hit: did any expected routing file appear in the traversal?
    route_hit = any(
        any(r in vf or vf.endswith(r) for vf in visited_files)
        for r in expected_routes
    ) if expected_routes else bool(visited_files)

    # Citation precision: how many expected citations are reachable?
    hits = 0
    for citation in expected_citations:
        citation_path = (index / citation).resolve()
        hit = (
            citation_path.exists()
            or any(vf == citation or vf.endswith(citation) for vf in visited_files)
        )
        if not hit and acceptable_families:
            hit = any(citation.startswith(fam) for fam in acceptable_families)
        if hit:
            hits += 1

    citation_precision = round(hits / max(len(expected_citations), 1), 2)
    success_at_1 = hits >= 1
    success_at_3 = hits >= min(3, len(expected_citations)) if expected_citations else success_at_1

    # Estimate tool calls: files that had term hits are the ones an agent would open.
    relevant_files = [f for f, s in traversal["term_hits_by_file"].items() if s > 0]
    tool_call_estimate = len(relevant_files) if relevant_files else files_read

    # Depth to first relevant hit
    path_length = None
    for step in traversal["route_path"]:
        if step["term_score"] > 0:
            path_length = step["depth"]
            break

    return {
        "id": entry.get("id", "?"),
        "query": query,
        "success_at_1": success_at_1,
        "success_at_3": success_at_3,
        "citation_precision": citation_precision,
        "route_found": route_hit,
        "files_read": files_read,
        "tool_call_estimate": tool_call_estimate,
        "path_length_to_first_hit": path_length,
        "expected_citations": expected_citations,
        "citations_matched": hits,
        "notes": entry.get("notes"),
    }


def load_evals(evals_path: Path) -> list[dict]:
    entries = []
    with evals_path.open(errors="ignore") as f:
        for i, line in enumerate(f):
            line = line.strip()
            if not line or line.startswith("//") or line.startswith("#"):
                continue
            try:
                entries.append(json.loads(line))
            except json.JSONDecodeError as exc:
                print(f"  [warn] skipping evals line {i+1}: {exc}", flush=True)
    return entries


def aggregate(results: list[dict]) -> dict:
    if not results:
        return {}
    n = len(results)
    success_1 = sum(1 for r in results if r.get("success_at_1")) / n
    success_3 = sum(1 for r in results if r.get("success_at_3")) / n
    precision = sum(r.get("citation_precision", 0) for r in results) / n
    route_hit = sum(1 for r in results if r.get("route_found")) / n
    tool_calls = [r.get("tool_call_estimate", 0) for r in results if r.get("tool_call_estimate") is not None]
    path_lengths = [r["path_length_to_first_hit"] for r in results if r.get("path_length_to_first_hit") is not None]
    return {
        "eval_count": n,
        "success_at_1": round(success_1, 3),
        "success_at_3": round(success_3, 3),
        "mean_citation_precision": round(precision, 3),
        "route_hit_rate": round(route_hit, 3),
        "mean_tool_call_estimate": round(sum(tool_calls) / len(tool_calls), 2) if tool_calls else None,
        "mean_path_length": round(sum(path_lengths) / len(path_lengths), 2) if path_lengths else None,
        "misses": [r["id"] for r in results if not r.get("success_at_1")],
    }


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--index", required=True, help="Semantic index root.")
    parser.add_argument(
        "--evals",
        help="Path to evals.jsonl (default: <index>/.semantic-index/evals.jsonl).",
    )
    parser.add_argument(
        "--max-hops", type=int, default=6,
        help="Max routing hops per query (default: 6).",
    )
    parser.add_argument(
        "--max-files", type=int, default=30,
        help="Max files opened per query traversal (default: 30).",
    )
    parser.add_argument(
        "--write-results", action="store_true",
        help="Write timestamped run to .semantic-index/benchmarks/.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    index = Path(args.index).expanduser().resolve()

    evals_path = Path(args.evals).expanduser().resolve() if args.evals else index / ".semantic-index" / "evals.jsonl"
    if not evals_path.exists():
        alt = index / "evals.md"
        print(json.dumps({"error": f"No eval file found at {evals_path} or {alt}. Create one first."}))
        return 1

    entries = load_evals(evals_path)
    if not entries:
        print(json.dumps({"error": "Eval file exists but contains no valid entries."}))
        return 1

    results = []
    for entry in entries:
        result = score_query(entry, index)
        results.append(result)

    agg = aggregate(results)
    report = {
        "run_at": utc_now(),
        "index": str(index),
        "evals_path": str(evals_path),
        "aggregate": agg,
        "results": results,
    }

    print(json.dumps(report, indent=2, sort_keys=True))

    if args.write_results:
        bench_dir = index / ".semantic-index" / "benchmarks"
        bench_dir.mkdir(parents=True, exist_ok=True)
        ts = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
        out_path = bench_dir / f"run-{ts}.json"
        out_path.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n")
        print(f"\n[benchmarks] Written to {out_path}", flush=True)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
