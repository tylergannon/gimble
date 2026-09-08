#!/usr/bin/env python3
"""Exercise the public CLI and retain the actual output, including failures."""
import json
from pathlib import Path
import subprocess
import sys


def check(workspace, mode):
    records = []
    for cents in (4999, 5000, 5001):
        shipping = 1200 if mode == "expedited" else (800 if cents < 5000 else 0)
        expected = {"subtotal_cents": cents, "shipping_cents": shipping,
                    "total_cents": cents + shipping, "mode": mode}
        command = [sys.executable, "quote.py", f"{cents / 100:.2f}", mode]
        result = subprocess.run(command, cwd=workspace, text=True,
                                capture_output=True, timeout=10)
        try:
            actual = json.loads(result.stdout)
        except json.JSONDecodeError:
            actual = None
        records.append({"command": command, "exit_code": result.returncode,
                        "stdout": result.stdout, "stderr": result.stderr,
                        "expected": expected,
                        "passed": result.returncode == 0 and actual == expected})
    return records


if __name__ == "__main__":
    mode = sys.argv[1]
    if mode not in ("standard", "expedited"):
        raise SystemExit("usage: python3 check.py standard|expedited")
    records = check(Path.cwd(), mode)
    Path(f"evidence-{mode}.json").write_text(json.dumps(records, indent=2) + "\n")
    print(json.dumps(records, indent=2))
    raise SystemExit(0 if all(record["passed"] for record in records) else 1)
