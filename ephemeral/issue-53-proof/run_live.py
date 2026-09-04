#!/usr/bin/env python3
"""Run the real three-harness convergence fixture in an isolated directory."""
import hashlib
import json
import os
import pathlib
import shutil
import subprocess
import sys
import tempfile

source = pathlib.Path(__file__).resolve().parent
binary = pathlib.Path(sys.argv[1]).resolve()
root = pathlib.Path(tempfile.mkdtemp(prefix="tractor-issue53-live-"))
workspace = root / "workspace"
shutil.copytree(source / "fixture", workspace)
observers = root / "bin"
observers.mkdir()
native = {}
for name in ("claude", "agy", "codex"):
    native[name] = shutil.which(name)
    assert native[name], f"Missing native {name} executable"
    (observers / name).symlink_to(source / "observe_native.py")
environment = dict(os.environ,
                   PATH=str(observers) + os.pathsep + os.environ["PATH"],
                   ISSUE53_NATIVE_BINARIES=json.dumps(native),
                   ISSUE53_NATIVE_RECEIPTS=str(root / "native-selections.jsonl"))
metadata = {"root": str(root), "binary": str(binary),
            "binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
            "source_revision": subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()}
(root / "metadata.json").write_text(json.dumps(metadata, indent=2) + "\n")
print(json.dumps(metadata), flush=True)
command = [str(binary), "run", str(workspace / "workflow.yaml"),
           "--workdir", str(workspace), "--logs", str(root / "logs")]
with (root / "stdout.log").open("w") as stdout, (root / "stderr.log").open("w") as stderr:
    result = subprocess.run(command, env=environment, stdout=stdout, stderr=stderr)
metadata["exit_code"] = result.returncode
if result.returncode == 0:
    oracle = subprocess.run([sys.executable, str(source / "verify_quote.py"), str(workspace)],
                            text=True, capture_output=True)
    (root / "final-oracle.jsonl").write_text(oracle.stdout)
    (root / "final-oracle.stderr").write_text(oracle.stderr)
    metadata["oracle_exit_code"] = oracle.returncode
(root / "metadata.json").write_text(json.dumps(metadata, indent=2) + "\n")
print(json.dumps(metadata), flush=True)
sys.exit(result.returncode or metadata.get("oracle_exit_code", 0))
