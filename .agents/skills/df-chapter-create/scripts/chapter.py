#!/usr/bin/env python3
"""Small chapter ledger helper for docs/chapters/ledger.yaml."""

from __future__ import annotations

import argparse
import datetime as dt
import re
import sys
from pathlib import Path

CHAPTER_RE = re.compile(r"^CHAPTER-(\d{4})$")
FIELDS = ("id", "title", "status", "doc", "created", "updated", "summary", "sprint_ids")
STATUSES = ("active", "paused", "done", "abandoned")


def project_root() -> Path:
    """Resolve the project root independent of where this skill is installed.

    This helper acts on a project's docs/chapters ledger and is documented to run
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


def ledger_path() -> Path:
    return project_root() / "docs" / "chapters" / "ledger.yaml"


def now() -> str:
    return dt.datetime.now(dt.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


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


def load() -> dict:
    path = ledger_path()
    if not path.exists():
        return {"chapters": []}

    lines = path.read_text().splitlines()
    chapters: list[dict] = []
    current: dict | None = None
    in_sprint_ids = False
    block_key: str | None = None
    block_lines: list[str] = []

    def flush_block() -> None:
        nonlocal block_key, block_lines
        if current is not None and block_key is not None:
            current[block_key] = " ".join(line.strip() for line in block_lines).strip()
        block_key = None
        block_lines = []

    for raw in lines:
        line = raw.rstrip()
        if not line or line.lstrip().startswith("#"):
            continue
        if line == "chapters:" or line == "chapters: []":
            continue

        stripped = line.lstrip()
        if block_key is not None and not stripped.startswith("- ") and ":" not in stripped:
            block_lines.append(stripped)
            continue
        flush_block()

        if stripped.startswith("- "):
            if in_sprint_ids and current is not None:
                current.setdefault("sprint_ids", []).append(yaml_unescape(stripped[2:]))
                continue
            if current is not None:
                chapters.append(current)
            current = {}
            in_sprint_ids = False
            stripped = stripped[2:]

        if current is None:
            continue
        if stripped == "sprint_ids:":
            current["sprint_ids"] = []
            in_sprint_ids = True
            continue
        if ":" in stripped:
            key, _, value = stripped.partition(":")
            key = key.strip()
            value = value.strip()
            if key == "summary" and value == ">-":
                current[key] = ""
                block_key = key
                block_lines = []
                in_sprint_ids = False
            elif key == "sprint_ids" and value == "[]":
                current[key] = []
                in_sprint_ids = False
            else:
                current[key] = yaml_unescape(value)
                in_sprint_ids = key == "sprint_ids"

    flush_block()
    if current is not None:
        chapters.append(current)
    return {"chapters": chapters}


def save(data: dict) -> None:
    path = ledger_path()
    path.parent.mkdir(parents=True, exist_ok=True)
    chapters = data.get("chapters", [])
    if not chapters:
        path.write_text("chapters: []\n")
        return

    lines = ["chapters:"]
    for chapter in chapters:
        first = True
        for key in FIELDS:
            if key not in chapter:
                continue
            prefix = "  - " if first else "    "
            if key == "sprint_ids":
                lines.append(f"{prefix}sprint_ids:")
                for sprint_id in chapter.get("sprint_ids", []):
                    lines.append(f"      - {yaml_escape(str(sprint_id))}")
            elif key == "summary":
                lines.append(f"{prefix}summary: >-")
                lines.append(f"      {chapter[key]}")
            else:
                lines.append(f"{prefix}{key}: {yaml_escape(str(chapter[key]))}")
            first = False
    path.write_text("\n".join(lines) + "\n")


def next_id(data: dict) -> str:
    max_n = 0
    for chapter in data.get("chapters", []):
        match = CHAPTER_RE.match(chapter.get("id", ""))
        if match:
            max_n = max(max_n, int(match.group(1)))
    return f"CHAPTER-{max_n + 1:04d}"


def find(data: dict, chapter_id: str) -> dict | None:
    for chapter in data.get("chapters", []):
        if chapter.get("id") == chapter_id:
            return chapter
    return None


def cmd_next_id(_args) -> int:
    print(next_id(load()))
    return 0


def cmd_add(args) -> int:
    data = load()
    chapter_id = args.id or next_id(data)
    if not CHAPTER_RE.match(chapter_id):
        sys.stderr.write("chapter id must look like CHAPTER-0001\n")
        return 1
    if find(data, chapter_id):
        sys.stderr.write(f"{chapter_id} already exists\n")
        return 1

    timestamp = now()
    data.setdefault("chapters", []).append({
        "id": chapter_id,
        "title": args.title,
        "status": args.status,
        "doc": args.doc,
        "created": timestamp,
        "updated": timestamp,
        "summary": args.summary,
        "sprint_ids": args.sprint_id or [],
    })
    save(data)
    print(chapter_id)
    return 0


def cmd_list(args) -> int:
    data = load()
    rows = data.get("chapters", [])
    if args.status:
        rows = [chapter for chapter in rows if chapter.get("status") == args.status]
    for chapter in rows:
        print(f"{chapter.get('id')}  {chapter.get('status')}  {chapter.get('title')}")
    return 0


def cmd_get(args) -> int:
    chapter = find(load(), args.id)
    if not chapter:
        sys.stderr.write(f"{args.id} not found\n")
        return 1
    for key in FIELDS:
        if key in chapter:
            value = chapter[key]
            if isinstance(value, list):
                print(f"{key}: {', '.join(value)}")
            else:
                print(f"{key}: {value}")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description="Manage chapter ledger")
    sub = parser.add_subparsers(dest="cmd", required=True)

    sub.add_parser("next-id", help="Print the next chapter id").set_defaults(func=cmd_next_id)

    add = sub.add_parser("add", help="Add a chapter ledger entry")
    add.add_argument("title")
    add.add_argument("--id", help="Override generated id")
    add.add_argument("--status", choices=STATUSES, default="active")
    add.add_argument("--doc", required=True, help="Chapter doc path")
    add.add_argument("--summary", required=True, help="Concise chapter summary")
    add.add_argument("--sprint-id", action="append", help="Existing sprint id to link")
    add.set_defaults(func=cmd_add)

    list_parser = sub.add_parser("list", help="List chapters")
    list_parser.add_argument("--status", choices=STATUSES)
    list_parser.set_defaults(func=cmd_list)

    get = sub.add_parser("get", help="Show one chapter")
    get.add_argument("id")
    get.set_defaults(func=cmd_get)

    args = parser.parse_args()
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
