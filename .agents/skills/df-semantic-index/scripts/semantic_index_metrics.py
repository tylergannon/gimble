#!/usr/bin/env python3
"""Report cheap structural metrics for a filesystem semantic index."""

from __future__ import annotations

import argparse
import json
import re
from collections import Counter, defaultdict
from datetime import datetime, timezone
from pathlib import Path
from statistics import mean, variance


TEXT_EXTENSIONS = {
    ".adoc",
    ".csv",
    ".html",
    ".json",
    ".jsonl",
    ".md",
    ".rst",
    ".sqlite",
    ".toml",
    ".tsv",
    ".txt",
    ".xml",
    ".yaml",
    ".yml",
}
ROUTING_NAMES = {
    "readme.md",
    "index.md",
    "semantic_index.md",
    "semantic-index.md",
    "taxonomy.md",
    "topics.md",
    "themes.md",
    "routes.md",
    "routes.yaml",
    "routes.yml",
    "manifest.json",
}
CITATION_RE = re.compile(
    r"(?:(?:[\w./@+-]+\.(?:md|txt|rst|py|js|ts|tsx|go|rs|ex|exs|java|c|cc|cpp|h|hpp|html|pdf|docx?))(?::L?\d+(?:-L?\d+)?|\s+p{1,2}\.?\s*\d+(?:-\d+)?)|https?://\S+)"
)
DEBT_RE = re.compile(r"\b(TODO|FIXME|debt|stale|orphan|manual review|unknown|blocked)\b", re.IGNORECASE)
LINK_RE = re.compile(r'\[.*?\]\(([^)#?]+)\)')


def utc_now() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat()


def safe_read(path: Path, limit: int = 2_000_000) -> str:
    try:
        if path.stat().st_size > limit:
            return ""
        return path.read_text(errors="ignore")
    except OSError:
        return ""


def all_files(root: Path) -> list[Path]:
    if not root.exists():
        return []
    return sorted(path for path in root.rglob("*") if path.is_file())


def depth_for(root: Path, path: Path) -> int:
    try:
        return len(path.relative_to(root).parts)
    except ValueError:
        return 0


def source_metrics(source: Path | None) -> dict[str, object]:
    if source is None:
        return {}
    files = all_files(source)
    by_ext = Counter(path.suffix.lower() or "[none]" for path in files)
    return {
        "path": str(source),
        "exists": source.exists(),
        "file_count": len(files),
        "extension_counts": dict(sorted(by_ext.items())),
    }


def _collect_outbound_refs(routing_text: str, routing_file: Path, index_root: Path) -> set[str]:
    """Return relative-to-index_root paths of files linked from a routing file."""
    refs: set[str] = set()
    for raw in LINK_RE.findall(routing_text):
        raw = raw.strip()
        if raw.startswith("http"):
            continue
        resolved = (routing_file.parent / raw).resolve()
        try:
            rel = resolved.relative_to(index_root).as_posix()
            refs.add(rel)
        except ValueError:
            pass
    # Also capture bare relative paths that look like index-internal paths.
    for m in re.finditer(r'`([^`]+\.(md|yaml|yml|json|jsonl|txt|sqlite))`', routing_text):
        raw = m.group(1).strip()
        resolved = (routing_file.parent / raw).resolve()
        try:
            rel = resolved.relative_to(index_root).as_posix()
            refs.add(rel)
        except ValueError:
            pass
    return refs


def _fan_out_stats(index: Path, files: list[Path]) -> dict[str, object]:
    """Compute fan-out (children per directory) across the index tree."""
    dir_children: Counter[str] = Counter()
    for path in files:
        rel = path.relative_to(index)
        parts = rel.parts
        for i in range(len(parts) - 1):
            parent = "/".join(parts[:i + 1]) if i > 0 else "."
            dir_children[parent] += 1
    # Also count direct children of root
    for path in files:
        rel = path.relative_to(index)
        if len(rel.parts) == 1:
            dir_children["."] = dir_children.get(".", 0)  # ensure root present
    counts = list(dir_children.values())
    if not counts:
        return {"mean": 0, "max": 0, "variance": 0, "skewed_nodes": []}
    mean_fo = round(mean(counts), 2)
    var_fo = round(variance(counts), 2) if len(counts) > 1 else 0
    max_fo = max(counts)
    # Nodes with > 2× mean fan-out are potentially overfull.
    threshold = max(mean_fo * 2, 10)
    skewed = sorted(
        (node for node, c in dir_children.items() if c > threshold),
        key=lambda n: -dir_children[n],
    )[:10]
    return {"mean": mean_fo, "max": max_fo, "variance": var_fo, "skewed_nodes": skewed}


def index_metrics(index: Path) -> dict[str, object]:
    files = all_files(index)
    by_ext = Counter(path.suffix.lower() or "[none]" for path in files)
    depths = [depth_for(index, path) for path in files]
    breadth: dict[int, set[str]] = defaultdict(set)
    for path in files:
        rel = path.relative_to(index)
        parts = rel.parts
        for level in range(1, len(parts) + 1):
            breadth[level].add("/".join(parts[:level]))

    routing_files: list[Path] = []
    likely_leaves: list[Path] = []
    citation_count = 0
    debt_count = 0
    benchmark_files: list[Path] = []
    newest_benchmark_mtime = None
    all_routing_refs: set[str] = set()

    for path in files:
        rel = path.relative_to(index).as_posix()
        lower_name = path.name.lower()
        lower_rel = rel.lower()
        is_routing = lower_name in ROUTING_NAMES or lower_rel.startswith(".semantic-index/")
        if is_routing:
            routing_files.append(path)
        elif path.suffix.lower() in TEXT_EXTENSIONS:
            likely_leaves.append(path)

        if "benchmark" in lower_rel or "eval" in lower_rel:
            benchmark_files.append(path)
            mtime = path.stat().st_mtime
            newest_benchmark_mtime = mtime if newest_benchmark_mtime is None else max(newest_benchmark_mtime, mtime)

        if path.suffix.lower() in TEXT_EXTENSIONS:
            text = safe_read(path)
            if text:
                citation_count += len(CITATION_RE.findall(text))
                debt_count += len(DEBT_RE.findall(text))
                if is_routing:
                    all_routing_refs |= _collect_outbound_refs(text, path, index)

    # Orphan detection: leaf files not referenced from any routing file.
    orphan_leaves: list[str] = []
    for leaf in likely_leaves:
        rel = leaf.relative_to(index).as_posix()
        leaf_name = leaf.name
        if rel not in all_routing_refs and leaf_name not in all_routing_refs:
            orphan_leaves.append(rel)

    leaf_depths = [depth_for(index, leaf) for leaf in likely_leaves]
    max_depth = max(depths, default=0)
    mean_leaf_d = round(mean(leaf_depths), 2) if leaf_depths else 0
    balance_ratio = round(max_depth / mean_leaf_d, 2) if mean_leaf_d else 0

    state_path = index / ".semantic-index" / "state.json"
    state = None
    if state_path.exists():
        try:
            state = json.loads(state_path.read_text())
        except json.JSONDecodeError:
            state = {"error": "state.json is not valid JSON"}

    routing_file_rels = [p.relative_to(index).as_posix() for p in routing_files]
    likely_leaf_rels = [p.relative_to(index).as_posix() for p in likely_leaves]

    return {
        "path": str(index),
        "exists": index.exists(),
        "file_count": len(files),
        "extension_counts": dict(sorted(by_ext.items())),
        "max_file_depth": max_depth,
        "mean_leaf_depth": mean_leaf_d,
        "balance_ratio": balance_ratio,
        "breadth_by_level": {str(level): len(nodes) for level, nodes in sorted(breadth.items())},
        "fan_out": _fan_out_stats(index, files),
        "routing_file_count": len(routing_files),
        "routing_files": routing_file_rels[:40],
        "likely_leaf_count": len(likely_leaves),
        "sample_likely_leaves": likely_leaf_rels[:40],
        "orphan_leaf_count": len(orphan_leaves),
        "orphan_leaves": orphan_leaves[:20],
        "citation_count": citation_count,
        "citation_density_per_leaf": round(citation_count / len(likely_leaves), 2) if likely_leaves else 0,
        "routing_density": round(citation_count / len(routing_files), 2) if routing_files else 0,
        "debt_marker_count": debt_count,
        "benchmark_file_count": len(benchmark_files),
        "last_benchmark_at": (
            datetime.fromtimestamp(newest_benchmark_mtime, timezone.utc).replace(microsecond=0).isoformat()
            if newest_benchmark_mtime is not None
            else None
        ),
        "state": state,
    }


def write_state(index: Path, report: dict[str, object], mode: str) -> Path:
    state_dir = index / ".semantic-index"
    state_dir.mkdir(parents=True, exist_ok=True)
    state_path = state_dir / "state.json"
    state = {
        "last_housekeeping_at": utc_now(),
        "last_mode": mode,
        "source": report.get("source", {}),
        "index_metrics": report["index"],
    }
    state_path.write_text(json.dumps(state, indent=2, sort_keys=True) + "\n")
    return state_path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--index", required=True, help="Semantic index root to inspect.")
    parser.add_argument("--source", help="Optional source corpus root.")
    parser.add_argument("--mode", default="audit", help="Maintenance mode for state writes.")
    parser.add_argument("--write-state", action="store_true", help="Write .semantic-index/state.json.")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    index = Path(args.index).expanduser().resolve()
    source = Path(args.source).expanduser().resolve() if args.source else None
    report: dict[str, object] = {
        "generated_at": utc_now(),
        "mode": args.mode,
        "source": source_metrics(source),
        "index": index_metrics(index),
    }
    if args.write_state:
        state_path = write_state(index, report, args.mode)
        report["state_written"] = str(state_path)
    print(json.dumps(report, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
