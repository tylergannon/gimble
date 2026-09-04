#!/usr/bin/env python3
"""Independent final oracle, outside the implementation agent's workspace."""
import json
import pathlib
import subprocess
import sys

workdir = pathlib.Path(sys.argv[1])
for mode in ("standard", "expedited"):
    for subtotal in (4999, 5000, 5001):
        amount = f"{subtotal // 100}.{subtotal % 100:02}"
        run = subprocess.run([sys.executable, "quote.py", amount, mode],
                             cwd=workdir, text=True, capture_output=True)
        assert run.returncode == 0, (mode, amount, run.stderr)
        actual = json.loads(run.stdout)
        shipping = 1200 if mode == "expedited" else (800 if subtotal < 5000 else 0)
        expected = {"subtotal_cents": subtotal, "shipping_cents": shipping,
                    "total_cents": subtotal + shipping, "mode": mode}
        assert actual == expected, (mode, amount, actual, expected)
        print(json.dumps({"mode": mode, "subtotal": amount, "actual": actual, "passed": True}))
