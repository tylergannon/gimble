#!/usr/bin/env python3
"""Validate Diffusion skill naming and mirror conventions."""

from __future__ import annotations

import difflib
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SKILL_ROOTS = [ROOT / ".claude" / "skills", ROOT / ".agents" / "skills"]
NAME_RE = re.compile(r"^name:\s*([a-z0-9-]+)\s*$", re.MULTILINE)
VALID_NAME_RE = re.compile(r"^df-[a-z0-9]+(?:-[a-z0-9]+)*$")
SKILL_PATH_RE = re.compile(r"(?<![\w/])(\.(?:claude|agents)/skills/[A-Za-z0-9._/-]+)")
FIND_INVOCATION_RE = re.compile(r"\bfind\s+(?P<body>[^\n|;&)>]+)")
KNOWN_HELPER_ROOTS = {".claude", "~/.claude", ".agents", "~/.agents", "~/.codex"}


def skill_name(skill_md: Path) -> str | None:
    match = NAME_RE.search(skill_md.read_text())
    return match.group(1) if match else None


def referenced_skill_paths(content: str) -> list[str]:
    paths: list[str] = []
    for match in SKILL_PATH_RE.finditer(content):
        paths.append(match.group(1).rstrip(".,):;"))
    return paths


def bundled_resources(skill_dir: Path) -> dict[str, Path]:
    return {
        path.relative_to(skill_dir).as_posix(): path
        for path in skill_dir.rglob("*")
        if path.is_file()
        and path.name != "SKILL.md"
        and "__pycache__" not in path.parts
        and path.suffix not in {".pyc", ".pyo"}
    }


def normalize_mirror_content(content: str, mirror: str) -> str:
    root_maps = {
        ".claude": {".claude": ".SKILL_SCOPE", "~/.claude": "~/.SKILL_HOME"},
        ".agents": {
            ".agents": ".SKILL_SCOPE",
            "~/.agents": "~/.SKILL_HOME",
            "~/.codex": "~/.SKILL_HOME",
        },
    }
    if mirror not in root_maps:
        raise ValueError(f"unknown mirror: {mirror}")

    def normalize_find(match: re.Match[str]) -> str:
        normalized: list[str] = []
        mapped_roots: set[str] = set()
        tokens = match.group("body").split()
        if not any(token in KNOWN_HELPER_ROOTS for token in tokens):
            return match.group(0)
        for token in tokens:
            mapped = root_maps[mirror].get(token, token)
            if mapped in {".SKILL_SCOPE", "~/.SKILL_HOME"}:
                if mapped in mapped_roots:
                    continue
                mapped_roots.add(mapped)
            normalized.append(mapped)
        return f"find {' '.join(normalized)} "

    content = FIND_INVOCATION_RE.sub(normalize_find, content)

    content = (
        content
        .replace(".claude/skills", ".SKILL_ROOT/skills")
        .replace(".agents/skills", ".SKILL_ROOT/skills")
        .replace(".codex/skills", ".SKILL_ROOT/skills")
    )
    # Agent runtimes have two equivalent user-scope roots in the helper docs.
    # Collapse repeated canonical paths regardless of comma/and/or prose style.
    content = re.sub(
        r"(`~/.SKILL_ROOT/skills/[^`]+`)"
        r"(?:\s*(?:,\s*)?(?:(?:or|and)\s+)?\1)+",
        r"\1",
        content,
    )
    if mirror == ".claude":
        replacements = [
            ("CLAUDE.md", "INSTRUCTIONS.md"),
            ("AGENTS.md", "SECONDARY_INSTRUCTIONS.md"),
        ]
    elif mirror == ".agents":
        replacements = [
            ("AGENTS.md", "INSTRUCTIONS.md"),
            ("CLAUDE.md", "SECONDARY_INSTRUCTIONS.md"),
        ]
    for source, target in replacements:
        content = content.replace(source, target)
    return content


def diff_excerpt(left: str, right: str, left_name: str, right_name: str) -> str:
    lines = list(
        difflib.unified_diff(
            left.splitlines(),
            right.splitlines(),
            fromfile=left_name,
            tofile=right_name,
            lineterm="",
            n=2,
        )
    )
    return "\n".join(lines[:18])


def helper_find_root_sets(content: str) -> list[set[str]]:
    root_sets: list[set[str]] = []
    for match in FIND_INVOCATION_RE.finditer(content):
        roots: set[str] = set()
        for token in match.group("body").split():
            if token.startswith("-"):
                continue
            if token in KNOWN_HELPER_ROOTS:
                roots.add(token)
        if roots:
            root_sets.append(roots)
    return root_sets


def wrong_mirror_helper(content: str, mirror: str) -> bool:
    wrong = {
        ".claude": {".agents", "~/.agents", "~/.codex"},
        ".agents": {".claude", "~/.claude"},
    }[mirror]
    for roots in helper_find_root_sets(content):
        if roots & wrong:
            return True
    return False


def missing_required_helper_roots(content: str, mirror: str) -> bool:
    required = {
        ".claude": {".claude", "~/.claude"},
        ".agents": {".agents", "~/.agents", "~/.codex"},
    }[mirror]
    for roots in helper_find_root_sets(content):
        if not required.issubset(roots):
            return True
    return False


def main() -> int:
    errors: list[str] = []
    skill_sets: dict[str, set[str]] = {}
    skill_contents: dict[tuple[str, str], str] = {}

    for root in SKILL_ROOTS:
        if not root.exists():
            continue
        skill_sets[str(root)] = set()
        for skill_dir in sorted(path for path in root.iterdir() if path.is_dir()):
            skill_sets[str(root)].add(skill_dir.name)
            skill_md = skill_dir / "SKILL.md"
            if not skill_md.exists():
                errors.append(f"{skill_dir}: missing SKILL.md")
                continue

            content = skill_md.read_text()
            name = skill_name(skill_md)
            if not name:
                errors.append(f"{skill_md}: missing frontmatter name")
                continue
            skill_contents[(root.parent.name, skill_dir.name)] = content

            if name != skill_dir.name:
                errors.append(f"{skill_md}: name {name!r} does not match directory {skill_dir.name!r}")
            if not VALID_NAME_RE.match(name):
                errors.append(f"{skill_md}: name {name!r} must use df-* kebab-case without colons")
            if ".Codex/skills" in content:
                errors.append(f"{skill_md}: use .agents/skills or .claude/skills, not stale .Codex/skills")
            if ".claude/skills" in content and ".agents/skills" in str(skill_md):
                errors.append(f"{skill_md}: .agents skill references .claude/skills path")
            if ".agents/skills" in content and ".claude/skills" in str(skill_md):
                errors.append(f"{skill_md}: .claude skill references .agents/skills path")
            if ".codex/skills" in content and ".claude/skills" in str(skill_md):
                errors.append(f"{skill_md}: .claude skill references .codex/skills path")
            mirror = root.parent.name
            if wrong_mirror_helper(content, mirror):
                errors.append(f"{skill_md}: helper search starts from the opposite mirror root")
            if missing_required_helper_roots(content, mirror):
                errors.append(f"{skill_md}: helper search omits a required mirror root")
            if "`AGENTS.md`, `AGENTS.md`" in content:
                errors.append(f"{skill_md}: duplicate AGENTS.md instruction-file reference")
            if "Codex/Codex/Gemini" in content or "Codex + Codex" in content:
                errors.append(f"{skill_md}: duplicate Codex agent wording; expected distinct Claude/Codex/Gemini roles")
            if "SPRINT-NNN-Codex-DRAFT" in content or "SPRINT-NNN-Codex-CRITIQUE" in content:
                errors.append(f"{skill_md}: mixed-case Codex draft names collide with CODEX on case-insensitive filesystems")
            for relative_path in referenced_skill_paths(content):
                if not (ROOT / relative_path).exists():
                    errors.append(f"{skill_md}: referenced skill path does not exist: {relative_path}")

    if len(skill_sets) == len(SKILL_ROOTS):
        roots = list(skill_sets)
        if skill_sets[roots[0]] != skill_sets[roots[1]]:
            left_only = sorted(skill_sets[roots[0]] - skill_sets[roots[1]])
            right_only = sorted(skill_sets[roots[1]] - skill_sets[roots[0]])
            errors.append(f"skill sets differ between mirrors: {roots[0]} only={left_only}; {roots[1]} only={right_only}")
        for skill in sorted(skill_sets[roots[0]] & skill_sets[roots[1]]):
            claude_content = skill_contents.get((".claude", skill))
            agents_content = skill_contents.get((".agents", skill))
            if not claude_content or not agents_content:
                continue
            normalized_claude = normalize_mirror_content(claude_content, ".claude")
            normalized_agents = normalize_mirror_content(agents_content, ".agents")
            if normalized_claude != normalized_agents:
                excerpt = diff_excerpt(
                    normalized_claude,
                    normalized_agents,
                    f".claude/skills/{skill}/SKILL.md",
                    f".agents/skills/{skill}/SKILL.md",
                )
                errors.append(
                    f"{skill}: .claude and .agents skill bodies differ beyond expected "
                    f"path/instruction-file swaps\n{excerpt}"
                )

            claude_resources = bundled_resources(SKILL_ROOTS[0] / skill)
            agents_resources = bundled_resources(SKILL_ROOTS[1] / skill)
            if set(claude_resources) != set(agents_resources):
                claude_only = sorted(set(claude_resources) - set(agents_resources))
                agents_only = sorted(set(agents_resources) - set(claude_resources))
                errors.append(
                    f"{skill}: bundled resource sets differ between mirrors: "
                    f".claude only={claude_only}; .agents only={agents_only}"
                )
            for relative in sorted(set(claude_resources) & set(agents_resources)):
                claude_path = claude_resources[relative]
                agents_path = agents_resources[relative]
                claude_bytes = claude_path.read_bytes()
                agents_bytes = agents_path.read_bytes()
                try:
                    claude_text = claude_bytes.decode("utf-8")
                    agents_text = agents_bytes.decode("utf-8")
                except UnicodeDecodeError:
                    if claude_bytes != agents_bytes:
                        errors.append(f"{skill}: binary resource differs between mirrors: {relative}")
                    continue

                if (
                    ".agents/skills" in claude_text
                    or "~/.agents/skills" in claude_text
                    or ".codex/skills" in claude_text
                    or "~/.codex/skills" in claude_text
                ):
                    errors.append(
                        f"{skill}: .claude resource contains an agent-only skill path: {relative}"
                    )
                if ".claude/skills" in agents_text or "~/.claude/skills" in agents_text:
                    errors.append(
                        f"{skill}: .agents resource contains a .claude skill path: {relative}"
                    )
                if wrong_mirror_helper(claude_text, ".claude"):
                    errors.append(
                        f"{skill}: .claude resource helper search starts from .agents: {relative}"
                    )
                if wrong_mirror_helper(agents_text, ".agents"):
                    errors.append(
                        f"{skill}: .agents resource helper search starts from .claude: {relative}"
                    )
                if missing_required_helper_roots(claude_text, ".claude"):
                    errors.append(
                        f"{skill}: .claude resource helper search omits a required root: {relative}"
                    )
                if missing_required_helper_roots(agents_text, ".agents"):
                    errors.append(
                        f"{skill}: .agents resource helper search omits a required root: {relative}"
                    )

                normalized_claude_resource = normalize_mirror_content(
                    claude_text, ".claude"
                )
                normalized_agents_resource = normalize_mirror_content(
                    agents_text, ".agents"
                )
                if normalized_claude_resource != normalized_agents_resource:
                    excerpt = diff_excerpt(
                        normalized_claude_resource,
                        normalized_agents_resource,
                        f".claude/skills/{skill}/{relative}",
                        f".agents/skills/{skill}/{relative}",
                    )
                    errors.append(
                        f"{skill}: bundled resource differs between mirrors: {relative}\n{excerpt}"
                    )

    if errors:
        for error in errors:
            print(error, file=sys.stderr)
        return 1

    print("Skill naming validation passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
