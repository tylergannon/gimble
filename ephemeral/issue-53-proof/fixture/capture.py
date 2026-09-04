#!/usr/bin/env python3
"""Record actual application behavior without deciding whether it is correct."""
import json
import pathlib
import subprocess
import sys

records = []
for mode in ("standard", "expedited"):
    for subtotal in ("49.99", "50.00", "50.01"):
        command = [sys.executable, "quote.py", subtotal, mode]
        result = subprocess.run(command, text=True, capture_output=True)
        records.append({"command": command, "exit_code": result.returncode,
                        "stdout": result.stdout, "stderr": result.stderr})
pathlib.Path("evidence.json").write_text(json.dumps(records, indent=2) + "\n")
print("Captured six real CLI invocations in evidence.json")
