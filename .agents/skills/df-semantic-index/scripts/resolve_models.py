#!/usr/bin/env python3
"""Resolve this installed skill's model defaults, user overrides, then project overrides."""

from __future__ import annotations

import argparse
import json
import re
import subprocess
from pathlib import Path

try:
    import yaml
except ImportError:
    yaml = None


class ConfigError(ValueError):
    pass


class UniqueLoader(yaml.SafeLoader if yaml else object):
    """Reject duplicate keys instead of silently dropping an override."""


def unique_mapping(loader, node):
    result = {}
    for key_node, value_node in node.value:
        key = loader.construct_object(key_node)
        if not isinstance(key, str) or key in result:
            raise ConfigError("mapping keys must be unique strings")
        result[key] = loader.construct_object(value_node)
    return result


if yaml:
    UniqueLoader.add_constructor(yaml.resolver.BaseResolver.DEFAULT_MAPPING_TAG, unique_mapping)


def read_yaml(text, label):
    if not any(line.strip() and not line.lstrip().startswith("#") for line in text.splitlines()):
        return {}
    if yaml is None:
        raise ConfigError(f"{label}: reading YAML overrides requires PyYAML; "
                          "see references/models.md for installation")
    try:
        result = yaml.load(text, Loader=UniqueLoader)
    except (yaml.YAMLError, ConfigError) as exc:
        raise ConfigError(f"{label}: {exc}") from exc
    if result is None:
        return {}
    if not isinstance(result, dict):
        raise ConfigError(f"{label}: expected a YAML mapping")
    return result


def plain_default_fields(text, indent):
    """Read only the bundled, plain scalar model fields, never user YAML."""
    result = {}
    for field, value in re.findall(rf"^{' ' * indent}(model|reasoning_effort):\s*([^\n]*)$",
                                  text, re.MULTILINE):
        value = value.split(" #", 1)[0].strip()
        if field in result or not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._:/-]*", value):
            raise ConfigError("Bundled defaults require unique, plain scalar model/effort fields")
        result[field] = None if value == "null" else value
    return result


def plain_defaults(block):
    """Stdlib path for the controlled SKILL.md layout when PyYAML is absent."""
    section = re.search(r"^(model_defaults|agents):\n(.*?)(?=^\S|\Z)", block,
                        re.MULTILINE | re.DOTALL)
    if section is None:
        raise ConfigError("Missing bundled model defaults section")
    roles = {}
    for role, body in re.findall(r"^  ([a-z][a-z0-9-]*):\n(.*?)(?=^  [a-z][a-z0-9-]*:|\Z)",
                                section[2], re.MULTILINE | re.DOTALL):
        if role in roles:
            raise ConfigError(f"Duplicate bundled role: {role}")
        roles[role] = plain_default_fields(body, 4)
    if section[1] == "agents":
        main = re.search(r"^top_level_agent:\n(.*?)(?=^\S|\Z)", block,
                         re.MULTILINE | re.DOTALL)
        if main is None:
            raise ConfigError("Easy Loop workflow requires top_level_agent")
        roles["main"] = plain_default_fields(main[1], 2)
    return roles


def defaults(skill_dir):
    path = skill_dir / "SKILL.md"
    text = path.read_text()
    name = re.search(r"^name: (df-[a-z0-9-]+)$", text, re.MULTILINE)
    if name is None:
        raise ConfigError(f"{path}: missing skill name")
    candidates = []
    for block in re.findall(r"^```yaml\n(.*?)^```", text, re.MULTILINE | re.DOTALL):
        if not re.search(r"^(model_defaults|agents):", block, re.MULTILINE):
            continue
        if yaml is None:
            candidates.append(plain_defaults(block))
            continue
        data = read_yaml(block, path)
        if "model_defaults" in data:
            candidates.append(data["model_defaults"])
        elif "agents" in data:
            if not isinstance(data["agents"], dict) or not isinstance(data.get("top_level_agent"), dict):
                raise ConfigError(f"{path}: Easy Loop requires agents and top_level_agent mappings")
            specs = {**data["agents"], "main": data["top_level_agent"]}
            if any(not isinstance(spec, dict) for spec in specs.values()):
                raise ConfigError(f"{path}: each agent must be a mapping")
            candidates.append({role: {key: value for key, value in spec.items()
                                      if key in {"model", "reasoning_effort"}}
                               for role, spec in specs.items()})
    if len(candidates) != 1 or not isinstance(candidates[0], dict) or not candidates[0]:
        raise ConfigError(f"{path}: expected exactly one model defaults mapping")
    return name.group(1), candidates[0]


def project_root(cwd):
    try:
        result = subprocess.run(["git", "-C", str(cwd), "rev-parse", "--show-toplevel"],
                                capture_output=True, text=True, check=False)
    except FileNotFoundError:
        return cwd
    return Path(result.stdout.strip()) if result.returncode == 0 else cwd


def validate_fields(spec, label, partial=False):
    if not isinstance(spec, dict) or set(spec) - {"model", "reasoning_effort"}:
        raise ConfigError(f"{label}: expected model and/or reasoning_effort fields")
    if not partial and "model" not in spec:
        raise ConfigError(f"{label}: missing model")
    if "model" in spec and (not isinstance(spec["model"], str) or
                            not re.fullmatch(r"[A-Za-z0-9][A-Za-z0-9._:/-]*", spec["model"])):
        raise ConfigError(f"{label}: model must be a nonempty model identifier")
    if "reasoning_effort" in spec and spec["reasoning_effort"] not in (
            None, "none", "minimal", "low", "medium", "high", "xhigh", "max", "ultra"):
        raise ConfigError(f"{label}: invalid reasoning_effort")


def invocation(skill, role, spec):
    model = spec["model"]
    effort = spec.get("reasoning_effort")
    if model == "inherit":
        if role not in {"main", "reader"} or effort is not None:
            raise ConfigError(f"{role}: inherit is only valid for main/reader without effort")
        return None, []
    if model == "cli-default":
        if role != "gemini" or effort is not None:
            raise ConfigError("cli-default is only valid for the Gemini lane without effort")
        return "agy", []
    cli = next((cli for prefix, cli in (("claude-", "claude"), ("gpt-", "codex"),
                                        ("gemini-", "agy")) if model.startswith(prefix)), None)
    if cli is None:
        raise ConfigError(f"{role}: unsupported model prefix: {model}")
    if skill.startswith("df-easy-loop-") and cli == "agy":
        raise ConfigError("Easy Loop supports Claude and Codex models only")
    if role in {"claude", "codex", "gemini"} and cli != {"gemini": "agy"}.get(role, role):
        raise ConfigError(f"{role}: model must match the selected provider lane")
    allowed = {"claude": {"low", "medium", "high", "xhigh", "max"},
               "codex": {"none", "minimal", "low", "medium", "high", "xhigh", "max", "ultra"},
               "agy": {"low", "medium", "high"}}
    if effort is not None and effort not in allowed[cli]:
        raise ConfigError(f"{role}: reasoning_effort {effort} is not supported by {cli}")
    args = ["--model", model]
    if effort is not None:
        args += (["-c", f'model_reasoning_effort="{effort}"'] if cli == "codex"
                 else ["--effort", effort])
    return cli, args


def resolve(skill_dir, project_dir, user_config):
    skill, roles = defaults(skill_dir)
    sources = {}
    for role, spec in roles.items():
        validate_fields(spec, f"{skill}/{role}")
        sources[role] = {field: "skill default" for field in spec}
    project_config = project_dir / ".diffusion/skills/models.yaml"
    for path in (user_config, project_config):
        if not path.exists():
            continue
        data = read_yaml(path.read_text(), path)
        if set(data) - {"skills"} or not isinstance(data.get("skills", {}), dict):
            raise ConfigError(f"{path}: expected a top-level skills mapping")
        overrides = data.get("skills", {}).get(skill, {})
        if not isinstance(overrides, dict):
            raise ConfigError(f"{path}: {skill} must be a role mapping")
        for role, spec in overrides.items():
            if role not in roles:
                raise ConfigError(f"{path}: unknown role {skill}/{role}")
            validate_fields(spec, f"{path}: {skill}/{role}", partial=True)
            roles[role].update(spec)
            sources[role].update({field: str(path) for field in spec})
    resolved = {}
    for role, spec in roles.items():
        cli, args = invocation(skill, role, spec)
        resolved[role] = {**spec, "cli": cli, "model_args": args, "sources": sources[role]}
    return {"skill": skill, "project_dir": str(project_dir), "roles": resolved}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--project-dir", type=Path,
                        help="Target project root; defaults to Git root, or cwd outside Git")
    parser.add_argument("--user-config", type=Path,
                        default=Path.home() / ".config/diffusion/skills/models.yaml",
                        help="Alternate user file for isolated tests")
    args = parser.parse_args()
    root = (args.project_dir.expanduser().resolve() if args.project_dir else
            project_root(Path.cwd()).resolve())
    try:
        if not root.is_dir():
            raise ConfigError(f"Project directory does not exist: {root}")
        result = resolve(Path(__file__).resolve().parents[1], root,
                         args.user_config.expanduser().resolve())
        print(json.dumps(result, indent=2))
    except (ConfigError, OSError) as exc:
        parser.exit(2, f"Model configuration error: {exc}\n")


if __name__ == "__main__":
    main()
