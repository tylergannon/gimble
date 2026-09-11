#!/usr/bin/env python3
"""Sprint ledger CLI.

Supports two ledger shapes:
- docs/sprints/ledger.yaml: a narrow YAML list of sprint records, including
  optional ``chapter`` fields.
- docs/sprints/ledger.tsv: the older TSV format used by early installs.

The YAML format is preferred when present because it can preserve chapter links.
"""

from __future__ import annotations

import argparse
import re
import sys
from datetime import datetime, timezone
from pathlib import Path


TSV_STATUSES = ["planned", "in_progress", "completed", "skipped"]
YAML_STATUSES = ["planned", "in-progress", "done", "abandoned"]
SPRINT_RE = re.compile(r"^SPRINT-(\d{4})$")
CHAPTER_RE = re.compile(r"^CHAPTER-\d{4}$")
FIELDS = ("id", "title", "status", "chapter", "executor", "created", "updated")
TSV_HEADER = "sprint_id\ttitle\tstatus\tcreated_at\tupdated_at"


def project_root() -> Path:
    """Resolve the project root independent of where this skill is installed.

    This helper acts on a project's docs/sprints ledger and is documented to run
    from the repo root, so resolve from the current working directory rather than
    from this file's location. That works whether the skill is installed at
    project scope (<project>/.agents/skills/...) or
    user scope (for example, `~/.agents/skills/...` or `~/.codex/skills/...`),
    where a __file__-relative guess would point outside the project. Walk up
    from cwd to the nearest ancestor with a .git or docs/ marker; fall back to
    cwd.
    """
    start = Path.cwd().resolve()
    for candidate in (start, *start.parents):
        if (candidate / ".git").exists() or (candidate / "docs").is_dir():
            return candidate
    return start


def default_ledger_path() -> Path:
    root = project_root()
    yaml_path = root / "docs" / "sprints" / "ledger.yaml"
    if yaml_path.exists():
        return yaml_path
    tsv_path = root / "docs" / "sprints" / "ledger.tsv"
    if tsv_path.exists():
        return tsv_path
    # Neither exists yet: default to the project's YAML ledger (preferred format)
    # rather than writing a stray ledger.tsv into the installed skill directory.
    return yaml_path


def now_iso() -> str:
    return datetime.now(timezone.utc).isoformat(timespec="seconds")


def yaml_escape(value: str) -> str:
    if value == "" or any(c in value for c in ":#\"'\n") or value.strip() != value:
        return '"' + value.replace("\\", "\\\\").replace('"', '\\"') + '"'
    return value


def yaml_unescape(value: str) -> str:
    value = value.strip()
    if len(value) >= 2 and value[0] == value[-1] and value[0] in ("'", '"'):
        inner = value[1:-1]
        if value[0] == '"':
            inner = inner.replace('\\"', '"').replace("\\\\", "\\")
        return inner
    return value


def parse_yaml(path: Path) -> list[dict[str, str]]:
    if not path.exists():
        return []
    entries: list[dict[str, str]] = []
    current: dict[str, str] | None = None
    for raw in path.read_text().splitlines():
        line = raw.rstrip()
        if not line or line.lstrip().startswith("#"):
            continue
        if line in ("sprints:", "sprints: []"):
            continue
        stripped = line.lstrip()
        if stripped.startswith("- "):
            if current is not None:
                entries.append(current)
            current = {}
            stripped = stripped[2:]
        if current is None:
            continue
        if ":" in stripped:
            key, _, value = stripped.partition(":")
            current[key.strip()] = yaml_unescape(value)
    if current is not None:
        entries.append(current)
    return entries


def write_yaml(path: Path, entries: list[dict[str, str]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    if not entries:
        path.write_text("sprints: []\n")
        return

    lines = ["sprints:"]
    for entry in entries:
        first = True
        for key in FIELDS:
            if key not in entry:
                continue
            prefix = "  - " if first else "    "
            lines.append(f"{prefix}{key}: {yaml_escape(str(entry[key]))}")
            first = False
    path.write_text("\n".join(lines) + "\n")


def parse_tsv(path: Path) -> list[dict[str, str]]:
    if not path.exists():
        return []
    lines = [line.strip() for line in path.read_text().splitlines() if line.strip()]
    if lines and lines[0] == TSV_HEADER:
        lines = lines[1:]
    entries = []
    for line in lines:
        parts = line.split("\t")
        if len(parts) != 5:
            raise ValueError(f"Invalid TSV line (expected 5 fields): {line}")
        sprint_id, title, status, created, updated = parts
        entries.append({
            "id": sprint_id.zfill(3),
            "title": title,
            "status": status,
            "created": created,
            "updated": updated,
        })
    return entries


def write_tsv(path: Path, entries: list[dict[str, str]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    lines = [TSV_HEADER]
    for entry in sorted(entries, key=sprint_sort_key):
        lines.append("\t".join([
            numeric_id(entry["id"]),
            entry.get("title", ""),
            entry.get("status", "planned"),
            entry.get("created", ""),
            entry.get("updated", ""),
        ]))
    path.write_text("\n".join(lines) + "\n")


def numeric_id(sprint_id: str) -> str:
    match = SPRINT_RE.match(sprint_id)
    if match:
        return match.group(1)[-3:]
    return sprint_id.zfill(3)


def full_id(sprint_id: str) -> str:
    if SPRINT_RE.match(sprint_id):
        return sprint_id
    return f"SPRINT-{int(sprint_id):04d}"


def sprint_sort_key(entry: dict[str, str]) -> int:
    value = entry.get("id", "")
    if SPRINT_RE.match(value):
        return int(value.split("-")[1])
    return int(value)


def extract_doc_title_and_chapter(path: Path) -> tuple[str, str | None]:
    title = f"Sprint {path.stem.replace('SPRINT-', '')}"
    chapter = None
    for line in path.read_text(errors="ignore").splitlines()[:40]:
        if line.startswith("# "):
            _, _, rest = line.partition(":")
            if rest:
                title = rest.strip()
        match = re.search(r"CHAPTER-\d{4}", line)
        if line.startswith("Chapter:") and match:
            chapter = match.group(0)
    return title, chapter


class Ledger:
    def __init__(self, path: Path | None = None):
        self.path = path or default_ledger_path()
        self.format = "yaml" if self.path.suffix in (".yaml", ".yml") else "tsv"
        self.entries: list[dict[str, str]] = []

    @property
    def statuses(self) -> list[str]:
        return YAML_STATUSES if self.format == "yaml" else TSV_STATUSES

    def canonical_id(self, sprint_id: str) -> str:
        return full_id(sprint_id) if self.format == "yaml" else numeric_id(sprint_id)

    def canonical_status(self, status: str) -> str:
        aliases = {
            "in_progress": "in-progress" if self.format == "yaml" else "in_progress",
            "in-progress": "in-progress" if self.format == "yaml" else "in_progress",
            "completed": "done" if self.format == "yaml" else "completed",
            "done": "done" if self.format == "yaml" else "completed",
            "skipped": "abandoned" if self.format == "yaml" else "skipped",
            "abandoned": "abandoned" if self.format == "yaml" else "skipped",
        }
        status = aliases.get(status, status)
        if status not in self.statuses:
            raise ValueError(f"Invalid status: {status}. Must be one of: {self.statuses}")
        return status

    def load(self) -> "Ledger":
        self.entries = parse_yaml(self.path) if self.format == "yaml" else parse_tsv(self.path)
        return self

    def save(self) -> None:
        self.entries.sort(key=sprint_sort_key)
        if self.format == "yaml":
            write_yaml(self.path, self.entries)
        else:
            write_tsv(self.path, self.entries)

    def find(self, sprint_id: str) -> dict[str, str] | None:
        sid = self.canonical_id(sprint_id)
        return next((entry for entry in self.entries if entry.get("id") == sid), None)

    def add(self, sprint_id: str | None, title: str, status: str = "planned", chapter: str | None = None) -> dict[str, str] | None:
        if sprint_id:
            sid = self.canonical_id(sprint_id)
        else:
            max_id = max((sprint_sort_key(entry) for entry in self.entries), default=0)
            sid = full_id(str(max_id + 1)) if self.format == "yaml" else str(max_id + 1).zfill(3)
        if self.find(sid):
            return None
        entry = {
            "id": sid,
            "title": title,
            "status": self.canonical_status(status),
            "created": now_iso(),
            "updated": now_iso(),
        }
        if chapter and self.format == "yaml":
            if not CHAPTER_RE.match(chapter):
                raise ValueError("chapter must look like CHAPTER-0001")
            entry["chapter"] = chapter
        self.entries.append(entry)
        return entry

    def set_status(self, sprint_id: str, status: str) -> bool:
        entry = self.find(sprint_id)
        if not entry:
            return False
        entry["status"] = self.canonical_status(status)
        entry["updated"] = now_iso()
        return True

    def set_chapter(self, sprint_id: str, chapter: str) -> bool:
        if self.format != "yaml":
            raise ValueError("chapter fields require a YAML sprint ledger")
        entry = self.find(sprint_id)
        if not entry:
            return False
        if chapter == "":
            entry.pop("chapter", None)
        else:
            if not CHAPTER_RE.match(chapter):
                raise ValueError("chapter must look like CHAPTER-0001, or '' to clear")
            entry["chapter"] = chapter
        entry["updated"] = now_iso()
        return True

    def current(self) -> dict[str, str] | None:
        wanted = "in-progress" if self.format == "yaml" else "in_progress"
        return next((entry for entry in self.entries if entry.get("status") == wanted), None)

    def next_planned(self) -> dict[str, str] | None:
        planned = [entry for entry in self.entries if entry.get("status") == "planned"]
        return min(planned, key=sprint_sort_key) if planned else None

    def sync_from_docs(self) -> tuple[int, int]:
        docs_dir = project_root() / "docs" / "sprints"
        if not docs_dir.exists():
            docs_dir = self.path.parent
        added = 0
        for doc in sorted(docs_dir.glob("SPRINT-*.md")):
            title, chapter = extract_doc_title_and_chapter(doc)
            sid = self.canonical_id(doc.stem.replace("SPRINT-", ""))
            entry = self.find(sid)
            if entry:
                if self.format == "yaml" and chapter and not entry.get("chapter"):
                    entry["chapter"] = chapter
                    entry["updated"] = now_iso()
                continue
            self.add(sid, title, chapter=chapter)
            added += 1
        return added, len(self.entries)


def print_entry(entry: dict[str, str], ledger: Ledger, include_started: bool = False) -> None:
    print(f"{entry.get('id')}\t{entry.get('title', '')}")
    print(f"  Doc: docs/sprints/{entry.get('id')}.md" if ledger.format == "yaml" else f"  Doc: docs/sprints/SPRINT-{entry.get('id')}.md")
    if entry.get("chapter"):
        print(f"  Chapter: {entry['chapter']}")
    if include_started:
        print(f"  Started: {entry.get('updated', '')}")


def main() -> int:
    parser = argparse.ArgumentParser(description="Manage sprint ledger")
    parser.add_argument("--ledger", type=Path, help="Path to ledger.yaml or ledger.tsv")
    subparsers = parser.add_subparsers(dest="command", required=True)

    add_parser = subparsers.add_parser("add", help="Add a new sprint")
    add_parser.add_argument("add_args", nargs="+", help="Either TITLE or SPRINT_ID TITLE")
    add_parser.add_argument("--status", default="planned")
    add_parser.add_argument("--chapter", help="Optional chapter id, e.g. CHAPTER-0001")

    start_parser = subparsers.add_parser("start", help="Mark sprint as in progress")
    start_parser.add_argument("sprint_id")

    complete_parser = subparsers.add_parser("complete", help="Mark sprint as complete")
    complete_parser.add_argument("sprint_id")

    skip_parser = subparsers.add_parser("skip", help="Mark sprint as skipped/abandoned")
    skip_parser.add_argument("sprint_id")

    status_parser = subparsers.add_parser("status", help="Update sprint status")
    status_parser.add_argument("sprint_id")
    status_parser.add_argument("new_status")

    set_chapter_parser = subparsers.add_parser("set-chapter", help="Set or clear a YAML sprint chapter")
    set_chapter_parser.add_argument("sprint_id")
    set_chapter_parser.add_argument("chapter", help="Chapter id like CHAPTER-0001, or '' to clear")

    subparsers.add_parser("next", help="Get next planned sprint")
    subparsers.add_parser("current", help="Get current in-progress sprint")
    subparsers.add_parser("stats", help="Show ledger statistics")

    list_parser = subparsers.add_parser("list", help="List sprints")
    list_parser.add_argument("--status", help="Filter by status")
    list_parser.add_argument("--chapter", help="Filter by chapter id")

    subparsers.add_parser("sync", help="Sync ledger from docs/sprints/*.md files")

    args = parser.parse_args()
    ledger = Ledger(args.ledger).load()

    try:
        if args.command == "add":
            if len(args.add_args) == 1:
                sprint_id = None
                title = args.add_args[0]
            else:
                sprint_id = args.add_args[0]
                title = " ".join(args.add_args[1:])
            entry = ledger.add(sprint_id, title, args.status, args.chapter)
            if not entry:
                print(f"Sprint {sprint_id} already exists", file=sys.stderr)
                return 1
            ledger.save()
            print(f"Added sprint {entry['id']}: {entry['title']} [{entry['status']}]")

        elif args.command == "start":
            if ledger.set_status(args.sprint_id, "in-progress"):
                ledger.save()
                print(f"Sprint {args.sprint_id} is now {ledger.canonical_status('in-progress')}")
            else:
                print(f"Sprint {args.sprint_id} not found", file=sys.stderr)
                return 1

        elif args.command == "complete":
            if ledger.set_status(args.sprint_id, "done"):
                ledger.save()
                print(f"Sprint {args.sprint_id} marked as {ledger.canonical_status('done')}")
            else:
                print(f"Sprint {args.sprint_id} not found", file=sys.stderr)
                return 1

        elif args.command == "skip":
            if ledger.set_status(args.sprint_id, "abandoned"):
                ledger.save()
                print(f"Sprint {args.sprint_id} marked as {ledger.canonical_status('abandoned')}")
            else:
                print(f"Sprint {args.sprint_id} not found", file=sys.stderr)
                return 1

        elif args.command == "status":
            if ledger.set_status(args.sprint_id, args.new_status):
                ledger.save()
                print(f"Sprint {args.sprint_id} status updated to '{ledger.canonical_status(args.new_status)}'")
            else:
                print(f"Sprint {args.sprint_id} not found", file=sys.stderr)
                return 1

        elif args.command == "set-chapter":
            if ledger.set_chapter(args.sprint_id, args.chapter):
                ledger.save()
                print(f"Sprint {args.sprint_id} chapter updated")
            else:
                print(f"Sprint {args.sprint_id} not found", file=sys.stderr)
                return 1

        elif args.command == "next":
            entry = ledger.next_planned()
            if not entry:
                print("No planned sprints", file=sys.stderr)
                return 1
            print_entry(entry, ledger)

        elif args.command == "current":
            entry = ledger.current()
            if not entry:
                print("No sprint currently in progress", file=sys.stderr)
                return 1
            print_entry(entry, ledger, include_started=True)

        elif args.command == "stats":
            counts = {status: 0 for status in ledger.statuses}
            for entry in ledger.entries:
                if entry.get("status") in counts:
                    counts[entry["status"]] += 1
            print(f"Total sprints: {len(ledger.entries)}")
            for status in ledger.statuses:
                print(f"  {status}: {counts[status]}")
            current = ledger.current()
            if current:
                print(f"\nCurrently working on: {current['id']} - {current.get('title', '')}")
            next_up = ledger.next_planned()
            if next_up:
                print(f"Next up: {next_up['id']} - {next_up.get('title', '')}")

        elif args.command == "list":
            rows = ledger.entries
            if args.status:
                wanted = ledger.canonical_status(args.status)
                rows = [entry for entry in rows if entry.get("status") == wanted]
            if args.chapter:
                rows = [entry for entry in rows if entry.get("chapter") == args.chapter]
            for entry in rows:
                chapter = f" {entry['chapter']}" if entry.get("chapter") else ""
                print(f"{entry.get('id')}: {entry.get('status')} {entry.get('title', '')}{chapter}")

        elif args.command == "sync":
            added, total = ledger.sync_from_docs()
            ledger.save()
            print(f"Synced: {added} new sprints added, {total} total in ledger")

    except ValueError as exc:
        print(str(exc), file=sys.stderr)
        return 1

    return 0


if __name__ == "__main__":
    sys.exit(main())
