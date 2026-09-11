#!/usr/bin/env python3
"""Manage durable repository Promise Loop state and README badges."""

from __future__ import annotations

import argparse
import contextlib
import datetime as dt
import fcntl
import functools
import hashlib
import json
import os
import re
import subprocess
import sys
import urllib.parse
import uuid
from pathlib import Path, PurePosixPath
from typing import Any


SCHEMA_VERSION = 2
SUPPORTED_SCHEMA_VERSIONS = {1, SCHEMA_VERSION}
PROMISE_ID_RE = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
COST_BANDS = ("lt-10", "10-100", "100-1k", "1k-10k", "10k-100k", "gte-100k", "unknown")
APPROVAL_COST_BANDS = {"100-1k", "1k-10k", "10k-100k", "gte-100k", "unknown"}
STAGED_COST_BANDS = {"10k-100k", "gte-100k"}
GATE_EVIDENCE_KINDS = {"coverage", "inspection", "test", "visual", "tool", "verification"}
COST_BAND_BOUNDS: dict[str, tuple[float, float | None]] = {
    "lt-10": (0.0, 10.0),
    "10-100": (10.0, 100.0),
    "100-1k": (100.0, 1_000.0),
    "1k-10k": (1_000.0, 10_000.0),
    "10k-100k": (10_000.0, 100_000.0),
    "gte-100k": (100_000.0, None),
}
BADGE_START = "<!-- df-promise-badges:start -->"
BADGE_END = "<!-- df-promise-badges:end -->"
BADGE_RE = re.compile(
    rf"(?:\r?\n)?{re.escape(BADGE_START)}.*?{re.escape(BADGE_END)}(?:\r?\n)?",
    re.DOTALL,
)
BADGE_ITEM_RE = re.compile(r"\[!\[[^\]]+\]\([^)]+\)\]\(([^)]+)\)")


class PromiseError(RuntimeError):
    """Expected user-facing error."""


@contextlib.contextmanager
def mutation_lock(root: Path):
    lock_dir = root / ".promises" / ".work"
    lock_dir.mkdir(parents=True, exist_ok=True)
    lock_path = lock_dir / "state.lock"
    with lock_path.open("a+") as stream:
        try:
            fcntl.flock(stream.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError as exc:
            raise PromiseError("another Promise state mutation is active; retry this bounded command") from exc
        try:
            yield
        finally:
            fcntl.flock(stream.fileno(), fcntl.LOCK_UN)


def serialized(command):
    @functools.wraps(command)
    def wrapper(args: argparse.Namespace) -> None:
        root = repo_root()
        if not (root / ".promises").exists():
            command(args)
            return
        with mutation_lock(root):
            command(args)

    return wrapper


def utc_now() -> dt.datetime:
    return dt.datetime.now(dt.timezone.utc)


def iso_time(value: dt.datetime | None = None) -> str:
    value = value or utc_now()
    return value.replace(microsecond=0).isoformat().replace("+00:00", "Z")


def parse_time(value: str) -> dt.datetime:
    return dt.datetime.fromisoformat(value.replace("Z", "+00:00"))


def run_git(root: Path, *args: str, check: bool = True) -> str:
    completed = subprocess.run(
        ["git", *args],
        cwd=root,
        check=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    if check and completed.returncode != 0:
        detail = completed.stderr.strip() or completed.stdout.strip()
        raise PromiseError(f"git {' '.join(args)} failed: {detail}")
    return completed.stdout.strip()


def repo_root() -> Path:
    completed = subprocess.run(
        ["git", "rev-parse", "--show-toplevel"],
        check=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    if completed.returncode != 0:
        raise PromiseError("df-promise must run inside a Git repository")
    return Path(completed.stdout.strip()).resolve()


def read_json(path: Path) -> dict[str, Any]:
    try:
        data = json.loads(path.read_text())
    except FileNotFoundError as exc:
        raise PromiseError(f"missing required state file: {path}") from exc
    except json.JSONDecodeError as exc:
        raise PromiseError(f"invalid JSON in {path}: {exc}") from exc
    if not isinstance(data, dict):
        raise PromiseError(f"expected a JSON object in {path}")
    return data


def write_json(path: Path, data: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_name(f".{path.name}.{os.getpid()}.tmp")
    temporary.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n")
    os.replace(temporary, path)


def append_jsonl(path: Path, data: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as stream:
        stream.write(json.dumps(data, sort_keys=True) + "\n")


def validate_id(promise_id: str) -> str:
    if not PROMISE_ID_RE.fullmatch(promise_id):
        raise PromiseError("promise id must use lowercase ASCII kebab-case")
    return promise_id


def promise_root(root: Path, promise_id: str) -> Path:
    return root / ".promises" / validate_id(promise_id)


def load_registry(root: Path) -> dict[str, Any]:
    registry_path = root / ".promises" / "registry.json"
    if not registry_path.exists():
        return {"schema_version": SCHEMA_VERSION, "promises": []}
    registry = read_json(registry_path)
    if registry.get("schema_version") not in SUPPORTED_SCHEMA_VERSIONS:
        raise PromiseError("unsupported Promise registry schema; upgrade df-promise before continuing")
    entries = registry.get("promises")
    if not isinstance(entries, list):
        raise PromiseError("registry promises field must be a list")
    seen: set[str] = set()
    for index, entry in enumerate(entries):
        if not isinstance(entry, dict):
            raise PromiseError(f"registry promise entry {index} must be an object")
        promise_id = str(entry.get("id", ""))
        if not PROMISE_ID_RE.fullmatch(promise_id):
            raise PromiseError(f"registry promise entry {index} has an invalid id")
        if promise_id in seen:
            raise PromiseError(f"registry contains duplicate promise id: {promise_id}")
        seen.add(promise_id)
        if not isinstance(entry.get("name"), str) or not entry["name"].strip():
            raise PromiseError(f"registry entry for {promise_id} needs a name")
        expected_contract = f".promises/{promise_id}/contract.json"
        if entry.get("contract") != expected_contract:
            raise PromiseError(
                f"registry entry for {promise_id} must use contract path {expected_contract}"
            )
        readme_value = entry.get("readme")
        if not isinstance(readme_value, str) or not readme_value.strip():
            raise PromiseError(f"registry entry for {promise_id} needs a README path")
        readme_path = PurePosixPath(readme_value)
        if not readme_path.parts or readme_path == PurePosixPath(".") or readme_path.is_absolute() or ".." in readme_path.parts:
            raise PromiseError(f"registry entry for {promise_id} has an unsafe README path")
    return registry


def registered_entry(root: Path, promise_id: str) -> dict[str, Any]:
    registry = load_registry(root)
    for entry in registry.get("promises", []):
        if str(entry.get("id")) == promise_id:
            return entry
    raise PromiseError(f"promise is not registered: {promise_id}; restore it before running")


def normalize_repo_path(root: Path, value: str) -> Path:
    candidate = (root / value).resolve() if not Path(value).is_absolute() else Path(value).resolve()
    try:
        candidate.relative_to(root)
    except ValueError as exc:
        raise PromiseError(f"path must remain inside the repository: {value}") from exc
    return candidate


def parse_key_value(value: str, label: str) -> tuple[str, str]:
    if "=" not in value:
        raise PromiseError(f"{label} must use id=description syntax: {value}")
    key, description = value.split("=", 1)
    key = key.strip()
    description = description.strip()
    if not PROMISE_ID_RE.fullmatch(key) or not description:
        raise PromiseError(f"invalid {label}: {value}")
    return key, description


def parse_tool(value: str) -> dict[str, str]:
    if "=" not in value:
        raise PromiseError(f"tool must use tools/path=purpose syntax: {value}")
    path_value, purpose = value.split("=", 1)
    path_value = path_value.strip()
    purpose = purpose.strip()
    if not path_value or not purpose:
        raise PromiseError(f"invalid tool declaration: {value}")
    path = PurePosixPath(path_value)
    if path.is_absolute() or ".." in path.parts:
        raise PromiseError(f"tool path must be relative and confined to tools/: {path_value}")
    if not path.parts or path.parts[0] != "tools":
        path = PurePosixPath("tools") / path
    return {"path": path.as_posix(), "purpose": purpose}


def evidence_reference_valid(root: Path, promise_id: str, mode: str, value: Any) -> bool:
    if not isinstance(value, str) or not value.strip():
        return False
    reference = value.strip()
    if mode in {"internal", "mixed"}:
        try:
            evidence_path = normalize_repo_path(root, reference)
            evidence_path.relative_to(promise_root(root, promise_id))
            if evidence_path.is_file():
                return True
        except (PromiseError, ValueError):
            pass
        if mode == "internal":
            return False
    if re.fullmatch(r"sha256:[0-9a-fA-F]{64}", reference):
        return True
    parsed = urllib.parse.urlparse(reference)
    return parsed.scheme in {"https", "s3", "gs", "artifact", "provider", "request"} and bool(
        parsed.netloc or parsed.path
    ) and len(reference) >= 12


def load_contract(root: Path, promise_id: str, *, require_tools: bool = True) -> dict[str, Any]:
    contract = read_json(promise_root(root, promise_id) / "contract.json")
    validate_contract(root, promise_id, contract, require_tools=require_tools)
    return contract


def load_state(root: Path, promise_id: str) -> dict[str, Any]:
    state = read_json(promise_root(root, promise_id) / "state.json")
    if state.get("schema_version") != SCHEMA_VERSION or state.get("id") != promise_id:
        raise PromiseError(f"invalid state identity or schema for {promise_id}")
    if not isinstance(state.get("status"), str):
        raise PromiseError(f"state for {promise_id} needs a status")
    if state.get("active_run") is not None and not isinstance(state.get("active_run"), str):
        raise PromiseError(f"state for {promise_id} has an invalid active_run")
    if state.get("last_run") is not None and not isinstance(state.get("last_run"), str):
        raise PromiseError(f"state for {promise_id} has an invalid last_run")
    if state.get("last_fulfillment") is not None and not isinstance(state.get("last_fulfillment"), dict):
        raise PromiseError(f"state for {promise_id} has an invalid last_fulfillment")
    lifetime = state.get("lifetime_metrics")
    if not isinstance(lifetime, dict):
        raise PromiseError(f"state for {promise_id} needs lifetime_metrics")
    for field in ("runs", "steps", "tool_calls", "tokens_known", "cost_usd_known", "wall_seconds"):
        value = lifetime.get(field)
        if isinstance(value, bool) or not isinstance(value, (int, float)) or value < 0:
            raise PromiseError(f"state for {promise_id} has invalid lifetime metric {field}")
    if not isinstance(lifetime.get("tokens_complete"), bool):
        raise PromiseError(f"state for {promise_id} has invalid lifetime metric tokens_complete")
    if not isinstance(lifetime.get("cost_complete"), bool):
        raise PromiseError(f"state for {promise_id} has invalid lifetime metric cost_complete")
    return state


def validate_contract(
    root: Path,
    promise_id: str,
    contract: dict[str, Any],
    *,
    require_tools: bool = True,
) -> None:
    if contract.get("schema_version") != SCHEMA_VERSION:
        if contract.get("schema_version") == 1:
            raise PromiseError(f"contract schema 1 for {promise_id} requires: promise.py migrate {promise_id} ...")
        raise PromiseError(f"unsupported contract schema for {promise_id}")
    if contract.get("id") != promise_id:
        raise PromiseError(f"contract id does not match directory for {promise_id}")
    if (
        not isinstance(contract.get("name"), str)
        or not contract["name"].strip()
        or not isinstance(contract.get("statement"), str)
        or not contract["statement"].strip()
    ):
        raise PromiseError(f"contract for {promise_id} needs a name and statement")
    if not isinstance(contract.get("human_signoff_required"), bool):
        raise PromiseError(f"contract for {promise_id} needs boolean human_signoff_required")
    gates = contract.get("acceptance_gates")
    if not isinstance(gates, list) or not gates:
        raise PromiseError(f"contract for {promise_id} needs at least one acceptance gate")
    gate_ids: set[str] = set()
    for gate in gates:
        if not isinstance(gate, dict):
            raise PromiseError(f"invalid acceptance gate in {promise_id}")
        gate_id = gate.get("id")
        if (
            not isinstance(gate_id, str)
            or not PROMISE_ID_RE.fullmatch(gate_id)
            or not isinstance(gate.get("description"), str)
            or not gate["description"].strip()
        ):
            raise PromiseError(f"invalid acceptance gate in {promise_id}: {gate}")
        if gate_id in gate_ids:
            raise PromiseError(f"duplicate acceptance gate in {promise_id}: {gate_id}")
        gate_ids.add(gate_id)
    if contract["human_signoff_required"] and "human-signoff" not in gate_ids:
        raise PromiseError(f"contract for {promise_id} requires a human-signoff acceptance gate")
    scope = contract.get("scope")
    if not isinstance(scope, dict):
        raise PromiseError(f"contract for {promise_id} needs a scope object")
    if not isinstance(scope.get("include"), list) or not scope.get("include"):
        raise PromiseError(f"contract for {promise_id} needs non-empty scope.include")
    if not isinstance(scope.get("exclude", []), list):
        raise PromiseError(f"contract for {promise_id} has invalid scope.exclude")
    for field in ("include", "exclude"):
        for pattern in scope.get(field, []):
            if not isinstance(pattern, str) or not pattern.strip():
                raise PromiseError(f"contract for {promise_id} has an invalid scope.{field} pattern")
            parsed_pattern = PurePosixPath(pattern)
            if parsed_pattern.is_absolute() or ".." in parsed_pattern.parts:
                raise PromiseError(f"contract for {promise_id} has an unsafe scope.{field} pattern: {pattern}")
    operation = contract.get("operation")
    if not isinstance(operation, dict):
        raise PromiseError(f"contract for {promise_id} needs an operation object")
    if (
        not isinstance(operation.get("cadence"), str)
        or not operation["cadence"].strip()
        or not isinstance(operation.get("runner"), str)
        or not operation["runner"].strip()
        or not isinstance(operation.get("authority"), str)
        or not operation["authority"].strip()
    ):
        raise PromiseError(f"contract for {promise_id} needs operation cadence, runner, and authority")
    plan_steps = operation.get("plan_steps")
    if not isinstance(plan_steps, list) or not plan_steps:
        raise PromiseError(f"contract for {promise_id} needs at least one operating-plan step")
    if any(not isinstance(step, str) or not step.strip() for step in plan_steps):
        raise PromiseError(f"contract for {promise_id} has an invalid operating-plan step")
    budgets = operation.get("budgets")
    budget_fields = {"max_steps", "max_tool_calls", "max_tokens", "max_wall_seconds"}
    if not isinstance(budgets, dict) or not budget_fields.issubset(budgets):
        raise PromiseError(f"contract for {promise_id} needs all operation budget fields")
    for field in sorted(budget_fields):
        value = budgets[field]
        if value is None:
            continue
        if isinstance(value, bool) or not isinstance(value, (int, float)) or value < 0:
            raise PromiseError(f"contract for {promise_id} has invalid operation budget {field}")
        if field != "max_wall_seconds" and not isinstance(value, int):
            raise PromiseError(f"contract for {promise_id} budget {field} must be an integer")
    invalidation = contract.get("invalidation")
    if not isinstance(invalidation, dict) or invalidation.get("scope_change") is not True:
        raise PromiseError(f"contract for {promise_id} must invalidate on scope changes")
    max_age = invalidation.get("max_age_hours")
    if max_age is not None and (
        isinstance(max_age, bool) or not isinstance(max_age, (int, float)) or max_age < 0
    ):
        raise PromiseError(f"contract for {promise_id} has invalid invalidation.max_age_hours")
    evidence = contract.get("evidence")
    if not isinstance(evidence, dict) or evidence.get("mode") not in {"internal", "external", "mixed"}:
        raise PromiseError(f"contract for {promise_id} needs evidence.mode internal, external, or mixed")
    if evidence["mode"] != "internal" and max_age is None:
        raise PromiseError(f"contract for {promise_id} uses external evidence and needs max_age_hours")
    verifier = contract.get("verifier")
    if (
        not isinstance(verifier, dict)
        or not isinstance(verifier.get("identity"), str)
        or not verifier["identity"].strip()
    ):
        raise PromiseError(f"contract for {promise_id} needs verifier.identity")
    economics = contract.get("economics")
    if not isinstance(economics, dict):
        raise PromiseError(f"contract for {promise_id} needs economics")
    cost_band = economics.get("full_run_cost_band")
    if cost_band not in COST_BANDS:
        raise PromiseError(f"contract for {promise_id} has invalid economics.full_run_cost_band")
    if not isinstance(economics.get("estimate_basis"), str) or not economics["estimate_basis"].strip():
        raise PromiseError(f"contract for {promise_id} needs economics.estimate_basis")
    if not isinstance(economics.get("approval_required"), bool):
        raise PromiseError(f"contract for {promise_id} needs boolean economics.approval_required")
    max_cost = economics.get("max_run_cost_usd")
    if max_cost is not None and (
        isinstance(max_cost, bool) or not isinstance(max_cost, (int, float)) or max_cost < 0
    ):
        raise PromiseError(f"contract for {promise_id} has invalid economics.max_run_cost_usd")
    if cost_band in APPROVAL_COST_BANDS and not economics["approval_required"]:
        raise PromiseError(f"contract for {promise_id} cost band {cost_band} requires per-run approval")
    if economics["approval_required"] and (max_cost is None or max_cost <= 0):
        raise PromiseError(f"contract for {promise_id} approval requires a positive max_run_cost_usd")
    if cost_band in COST_BAND_BOUNDS and max_cost is not None:
        lower, upper = COST_BAND_BOUNDS[cost_band]
        if max_cost < lower or (cost_band == "lt-10" and max_cost >= upper) or (
            cost_band != "lt-10" and upper is not None and max_cost > upper
        ):
            raise PromiseError(f"contract for {promise_id} max_run_cost_usd conflicts with cost band {cost_band}")
    staging_plan = economics.get("staging_plan")
    if staging_plan is not None and (not isinstance(staging_plan, str) or not staging_plan.strip()):
        raise PromiseError(f"contract for {promise_id} has invalid economics.staging_plan")
    if cost_band in STAGED_COST_BANDS and not staging_plan:
        raise PromiseError(f"contract for {promise_id} cost band {cost_band} requires a staging plan")
    tools = contract.get("tools")
    if not isinstance(tools, list):
        raise PromiseError(f"contract for {promise_id} needs a tools list")
    for tool in tools:
        if (
            not isinstance(tool, dict)
            or not isinstance(tool.get("purpose"), str)
            or not tool["purpose"].strip()
        ):
            raise PromiseError(f"invalid promise tool declaration in {promise_id}")
        relative = PurePosixPath(str(tool.get("path", "")))
        if relative.is_absolute() or ".." in relative.parts or not relative.parts or relative.parts[0] != "tools":
            raise PromiseError(f"tool escapes promise-local tools/ in {promise_id}: {relative}")
        tool_path = promise_root(root, promise_id) / relative
        if require_tools and not tool_path.is_file():
            raise PromiseError(f"declared promise tool is not a file: {tool_path}")
    badge = contract.get("badge")
    if not isinstance(badge, dict):
        raise PromiseError(f"contract for {promise_id} needs a badge object")
    if (
        not isinstance(badge.get("label"), str)
        or not badge["label"].strip()
        or not isinstance(badge.get("readme"), str)
        or not badge["readme"].strip()
    ):
        raise PromiseError(f"contract for {promise_id} needs badge label and README path")
    readme = normalize_repo_path(root, str(badge["readme"]))
    if readme.is_dir():
        raise PromiseError(f"badge README path is a directory: {readme}")


@functools.lru_cache(maxsize=512)
def glob_regex(pattern: str) -> re.Pattern[str]:
    pieces: list[str] = []
    index = 0
    while index < len(pattern):
        character = pattern[index]
        if character == "*":
            if index + 1 < len(pattern) and pattern[index + 1] == "*":
                index += 2
                while index < len(pattern) and pattern[index] == "*":
                    index += 1
                if index < len(pattern) and pattern[index] == "/":
                    pieces.append("(?:.*/)?")
                    index += 1
                else:
                    pieces.append(".*")
                continue
            pieces.append("[^/]*")
        elif character == "?":
            pieces.append("[^/]")
        else:
            pieces.append(re.escape(character))
        index += 1
    return re.compile("".join(pieces))


def matches_pattern(path: str, pattern: str) -> bool:
    return glob_regex(pattern).fullmatch(path) is not None


def in_scope(path: str, contract: dict[str, Any]) -> bool:
    if path == ".promises" or path.startswith(".promises/") or path == ".git" or path.startswith(".git/"):
        return False
    scope = contract["scope"]
    if not any(matches_pattern(path, pattern) for pattern in scope["include"]):
        return False
    return not any(matches_pattern(path, pattern) for pattern in scope.get("exclude", []))


def managed_readme_paths(root: Path, contract: dict[str, Any]) -> set[str]:
    readmes = {normalize_repo_path(root, contract["badge"]["readme"]).relative_to(root).as_posix()}
    registry = load_registry(root)
    for entry in registry.get("promises", []):
        readme_value = entry.get("readme")
        if not readme_value and entry.get("contract"):
            try:
                other_contract = read_json(normalize_repo_path(root, str(entry["contract"])))
                readme_value = other_contract.get("badge", {}).get("readme")
            except PromiseError:
                continue
        if readme_value:
            readmes.add(normalize_repo_path(root, str(readme_value)).relative_to(root).as_posix())
    return readmes


def strip_badge_block(text: str) -> str:
    return BADGE_RE.sub("", text)


def insert_badge_block(text: str, badges: list[str]) -> str:
    without_badges = strip_badge_block(text)
    if not badges:
        return without_badges
    block = f"{BADGE_START}\n{' '.join(sorted(badges))}\n{BADGE_END}"
    if without_badges.startswith("# "):
        title_end = without_badges.find("\n")
        if title_end < 0:
            return f"{without_badges}\n{block}"
        insertion = title_end + 1
        return f"{without_badges[:insertion]}\n{block}\n{without_badges[insertion:]}"
    return f"{block}\n{without_badges}" if without_badges else block


def managed_badge_items(text: str) -> list[re.Match[str]]:
    managed = BADGE_RE.search(text)
    return [] if managed is None else list(BADGE_ITEM_RE.finditer(managed.group(0)))


def remove_promise_badge(root: Path, promise_id: str, readme_value: str) -> list[str]:
    readme = normalize_repo_path(root, readme_value)
    if not readme.exists():
        return []
    existing = readme.read_text()
    items = managed_badge_items(existing)
    if not items:
        return []
    target = f".promises/{promise_id}/PROMISE.md"
    retained = [
        match.group(0)
        for match in items
        if not match.group(1).endswith(target)
    ]
    revised = insert_badge_block(existing, retained)
    if revised == existing:
        return []
    readme.write_text(revised)
    return [readme.relative_to(root).as_posix()]


def normalize_badge_block(content: bytes) -> bytes:
    text = content.decode("utf-8", errors="surrogateescape")
    return strip_badge_block(text).encode("utf-8", errors="surrogateescape")


def repository_files(root: Path) -> list[str]:
    completed = subprocess.run(
        ["git", "ls-files", "-z", "--cached", "--others", "--exclude-standard"],
        cwd=root,
        check=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    if completed.returncode != 0:
        detail = completed.stderr.decode("utf-8", errors="replace").strip()
        raise PromiseError(f"git ls-files failed: {detail}")
    return sorted(item.decode("utf-8", errors="surrogateescape") for item in completed.stdout.split(b"\0") if item)


_FILE_DIGEST_CACHE: dict[tuple[Any, ...], bytes] = {}


def repository_file_digest(path: Path, *, normalize_badges: bool) -> bytes:
    metadata = path.lstat()
    if path.is_symlink():
        link_target = os.readlink(path)
        cache_key = (str(path), "symlink", link_target, normalize_badges)
        content = link_target.encode("utf-8", errors="surrogateescape")
    else:
        cache_key = (
            str(path),
            metadata.st_size,
            metadata.st_mtime_ns,
            metadata.st_ctime_ns,
            normalize_badges,
        )
        cached = _FILE_DIGEST_CACHE.get(cache_key)
        if cached is not None:
            return cached
        content = path.read_bytes()
    if normalize_badges:
        content = normalize_badge_block(content)
    digest = hashlib.sha256(content).digest()
    _FILE_DIGEST_CACHE[cache_key] = digest
    return digest


def subject_fingerprint(root: Path, contract: dict[str, Any]) -> tuple[str, int]:
    digest = hashlib.sha256()
    count = 0
    readme_relatives = managed_readme_paths(root, contract)
    for relative in repository_files(root):
        if not in_scope(relative, contract):
            continue
        path = root / relative
        if not path.exists() and not path.is_symlink():
            continue
        if not path.is_symlink() and not path.is_file():
            continue
        digest.update(relative.encode("utf-8", errors="surrogateescape"))
        digest.update(b"\0")
        digest.update(repository_file_digest(path, normalize_badges=relative in readme_relatives))
        count += 1
    return digest.hexdigest(), count


def mechanism_fingerprint(root: Path, promise_id: str) -> str:
    base = promise_root(root, promise_id)
    paths = [base / "contract.json", base / "PROMISE.md"]
    contract = read_json(base / "contract.json")
    paths.extend(base / str(tool["path"]) for tool in contract.get("tools", []))
    digest = hashlib.sha256()
    for path in paths:
        if not path.exists() and not path.is_symlink():
            raise PromiseError(f"missing promise mechanism file: {path}")
        relative = path.relative_to(base).as_posix()
        content = os.readlink(path).encode() if path.is_symlink() else path.read_bytes()
        digest.update(relative.encode())
        digest.update(b"\0")
        digest.update(hashlib.sha256(content).digest())
    return digest.hexdigest()


def subject_dirty_paths(root: Path, contract: dict[str, Any]) -> list[str]:
    changed = set(filter(None, run_git(root, "diff", "--name-only", "HEAD").splitlines()))
    changed.update(filter(None, run_git(root, "ls-files", "--others", "--exclude-standard").splitlines()))
    readme_relatives = managed_readme_paths(root, contract)
    dirty: list[str] = []
    for relative in sorted(changed):
        if not in_scope(relative, contract):
            continue
        if relative in readme_relatives and (root / relative).exists():
            completed = subprocess.run(
                ["git", "show", f"HEAD:{relative}"],
                cwd=root,
                check=False,
                stdout=subprocess.PIPE,
                stderr=subprocess.DEVNULL,
            )
            head_content = completed.stdout
            if normalize_badge_block((root / relative).read_bytes()) == normalize_badge_block(head_content):
                continue
        dirty.append(relative)
    return dirty


def promise_validity(root: Path, promise_id: str, contract: dict[str, Any], state: dict[str, Any]) -> tuple[bool, str]:
    fulfillment = state.get("last_fulfillment")
    if not isinstance(fulfillment, dict):
        return False, "never fulfilled"
    required_fields = ["fulfilled_at", "validated_commit", "subject_fingerprint", "mechanism_fingerprint"]
    if any(not fulfillment.get(field) for field in required_fields):
        return False, "fulfillment state is incomplete"
    current_subject, subject_files = subject_fingerprint(root, contract)
    if subject_files == 0:
        return False, "promise scope resolves to zero files"
    if current_subject != fulfillment.get("subject_fingerprint"):
        return False, "promise scope changed"
    try:
        current_mechanism = mechanism_fingerprint(root, promise_id)
    except PromiseError:
        return False, "promise contract or tools changed"
    if current_mechanism != fulfillment.get("mechanism_fingerprint"):
        return False, "promise contract or tools changed"
    max_age = contract.get("invalidation", {}).get("max_age_hours")
    if max_age is not None:
        try:
            age = utc_now() - parse_time(str(fulfillment.get("fulfilled_at")))
        except (TypeError, ValueError):
            return False, "fulfillment state is incomplete"
        if age.total_seconds() > float(max_age) * 3600:
            return False, "promise fulfillment expired"
    return True, "fulfilled"


def badge_markdown(root: Path, promise_id: str, contract: dict[str, Any], state: dict[str, Any]) -> str:
    fulfillment = state["last_fulfillment"]
    date = str(fulfillment["fulfilled_at"])[:10]
    commit = str(fulfillment["validated_commit"])[:7]
    label = contract["badge"]["label"]
    dirty = bool(fulfillment.get("dirty_subject_at_start"))
    message = f"fulfilled{' dirty' if dirty else ''} {date} @ {commit}"
    query = urllib.parse.urlencode(
        {"label": label, "message": message, "color": "brightgreen"}
    )
    image = f"https://img.shields.io/static/v1?{query}"
    readme = normalize_repo_path(root, contract["badge"]["readme"])
    promise_doc = promise_root(root, promise_id) / "PROMISE.md"
    link = os.path.relpath(promise_doc, readme.parent)
    return f"[![{label}]({image})]({link})"


def managed_badge_files(root: Path) -> set[Path]:
    completed = subprocess.run(
        ["git", "grep", "-lz", "--untracked", "-F", BADGE_START, "--"],
        cwd=root,
        check=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    if completed.returncode not in {0, 1}:
        detail = completed.stderr.decode("utf-8", errors="replace").strip()
        raise PromiseError(f"git grep failed while locating managed badges: {detail}")
    found: set[Path] = set()
    for raw_path in completed.stdout.split(b"\0"):
        if not raw_path:
            continue
        relative = raw_path.decode("utf-8", errors="surrogateescape")
        path = normalize_repo_path(root, relative)
        if path.is_symlink() or not path.is_file():
            continue
        content = path.read_text(errors="surrogateescape")
        lines = {line.strip() for line in content.splitlines()}
        if BADGE_START in lines and BADGE_END in lines:
            found.add(path)
    return found


def render_badges(root: Path, extra_readmes: list[str] | None = None) -> dict[str, Any]:
    registry = load_registry(root)
    by_readme: dict[Path, list[str]] = {}
    valid_ids: list[str] = []
    errors: list[dict[str, str]] = []
    known_readmes = managed_badge_files(root)
    known_readmes.update(normalize_repo_path(root, value) for value in (extra_readmes or []))
    for entry in registry.get("promises", []):
        promise_id = str(entry["id"])
        if entry.get("readme"):
            known_readmes.add(normalize_repo_path(root, str(entry["readme"])))
        try:
            contract = load_contract(root, promise_id, require_tools=False)
            state = load_state(root, promise_id)
        except PromiseError as exc:
            error = str(exc)
            errors.append({"id": promise_id, "error": error})
            if "contract schema 1" in error and entry.get("readme"):
                readme = normalize_repo_path(root, str(entry["readme"]))
                known_readmes.add(readme)
                target = f".promises/{promise_id}/PROMISE.md"
                existing = readme.read_text() if readme.exists() else ""
                retained = next(
                    (match.group(0) for match in managed_badge_items(existing) if match.group(1).endswith(target)),
                    None,
                )
                if retained:
                    by_readme.setdefault(readme, []).append(retained)
            continue
        readme = normalize_repo_path(root, contract["badge"]["readme"])
        known_readmes.add(readme)
        valid, _ = promise_validity(root, promise_id, contract, state)
        if not valid:
            continue
        by_readme.setdefault(readme, []).append(badge_markdown(root, promise_id, contract, state))
        valid_ids.append(promise_id)
    updated: list[str] = []
    for readme in sorted(known_readmes):
        existing = readme.read_text() if readme.exists() else ""
        badges = sorted(by_readme.get(readme, []))
        revised = insert_badge_block(existing, badges)
        if revised != existing:
            readme.parent.mkdir(parents=True, exist_ok=True)
            readme.write_text(revised)
            updated.append(readme.relative_to(root).as_posix())
    return {"updated": updated, "valid": sorted(valid_ids), "errors": errors}


def promise_document(contract: dict[str, Any]) -> str:
    gates = "\n".join(
        f"- `{gate['id']}` — {gate['description']}" for gate in contract["acceptance_gates"]
    )
    includes = "\n".join(f"- `{item}`" for item in contract["scope"]["include"])
    excludes = "\n".join(f"- `{item}`" for item in contract["scope"].get("exclude", [])) or "- None beyond Promise state."
    steps = "\n".join(
        f"{number}. {step}" for number, step in enumerate(contract["operation"]["plan_steps"], start=1)
    )
    tools = "\n".join(
        f"- `{tool['path']}` — {tool['purpose']}" for tool in contract.get("tools", [])
    ) or "- No promise-local executable tools are required initially."
    human = "required" if contract["human_signoff_required"] else "not required"
    economics = contract["economics"]
    approval = "required for every run" if economics["approval_required"] else "not required"
    max_cost = economics.get("max_run_cost_usd")
    cap = "not set" if max_cost is None else f"${max_cost:,.2f}"
    staging = economics.get("staging_plan") or "Not required for this cost band."
    return f"""# Promise: {contract['name']}

{contract['statement']}

## Scope

Included:

{includes}

Excluded:

{excludes}

## Fulfillment gates

{gates}

Human sign-off is {human}. A fulfilled run is pinned to its reviewed commit,
subject fingerprint, contract and tool fingerprint, verifier, evidence, and
resource metrics.

## Operating plan

{steps}

Cadence: **{contract['operation']['cadence']}**

Runner: **{contract['operation']['runner']}**

Default authority: **{contract['operation']['authority']}**

## Economics and evidence

Estimated full-run cost: **{economics['full_run_cost_band']}** (USD band)

Estimate basis: {economics['estimate_basis']}

Per-run approval: **{approval}**. Maximum run cost: **{cap}**.

Staging plan: {staging}

Evidence mode: **{contract['evidence']['mode']}**. Every fulfillment gate needs a
passing event linked to durable evidence.

Verifier: **{contract['verifier']['identity']}**

## Promise-local tools

{tools}

## State

Current machine-readable state is in `state.json`. Compact run summaries and
append-only events live under `runs/`; latest per-subject coverage is in
`coverage.json`. Large or sensitive raw artifacts belong in ignored `.work/`,
with hashes or durable external references recorded as evidence.
"""


def prepare_init(
    root: Path, args: argparse.Namespace
) -> tuple[str, Path, dict[str, Any], dict[str, Any]]:
    promise_id = validate_id(args.promise_id)
    base = promise_root(root, promise_id)
    if base.exists():
        raise PromiseError(f"promise already exists: {base}")
    registry = load_registry(root)
    gates = []
    for value in args.gate:
        gate_id, description = parse_key_value(value, "gate")
        gates.append({"id": gate_id, "description": description})
    if len({gate["id"] for gate in gates}) != len(gates):
        raise PromiseError("acceptance gate ids must be unique")
    if args.human_signoff and "human-signoff" not in {gate["id"] for gate in gates}:
        raise PromiseError("--human-signoff requires a human-signoff=... acceptance gate")
    tools = [parse_tool(value) for value in args.tool]
    budgets = {
        "max_steps": args.max_steps,
        "max_tool_calls": args.max_tool_calls,
        "max_tokens": args.max_tokens,
        "max_wall_seconds": args.max_wall_seconds,
    }
    contract: dict[str, Any] = {
        "schema_version": SCHEMA_VERSION,
        "id": promise_id,
        "name": args.name,
        "statement": args.statement,
        "created_at": iso_time(),
        "scope": {
            "include": args.include,
            "exclude": args.exclude or [],
        },
        "acceptance_gates": gates,
        "human_signoff_required": args.human_signoff,
        "invalidation": {"scope_change": True, "max_age_hours": args.max_age_hours},
        "evidence": {"mode": args.evidence_mode},
        "verifier": {"identity": args.verifier},
        "economics": {
            "full_run_cost_band": args.cost_band,
            "estimate_basis": args.cost_basis,
            "approval_required": args.approval_required,
            "max_run_cost_usd": args.max_run_cost_usd,
            "staging_plan": args.staging_plan,
        },
        "operation": {
            "cadence": args.cadence,
            "runner": args.runner,
            "authority": args.authority,
            "plan_steps": args.plan_step,
            "budgets": budgets,
        },
        "tools": tools,
        "badge": {"label": args.badge_label or f"Promise: {args.name}", "readme": args.readme},
    }
    validate_contract(root, promise_id, contract, require_tools=False)
    return promise_id, base, contract, registry


def command_init_locked(args: argparse.Namespace) -> None:
    root = repo_root()
    promise_id, base, contract, registry = prepare_init(root, args)
    registry_path = root / ".promises" / "registry.json"
    base.mkdir(parents=True)
    (base / "tools").mkdir()
    (base / "runs").mkdir()
    (base / "evidence").mkdir()
    (base / ".work").mkdir()
    write_json(base / "contract.json", contract)
    (base / "PROMISE.md").write_text(promise_document(contract))
    state = {
        "schema_version": SCHEMA_VERSION,
        "id": promise_id,
        "status": "configured",
        "active_run": None,
        "last_run": None,
        "last_fulfillment": None,
        "lifetime_metrics": {
            "runs": 0,
            "steps": 0,
            "tool_calls": 0,
            "tokens_known": 0,
            "tokens_complete": True,
            "cost_usd_known": 0.0,
            "cost_complete": True,
            "wall_seconds": 0.0,
        },
        "updated_at": iso_time(),
    }
    write_json(base / "state.json", state)
    write_json(base / "coverage.json", {"schema_version": SCHEMA_VERSION, "subjects": {}})
    ignore = root / ".promises" / ".gitignore"
    ignore_lines = ignore.read_text().splitlines() if ignore.exists() else []
    for required in ["**/.work/", "**/__pycache__/", "**/*.py[cod]", "**/.*.tmp"]:
        if required not in ignore_lines:
            ignore_lines.append(required)
    ignore.write_text("\n".join(ignore_lines) + "\n")
    promises = registry.setdefault("promises", [])
    if not isinstance(promises, list):
        raise PromiseError("registry promises field must be a list")
    promises.append(
        {
            "id": promise_id,
            "name": args.name,
            "contract": f".promises/{promise_id}/contract.json",
            "readme": args.readme,
        }
    )
    registry["promises"] = sorted(registry["promises"], key=lambda item: item["id"])
    write_json(registry_path, registry)
    print(json.dumps({"id": promise_id, "path": str(base), "status": "configured"}))


def command_init(args: argparse.Namespace) -> None:
    root = repo_root()
    prepare_init(root, args)
    with mutation_lock(root):
        command_init_locked(args)


def command_validate(args: argparse.Namespace) -> None:
    root = repo_root()
    registry_path = root / ".promises" / "registry.json"
    if args.promise_id:
        registered_entry(root, args.promise_id)
    if not registry_path.exists():
        print(json.dumps({"promises": []}, indent=2))
        return
    registry = load_registry(root)
    ids = [args.promise_id] if args.promise_id else [str(entry["id"]) for entry in registry.get("promises", [])]
    results = []
    for promise_id in ids:
        try:
            contract = load_contract(
                root,
                promise_id,
                require_tools=not args.allow_incomplete_tools,
            )
            state = load_state(root, promise_id)
        except PromiseError as exc:
            if args.promise_id:
                raise
            results.append({"id": promise_id, "error": str(exc)})
            continue
        if state.get("id") != promise_id:
            message = f"state id does not match directory for {promise_id}"
            if args.promise_id:
                raise PromiseError(message)
            results.append({"id": promise_id, "error": message})
            continue
        subject, files = subject_fingerprint(root, contract)
        if files == 0:
            message = "promise scope resolves to zero files; fix scope.include before running"
            if args.promise_id:
                raise PromiseError(message)
            results.append({"id": promise_id, "subject_files": 0, "error": message})
            continue
        try:
            mechanism = mechanism_fingerprint(root, promise_id)
            mechanism_error = None
        except PromiseError as exc:
            if not args.allow_incomplete_tools:
                raise
            mechanism = None
            mechanism_error = str(exc)
        valid, reason = promise_validity(root, promise_id, contract, state)
        result = {
            "id": promise_id,
            "subject_files": files,
            "subject_fingerprint": subject,
            "mechanism_fingerprint": mechanism,
            "fulfilled": valid,
            "reason": reason,
        }
        if mechanism_error:
            result["mechanism_error"] = mechanism_error
        results.append(result)
    print(json.dumps({"promises": results}, indent=2))


def wall_segment_seconds(start_value: Any, end_value: Any) -> float:
    try:
        start = parse_time(str(start_value))
        end = parse_time(str(end_value))
    except (TypeError, ValueError) as exc:
        raise PromiseError("active run has invalid wall-time segment state") from exc
    return max(0.0, (end - start).total_seconds())


def active_wall_total(run: dict[str, Any]) -> float:
    value = run.get("active_wall_seconds", 0.0)
    if isinstance(value, bool) or not isinstance(value, (int, float)) or value < 0:
        raise PromiseError("active run has invalid accumulated wall time")
    return float(value)


def command_start(args: argparse.Namespace) -> None:
    root = repo_root()
    registered_entry(root, args.promise_id)
    contract = load_contract(root, args.promise_id)
    state = load_state(root, args.promise_id)
    if state.get("active_run"):
        run_id = str(state["active_run"])
        run_path, run, state = active_run(root, args.promise_id, run_id)
        resumed_at = utc_now()
        segment_start = run.get("active_segment_started_at", run.get("started_at"))
        last_activity = run.get("last_activity_at", segment_start)
        prior_segment = wall_segment_seconds(segment_start, last_activity)
        run["active_wall_seconds"] = round(active_wall_total(run) + prior_segment, 6)
        run["active_segment_started_at"] = iso_time(resumed_at)
        run["last_activity_at"] = iso_time(resumed_at)
        run["resume_count"] = int(run.get("resume_count", 0)) + 1
        write_json(run_path, run)
        append_jsonl(
            run_path.parent / "events.jsonl",
            {
                "at": iso_time(resumed_at),
                "kind": "run-resume",
                "result": "info",
                "prior_segment_wall_seconds": prior_segment,
            },
        )
        state["updated_at"] = iso_time(resumed_at)
        write_json(promise_root(root, args.promise_id) / "state.json", state)
        print(
            json.dumps(
                {
                    "id": args.promise_id,
                    "run_id": run_id,
                    "status": "resume",
                    "active_wall_seconds": run["active_wall_seconds"],
                }
            )
        )
        return
    approval_ref = (args.approval_ref or "").strip()
    if contract["economics"]["approval_required"] and not approval_ref:
        raise PromiseError("this Promise cost band requires --approval-ref before a run can start")
    if not run_git(root, "rev-parse", "--verify", "HEAD", check=False):
        raise PromiseError("commit the repository at least once before starting a Promise run")
    dirty = subject_dirty_paths(root, contract)
    if dirty and not args.allow_dirty:
        preview = ", ".join(dirty[:10])
        raise PromiseError(f"promise scope has uncommitted changes; commit or pass --allow-dirty: {preview}")
    start = utc_now()
    run_id = f"{start.strftime('%Y%m%dT%H%M%SZ')}-{uuid.uuid4().hex[:8]}"
    base = promise_root(root, args.promise_id)
    run_dir = base / "runs" / run_id
    subject, files = subject_fingerprint(root, contract)
    if files == 0:
        raise PromiseError("promise scope resolves to zero files; fix scope.include before starting")
    run_dir.mkdir(parents=True)
    mechanism = mechanism_fingerprint(root, args.promise_id)
    run = {
        "schema_version": SCHEMA_VERSION,
        "promise_id": args.promise_id,
        "run_id": run_id,
        "status": "running",
        "verdict": None,
        "started_at": iso_time(start),
        "active_segment_started_at": iso_time(start),
        "last_activity_at": iso_time(start),
        "active_wall_seconds": 0.0,
        "resume_count": 0,
        "ended_at": None,
        "base_commit": run_git(root, "rev-parse", "HEAD"),
        "subject_fingerprint": subject,
        "subject_file_count": files,
        "mechanism_fingerprint": mechanism,
        "approval_ref": approval_ref or None,
        "dirty_subject_at_start": dirty,
        "observed_metrics": {
            "steps": 0,
            "tool_calls": 0,
            "tokens_known": 0,
            "token_samples": 0,
            "cost_usd_known": 0.0,
            "cost_samples": 0,
        },
        "metrics": None,
        "gates": {},
        "verifier": None,
        "summary": None,
    }
    write_json(run_dir / "run.json", run)
    append_jsonl(
        run_dir / "events.jsonl",
        {
            "at": iso_time(start),
            "kind": "run-start",
            "result": "info",
            "commit": run["base_commit"],
            "approval_ref": run["approval_ref"],
        },
    )
    state["status"] = "running"
    state["active_run"] = run_id
    state["last_run"] = run_id
    state["updated_at"] = iso_time(start)
    write_json(base / "state.json", state)
    print(json.dumps({"id": args.promise_id, "run_id": run_id, "status": "started"}))


def active_run(root: Path, promise_id: str, run_id: str) -> tuple[Path, dict[str, Any], dict[str, Any]]:
    state = load_state(root, promise_id)
    if state.get("active_run") != run_id:
        raise PromiseError(f"{run_id} is not the active run for {promise_id}")
    run_path = promise_root(root, promise_id) / "runs" / run_id / "run.json"
    run = read_json(run_path)
    if run.get("status") != "running":
        raise PromiseError(f"run is not active: {run_id}")
    return run_path, run, state


def command_record(args: argparse.Namespace) -> None:
    root = repo_root()
    registered_entry(root, args.promise_id)
    contract = load_contract(root, args.promise_id, require_tools=False)
    run_path, run, _ = active_run(root, args.promise_id, args.run_id)
    gates = sorted(set(args.gate or []))
    expected_gates = {str(gate["id"]) for gate in contract["acceptance_gates"]}
    unknown_gates = sorted(set(gates) - expected_gates)
    if unknown_gates:
        raise PromiseError(f"record references unknown gates: {unknown_gates}")
    if gates and args.result == "pass" and not (args.evidence or "").strip():
        raise PromiseError("passing gate evidence requires --evidence")
    if args.cost_usd is not None and args.cost_usd < 0:
        raise PromiseError("recorded cost_usd cannot be negative")
    observed = run["observed_metrics"]
    projected_cost = float(observed.get("cost_usd_known", 0.0))
    if args.cost_usd is not None:
        projected_cost += args.cost_usd
    max_cost = contract["economics"].get("max_run_cost_usd")
    if max_cost is not None and projected_cost > max_cost and not (
        args.kind == "blocker" and args.result == "blocked"
    ):
        raise PromiseError(
            f"cost would exceed Promise cap (${projected_cost:.2f} > ${max_cost:.2f}); "
            "stop work or record the actual overrun as a blocked blocker event"
        )
    if gates and args.result == "pass" and args.kind not in GATE_EVIDENCE_KINDS:
        raise PromiseError(f"{args.kind} events cannot serve as passing gate evidence")
    if gates and args.result == "pass" and not evidence_reference_valid(
        root, args.promise_id, contract["evidence"]["mode"], args.evidence
    ):
        raise PromiseError(f"invalid {contract['evidence']['mode']} gate evidence reference: {args.evidence}")
    event = {
        "at": iso_time(),
        "kind": args.kind,
        "result": args.result,
        "subject": args.subject,
        "summary": args.summary,
        "evidence": args.evidence,
        "gates": gates,
        "commit": run_git(root, "rev-parse", "HEAD"),
        "metrics": {
            "steps": args.steps,
            "tool_calls": args.tool_calls,
            "tokens": args.tokens,
            "cost_usd": args.cost_usd,
        },
    }
    append_jsonl(run_path.parent / "events.jsonl", event)
    observed["steps"] += args.steps
    observed["tool_calls"] += args.tool_calls
    if args.tokens is not None:
        observed["tokens_known"] += args.tokens
        observed["token_samples"] += 1
    if args.cost_usd is not None:
        observed["cost_usd_known"] = round(projected_cost, 6)
        observed["cost_samples"] = int(observed.get("cost_samples", 0)) + 1
    run["last_activity_at"] = event["at"]
    write_json(run_path, run)
    if args.subject and args.kind in {"coverage", "inspection", "visual", "test"}:
        coverage_path = promise_root(root, args.promise_id) / "coverage.json"
        coverage = read_json(coverage_path)
        coverage.setdefault("subjects", {})[args.subject] = {
            "reviewed_at": event["at"],
            "run_id": args.run_id,
            "commit": event["commit"],
            "kind": args.kind,
            "result": args.result,
            "summary": args.summary,
            "evidence": args.evidence,
        }
        write_json(coverage_path, coverage)
    print(json.dumps({"run_id": args.run_id, "recorded": args.kind, "gates": gates, "metrics": observed}))


def passing_evidence_gates(root: Path, promise_id: str, evidence_mode: str, path: Path) -> set[str]:
    latest: dict[str, bool] = {}
    try:
        lines = path.read_text().splitlines()
    except FileNotFoundError as exc:
        raise PromiseError(f"missing run event log: {path}") from exc
    for line_number, line in enumerate(lines, start=1):
        if not line.strip():
            continue
        try:
            event = json.loads(line)
        except json.JSONDecodeError as exc:
            raise PromiseError(f"invalid JSON in {path}:{line_number}: {exc}") from exc
        if not isinstance(event, dict):
            raise PromiseError(f"expected a JSON object in {path}:{line_number}")
        evidence = event.get("evidence")
        gates = event.get("gates")
        if isinstance(gates, list) and event.get("result") != "info":
            passed = (
                event.get("result") == "pass"
                and event.get("kind") in GATE_EVIDENCE_KINDS
                and isinstance(evidence, str)
                and bool(evidence.strip())
            )
            if passed:
                passed = evidence_reference_valid(root, promise_id, evidence_mode, evidence)
            for gate in gates:
                if isinstance(gate, str):
                    latest[gate] = passed
    return {gate for gate, passed in latest.items() if passed}


def parse_gate_results(values: list[str]) -> dict[str, str]:
    results: dict[str, str] = {}
    for value in values:
        gate_id, result = parse_key_value(value, "gate result")
        if result not in {"pass", "fail", "blocked"}:
            raise PromiseError(f"gate result must be pass, fail, or blocked: {value}")
        if gate_id in results:
            raise PromiseError(f"duplicate gate result: {gate_id}")
        results[gate_id] = result
    return results


def final_metrics(args: argparse.Namespace, run: dict[str, Any], ended: dt.datetime) -> dict[str, Any]:
    observed = run["observed_metrics"]
    steps = args.steps_total if args.steps_total is not None else observed["steps"]
    tool_calls = args.tool_calls_total if args.tool_calls_total is not None else observed["tool_calls"]
    if steps < observed["steps"] or tool_calls < observed["tool_calls"]:
        raise PromiseError("final metric totals cannot be lower than observed event totals")
    if args.tokens_total is not None:
        if args.tokens_total < observed["tokens_known"]:
            raise PromiseError("final token total cannot be lower than observed tokens")
        tokens: int | None = args.tokens_total
        token_accounting = "runtime-total"
    elif observed["token_samples"]:
        tokens = observed["tokens_known"]
        token_accounting = "partial-event-log"
    else:
        tokens = None
        token_accounting = "unavailable"
    observed_cost = float(observed.get("cost_usd_known", 0.0))
    if args.cost_usd_total is not None:
        if args.cost_usd_total < observed_cost or args.cost_usd_total < 0:
            raise PromiseError("final cost total cannot be negative or lower than observed cost")
        cost_usd: float | None = round(args.cost_usd_total, 6)
        cost_accounting = "runtime-total"
    elif observed.get("cost_samples"):
        cost_usd = round(observed_cost, 6)
        cost_accounting = "partial-event-log"
    else:
        cost_usd = None
        cost_accounting = "unavailable"
    segment_start = run.get("active_segment_started_at", run.get("started_at"))
    wall = active_wall_total(run) + wall_segment_seconds(segment_start, iso_time(ended))
    return {
        "steps": steps,
        "steps_source": "runtime-total" if args.steps_total is not None else "event-log",
        "tool_calls": tool_calls,
        "tool_calls_source": "runtime-total" if args.tool_calls_total is not None else "event-log",
        "tokens": tokens,
        "token_accounting": token_accounting,
        "cost_usd": cost_usd,
        "cost_accounting": cost_accounting,
        "wall_seconds": round(wall, 6),
        "wall_accounting": "active-segments",
    }


def enforce_budgets(contract: dict[str, Any], metrics: dict[str, Any]) -> None:
    budgets = contract["operation"]["budgets"]
    remedy = "; finish the active run with --verdict budget-exhausted and non-passing gate results"
    comparisons = [
        ("max_steps", "steps"),
        ("max_tool_calls", "tool_calls"),
        ("max_wall_seconds", "wall_seconds"),
    ]
    for budget_name, metric_name in comparisons:
        limit = budgets.get(budget_name)
        if limit is not None and metrics[metric_name] > limit:
            raise PromiseError(
                f"{metric_name} exceeded promise budget ({metrics[metric_name]} > {limit}){remedy}"
            )
    if budgets.get("max_tokens") is not None:
        if metrics["tokens"] is None or metrics["token_accounting"] != "runtime-total":
            raise PromiseError(f"max_tokens requires an exact --tokens-total at finish{remedy}")
        if metrics["tokens"] > budgets["max_tokens"]:
            raise PromiseError(
                f"tokens exceeded promise budget ({metrics['tokens']} > {budgets['max_tokens']}){remedy}"
            )
    max_cost = contract["economics"].get("max_run_cost_usd")
    if max_cost is not None:
        if metrics["cost_usd"] is None or metrics["cost_accounting"] != "runtime-total":
            raise PromiseError(f"max_run_cost_usd requires an exact --cost-usd-total at finish{remedy}")
        if metrics["cost_usd"] > max_cost:
            raise PromiseError(
                f"cost exceeded Promise cap (${metrics['cost_usd']:.2f} > ${max_cost:.2f}){remedy}"
            )


def command_finish(args: argparse.Namespace) -> None:
    root = repo_root()
    registered_entry(root, args.promise_id)
    contract = load_contract(root, args.promise_id, require_tools=False)
    run_path, run, state = active_run(root, args.promise_id, args.run_id)
    ended = utc_now()
    gates = parse_gate_results(args.gate)
    expected = {str(gate["id"]) for gate in contract["acceptance_gates"]}
    if set(gates) != expected:
        missing = sorted(expected - set(gates))
        extra = sorted(set(gates) - expected)
        raise PromiseError(f"gate results must match contract; missing={missing}, extra={extra}")
    metrics = final_metrics(args, run, ended)
    if args.verdict == "fulfilled":
        if any(result != "pass" for result in gates.values()):
            raise PromiseError("a fulfilled run requires every acceptance gate to pass")
        expected_verifier = contract["verifier"]["identity"]
        if args.verifier.strip() != expected_verifier:
            raise PromiseError(f"fulfilled verifier must exactly match contract identity: {expected_verifier}")
        if run["observed_metrics"]["steps"] < 1:
            raise PromiseError("a fulfilled run requires recorded evidence events")
        evidence_gates = passing_evidence_gates(
            root,
            args.promise_id,
            contract["evidence"]["mode"],
            run_path.parent / "events.jsonl",
        )
        missing_evidence = sorted(expected - evidence_gates)
        if missing_evidence:
            raise PromiseError(f"fulfilled gates need passing durable evidence: {missing_evidence}")
        subject, subject_files = subject_fingerprint(root, contract)
        if subject_files == 0:
            raise PromiseError("a fulfilled run cannot attest to an empty promise scope")
        if subject != run["subject_fingerprint"]:
            raise PromiseError("promise scope changed during the run; start a new run")
        mechanism = mechanism_fingerprint(root, args.promise_id)
        if mechanism != run["mechanism_fingerprint"]:
            raise PromiseError("promise contract or tools changed during the run; start a new run")
        enforce_budgets(contract, metrics)
    run["status"] = "finished"
    run["verdict"] = args.verdict
    run["ended_at"] = iso_time(ended)
    run["metrics"] = metrics
    run["gates"] = gates
    run["verifier"] = args.verifier
    run["summary"] = args.summary
    write_json(run_path, run)
    append_jsonl(
        run_path.parent / "events.jsonl",
        {
            "at": iso_time(ended),
            "kind": "run-finish",
            "result": args.verdict,
            "summary": args.summary,
            "verifier": args.verifier,
            "gates": gates,
            "metrics": metrics,
        },
    )
    state["status"] = args.verdict
    state["active_run"] = None
    state["last_run"] = args.run_id
    state["updated_at"] = iso_time(ended)
    lifetime = state["lifetime_metrics"]
    lifetime["runs"] += 1
    lifetime["steps"] += metrics["steps"]
    lifetime["tool_calls"] += metrics["tool_calls"]
    lifetime["wall_seconds"] = round(lifetime["wall_seconds"] + metrics["wall_seconds"], 3)
    if metrics["tokens"] is not None:
        lifetime["tokens_known"] += metrics["tokens"]
    if metrics["token_accounting"] != "runtime-total":
        lifetime["tokens_complete"] = False
    if metrics["cost_usd"] is not None:
        lifetime["cost_usd_known"] = round(lifetime["cost_usd_known"] + metrics["cost_usd"], 6)
    if metrics["cost_accounting"] != "runtime-total":
        lifetime["cost_complete"] = False
    if args.verdict == "fulfilled":
        state["last_fulfillment"] = {
            "run_id": args.run_id,
            "fulfilled_at": iso_time(ended),
            "validated_commit": run["base_commit"],
            "subject_fingerprint": run["subject_fingerprint"],
            "mechanism_fingerprint": run["mechanism_fingerprint"],
            "verifier": args.verifier,
            "gates": gates,
            "metrics": metrics,
            "summary": args.summary,
            "dirty_subject_at_start": run["dirty_subject_at_start"],
            "approval_ref": run.get("approval_ref"),
        }
    elif args.verdict == "not-fulfilled" or any(result == "fail" for result in gates.values()):
        state["last_fulfillment"] = None
    write_json(promise_root(root, args.promise_id) / "state.json", state)
    try:
        badges = render_badges(root)
    except (OSError, PromiseError) as exc:
        raise PromiseError(
            f"run finished, but badge reconciliation failed: {exc}; repair the registry or README, "
            "then run promise.py badge"
        ) from exc
    print(
        json.dumps(
            {"id": args.promise_id, "run_id": args.run_id, "verdict": args.verdict, "metrics": metrics, "badges": badges},
            indent=2,
        )
    )


def command_status(args: argparse.Namespace) -> None:
    root = repo_root()
    registry_path = root / ".promises" / "registry.json"
    if not registry_path.exists() and not args.promise_id:
        print(json.dumps({"promises": []}, indent=2))
        return
    registry = load_registry(root)
    registered_ids = {str(entry["id"]) for entry in registry.get("promises", [])}
    ids = [args.promise_id] if args.promise_id else [str(entry["id"]) for entry in registry.get("promises", [])]
    statuses = []
    for promise_id in ids:
        try:
            contract = load_contract(root, promise_id, require_tools=False)
            state = load_state(root, promise_id)
        except PromiseError as exc:
            error = str(exc)
            status = "migration-required" if "schema 1" in error else "invalid"
            statuses.append({"id": promise_id, "status": status, "error": error})
            continue
        valid, reason = promise_validity(root, promise_id, contract, state)
        registered = promise_id in registered_ids
        if valid and not registered:
            reason = "valid attestation retained; promise is retired or unregistered"
        statuses.append(
            {
                "id": promise_id,
                "name": contract["name"],
                "status": state["status"],
                "registered": registered,
                "active_run": state["active_run"],
                "badge_valid": valid and registered,
                "validity": reason,
                "last_fulfillment": state.get("last_fulfillment"),
                "lifetime_metrics": state["lifetime_metrics"],
            }
        )
    print(json.dumps({"promises": statuses}, indent=2))


def registry_contracts_schema_2(
    root: Path,
    registry: dict[str, Any],
    candidate_id: str | None = None,
    candidate: dict[str, Any] | None = None,
) -> bool:
    for entry in registry.get("promises", []):
        entry_id = str(entry["id"])
        try:
            entry_contract = (
                candidate
                if entry_id == candidate_id and candidate is not None
                else read_json(promise_root(root, entry_id) / "contract.json")
            )
        except PromiseError:
            return False
        if entry_contract.get("schema_version") != SCHEMA_VERSION:
            return False
    return True


def command_migrate(args: argparse.Namespace) -> None:
    root = repo_root()
    base = promise_root(root, args.promise_id)
    if not base.is_dir():
        raise PromiseError(f"promise does not exist: {args.promise_id}")
    contract_path = base / "contract.json"
    state_path = base / "state.json"
    coverage_path = base / "coverage.json"
    contract = read_json(contract_path)
    state = read_json(state_path)
    coverage = read_json(coverage_path)
    versions = {
        "contract": contract.get("schema_version"),
        "state": state.get("schema_version"),
        "coverage": coverage.get("schema_version"),
    }
    if all(version == SCHEMA_VERSION for version in versions.values()):
        validate_contract(root, args.promise_id, contract)
        load_state(root, args.promise_id)
        registry = load_registry(root)
        all_schema_2 = registry_contracts_schema_2(root, registry)
        if all_schema_2 and registry.get("schema_version") != SCHEMA_VERSION:
            registry["schema_version"] = SCHEMA_VERSION
            write_json(root / ".promises" / "registry.json", registry)
        valid, _ = promise_validity(root, args.promise_id, contract, state)
        registered = any(str(entry["id"]) == args.promise_id for entry in registry.get("promises", []))
        badges = {
            "updated": []
            if valid
            else remove_promise_badge(root, args.promise_id, contract["badge"]["readme"]),
            "valid": [args.promise_id] if valid and registered else [],
            "deferred": not all_schema_2,
        }
        print(json.dumps({"id": args.promise_id, "status": "already-schema-2", "badges": badges}, indent=2))
        return
    if any(version not in SUPPORTED_SCHEMA_VERSIONS for version in versions.values()):
        raise PromiseError(f"cannot migrate unsupported schemas for {args.promise_id}: {versions}")
    if state.get("active_run"):
        raise PromiseError("finish the active run before migrating this Promise")
    if state.get("id") != args.promise_id:
        raise PromiseError(f"state does not match {args.promise_id}")
    verifier = contract.get("verifier") if isinstance(contract.get("verifier"), dict) else {}
    verifier["identity"] = args.verifier
    contract["schema_version"] = SCHEMA_VERSION
    contract["operation"]["runner"] = args.runner
    contract["evidence"] = {"mode": args.evidence_mode}
    contract["verifier"] = verifier
    contract["economics"] = {
        "full_run_cost_band": args.cost_band,
        "estimate_basis": args.cost_basis,
        "approval_required": args.approval_required,
        "max_run_cost_usd": args.max_run_cost_usd,
        "staging_plan": args.staging_plan,
    }
    if args.max_age_hours is not None:
        contract["invalidation"]["max_age_hours"] = args.max_age_hours
    validate_contract(root, args.promise_id, contract)
    lifetime = state.get("lifetime_metrics")
    if not isinstance(lifetime, dict):
        raise PromiseError(f"schema-1 state for {args.promise_id} needs lifetime_metrics")
    lifetime["cost_usd_known"] = 0.0
    lifetime["cost_complete"] = int(lifetime.get("runs", 0)) == 0
    state["schema_version"] = SCHEMA_VERSION
    state["updated_at"] = iso_time()
    coverage["schema_version"] = SCHEMA_VERSION
    registry = load_registry(root)
    all_schema_2 = registry_contracts_schema_2(root, registry, args.promise_id, contract)
    if all_schema_2:
        registry["schema_version"] = SCHEMA_VERSION
    # State and coverage are written before the contract so an interrupted
    # migration can always be resumed through this same command.
    write_json(state_path, state)
    write_json(coverage_path, coverage)
    (base / "PROMISE.md").write_text(promise_document(contract))
    write_json(contract_path, contract)
    write_json(root / ".promises" / "registry.json", registry)
    badges = (
        render_badges(root)
        if all_schema_2
        else {
            "updated": remove_promise_badge(root, args.promise_id, contract["badge"]["readme"]),
            "valid": [],
            "deferred": True,
        }
    )
    print(json.dumps({"id": args.promise_id, "status": "migrated", "badges": badges}, indent=2))


def command_retire(args: argparse.Namespace) -> None:
    root = repo_root()
    registry_path = root / ".promises" / "registry.json"
    registry = load_registry(root)
    entries = [entry for entry in registry.get("promises", []) if str(entry.get("id")) == args.promise_id]
    if not entries:
        raise PromiseError(f"promise is not registered: {args.promise_id}")
    entry = entries[0]
    readme = str(entry.get("readme", "README.md"))
    state_path = promise_root(root, args.promise_id) / "state.json"
    if state_path.exists():
        state = load_state(root, args.promise_id)
        if state.get("active_run"):
            raise PromiseError("finish the active run before retiring this promise")
        state["status"] = "retired"
        state["updated_at"] = iso_time()
        write_json(state_path, state)
    registry["promises"] = [
        item for item in registry.get("promises", []) if str(item.get("id")) != args.promise_id
    ]
    write_json(registry_path, registry)
    badges = render_badges(root, extra_readmes=[readme])
    print(json.dumps({"id": args.promise_id, "status": "retired", "badges": badges}, indent=2))


def command_restore(args: argparse.Namespace) -> None:
    root = repo_root()
    registry_path = root / ".promises" / "registry.json"
    registry = load_registry(root)
    if any(str(entry.get("id")) == args.promise_id for entry in registry.get("promises", [])):
        raise PromiseError(f"promise is already registered: {args.promise_id}")
    contract = load_contract(root, args.promise_id, require_tools=False)
    state = load_state(root, args.promise_id)
    retained_valid, _ = promise_validity(root, args.promise_id, contract, state)
    registry.setdefault("promises", []).append(
        {
            "id": args.promise_id,
            "name": contract["name"],
            "contract": f".promises/{args.promise_id}/contract.json",
            "readme": contract["badge"]["readme"],
        }
    )
    registry["promises"] = sorted(registry["promises"], key=lambda item: item["id"])
    write_json(registry_path, registry)
    if not state.get("active_run"):
        state["status"] = "fulfilled" if retained_valid else "configured"
    state["updated_at"] = iso_time()
    write_json(promise_root(root, args.promise_id) / "state.json", state)
    badges = render_badges(root)
    print(json.dumps({"id": args.promise_id, "status": "restored", "badges": badges}, indent=2))


def command_badge(_: argparse.Namespace) -> None:
    print(json.dumps(render_badges(repo_root()), indent=2))


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description=__doc__)
    subparsers = parser.add_subparsers(dest="command", required=True)

    init = subparsers.add_parser("init", help="Create a Promise contract after the setup interview")
    init.add_argument("promise_id")
    init.add_argument("--name", required=True)
    init.add_argument("--statement", required=True)
    init.add_argument(
        "--include",
        action="append",
        required=True,
        help="repo-relative glob; * stays in one path segment and ** spans directories",
    )
    init.add_argument(
        "--exclude",
        action="append",
        default=[],
        help="repo-relative glob; * stays in one path segment and ** spans directories",
    )
    init.add_argument("--gate", action="append", required=True, help="id=description; repeat for every gate")
    init.add_argument("--plan-step", action="append", required=True, help="repeat for each operating-plan step")
    init.add_argument("--tool", action="append", default=[], help="tools/path=purpose")
    init.add_argument("--cadence", default="manual")
    init.add_argument("--runner", required=True, help="runtime and operator responsible for invocation")
    init.add_argument("--authority", default="read-only")
    init.add_argument("--evidence-mode", required=True, choices=["internal", "external", "mixed"])
    init.add_argument("--verifier", required=True, help="exact identity required at fulfilled finish")
    init.add_argument("--cost-band", required=True, choices=COST_BANDS)
    init.add_argument("--cost-basis", required=True)
    init.add_argument("--approval-required", action="store_true")
    init.add_argument("--max-run-cost-usd", type=float)
    init.add_argument("--staging-plan")
    init.add_argument("--max-age-hours", type=float)
    init.add_argument("--max-steps", type=int)
    init.add_argument("--max-tool-calls", type=int)
    init.add_argument("--max-tokens", type=int)
    init.add_argument("--max-wall-seconds", type=float)
    init.add_argument("--human-signoff", action="store_true")
    init.add_argument("--readme", default="README.md")
    init.add_argument("--badge-label")
    init.set_defaults(func=command_init)

    validate = subparsers.add_parser("validate", help="Validate contracts, tools, state, and fingerprints")
    validate.add_argument("promise_id", nargs="?")
    validate.add_argument(
        "--allow-incomplete-tools",
        action="store_true",
        help="inspect setup state before all declared promise-local tools exist",
    )
    validate.set_defaults(func=command_validate)

    start = subparsers.add_parser("start", help="Start or resume one bounded Promise run")
    start.add_argument("promise_id")
    start.add_argument("--allow-dirty", action="store_true", help="record and review a dirty scoped workspace")
    start.add_argument("--approval-ref", help="durable approval receipt required by material cost bands")
    start.set_defaults(func=serialized(command_start))

    record = subparsers.add_parser("record", help="Append evidence and update compact coverage")
    record.add_argument("promise_id")
    record.add_argument("run_id")
    record.add_argument(
        "--kind",
        required=True,
        choices=["coverage", "inspection", "finding", "test", "visual", "tool", "decision", "verification", "blocker"],
    )
    record.add_argument("--result", required=True, choices=["pass", "fail", "info", "blocked"])
    record.add_argument("--subject")
    record.add_argument("--gate", action="append", help="contract gate supported by this evidence; repeat as needed")
    record.add_argument("--summary", required=True)
    record.add_argument("--evidence")
    record.add_argument("--steps", type=int, default=1)
    record.add_argument("--tool-calls", type=int, default=0)
    record.add_argument("--tokens", type=int)
    record.add_argument("--cost-usd", type=float, help="measured direct run cost incurred by this event")
    record.set_defaults(func=serialized(command_record))

    finish = subparsers.add_parser("finish", help="Close a run and issue badges only for verified fulfillment")
    finish.add_argument("promise_id")
    finish.add_argument("run_id")
    finish.add_argument("--verdict", required=True, choices=["fulfilled", "not-fulfilled", "blocked", "budget-exhausted"])
    finish.add_argument("--gate", action="append", required=True, help="id=pass|fail|blocked")
    finish.add_argument("--verifier", default="")
    finish.add_argument("--summary", required=True)
    finish.add_argument("--steps-total", type=int)
    finish.add_argument("--tool-calls-total", type=int)
    finish.add_argument("--tokens-total", type=int)
    finish.add_argument("--cost-usd-total", type=float, help="exact direct run cost required when a USD cap is set")
    finish.set_defaults(func=serialized(command_finish))

    status = subparsers.add_parser("status", help="Report Promise state without changing it")
    status.add_argument("promise_id", nargs="?")
    status.set_defaults(func=command_status)

    migrate = subparsers.add_parser("migrate", help="Explicitly migrate one schema-1 Promise contract")
    migrate.add_argument("promise_id")
    migrate.add_argument("--runner", required=True)
    migrate.add_argument("--evidence-mode", required=True, choices=["internal", "external", "mixed"])
    migrate.add_argument("--verifier", required=True)
    migrate.add_argument("--cost-band", required=True, choices=COST_BANDS)
    migrate.add_argument("--cost-basis", required=True)
    migrate.add_argument("--approval-required", action="store_true")
    migrate.add_argument("--max-run-cost-usd", type=float)
    migrate.add_argument("--staging-plan")
    migrate.add_argument("--max-age-hours", type=float)
    migrate.set_defaults(func=serialized(command_migrate))

    retire = subparsers.add_parser("retire", help="Deregister a Promise without deleting its evidence")
    retire.add_argument("promise_id")
    retire.set_defaults(func=serialized(command_retire))

    restore = subparsers.add_parser("restore", help="Re-register a retired Promise for future runs")
    restore.add_argument("promise_id")
    restore.set_defaults(func=serialized(command_restore))

    badge = subparsers.add_parser("badge", help="Reconcile the managed README badge block")
    badge.set_defaults(func=serialized(command_badge))
    return parser


def main() -> int:
    try:
        args = build_parser().parse_args()
        args.func(args)
        return 0
    except PromiseError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
