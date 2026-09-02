#!/usr/bin/env python3
"""Run the deterministic gates for the all-skills-validate promise."""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[3]
PROMISE_ROOT = ROOT / ".promises" / "all-skills-validate"
MIRRORS = (ROOT / ".claude" / "skills", ROOT / ".agents" / "skills")


def run(*command: str, env: dict[str, str] | None = None) -> dict[str, Any]:
    completed = subprocess.run(
        command,
        cwd=ROOT,
        check=False,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    return {
        "command": list(command),
        "returncode": completed.returncode,
        "stdout": completed.stdout.strip(),
        "stderr": completed.stderr.strip(),
        "passed": completed.returncode == 0,
    }


def skill_names(root: Path) -> list[str]:
    if not root.is_dir():
        return []
    return sorted(
        path.name
        for path in root.iterdir()
        if path.is_dir() and path.name.startswith("df-")
    )


def python_helpers() -> list[str]:
    helpers = {"scripts/validate-skills.py"}
    for mirror in MIRRORS:
        for helper in mirror.glob("*/**/*.py"):
            if "__pycache__" not in helper.parts:
                helpers.add(helper.relative_to(ROOT).as_posix())
    return sorted(helpers)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", type=Path)
    args = parser.parse_args()

    mirror_inventory = {
        mirror.relative_to(ROOT).as_posix(): skill_names(mirror)
        for mirror in MIRRORS
    }
    inventory_values = list(mirror_inventory.values())
    inventory_passed = bool(inventory_values[0]) and all(
        names == inventory_values[0] for names in inventory_values[1:]
    )

    helpers = python_helpers()
    validator = run(sys.executable, "scripts/validate-skills.py")
    compile_env = os.environ.copy()
    compile_env["PYTHONPYCACHEPREFIX"] = str(
        ROOT / ".promises" / "all-skills-validate" / ".work" / "pycache"
    )
    compile_result = run(
        sys.executable,
        "-m",
        "py_compile",
        *helpers,
        env=compile_env,
    )
    gates = {
        "inventory": inventory_passed,
        "validator": validator["passed"],
        "python-compile": compile_result["passed"],
    }
    report = {
        "schema_version": 1,
        "generated_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "repository": ".",
        "mirrors": mirror_inventory,
        "skill_counts": {
            mirror: len(names) for mirror, names in mirror_inventory.items()
        },
        "python_helpers": helpers,
        "gates": gates,
        "commands": {
            "validator": validator,
            "python_compile": compile_result,
        },
        "passed": all(gates.values()),
    }
    rendered = json.dumps(report, indent=2, sort_keys=True) + "\n"
    if args.output:
        output = args.output if args.output.is_absolute() else ROOT / args.output
        output = output.resolve()
        try:
            output.relative_to(PROMISE_ROOT.resolve())
        except ValueError:
            parser.error(f"--output must stay within {PROMISE_ROOT.relative_to(ROOT)}")
        output.parent.mkdir(parents=True, exist_ok=True)
        output.write_text(rendered)
    print(rendered, end="")
    return 0 if report["passed"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
