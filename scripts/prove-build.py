#!/usr/bin/env python3
"""Run the canonical loop with native agents; exit nonzero unless it is proved."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import runpy
import shutil
import signal
import subprocess
import sys
import tempfile
import time

REPO = Path(__file__).resolve().parents[1]
FIXTURE = REPO / "examples/loops/canonical"
ITEMS = ["Standard shipping", "Expedited shipping"]


def require(condition, message):
    if not condition:
        raise RuntimeError(message)


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(*args, cwd=REPO):
    return subprocess.check_output(args, cwd=cwd, text=True).strip()


def save(path, value):
    path.write_text(json.dumps(value, indent=2) + "\n")


def source_hash():
    digest = hashlib.sha256()
    names = run("git", "ls-files", "--cached", "--others", "--exclude-standard", "-z").split("\0")
    for name in sorted(set(names)):
        path = REPO / name
        if name.startswith(("ephemeral/", "reference/")) or not path.is_file():
            continue
        digest.update(name.encode() + b"\0" + path.read_bytes() + b"\0")
    return digest.hexdigest()


def ledger_contract(path):
    # The fixture uses single-line YAML scalars; the engine may reindent YAML
    # and add done, but may not change the authored acceptance contract.
    return "\n".join(line.strip() for line in path.read_text().splitlines()
                     if not re.fullmatch(r"\s*done: (true|false)\s*", line))


def verify_run(root):
    workspace = root / "workspace"
    events = [json.loads(line) for line in (root / "run/timeline.jsonl").read_text().splitlines()]
    require(any(e["type"] == "PipelineCompleted" for e in events), "pipeline did not complete")
    selected = [e["item"] for e in events if e["type"] == "LoopItemSelected"]
    require(list(dict.fromkeys(selected)) == ITEMS, f"wrong item dispatch order: {selected}")
    stages = [e for e in events if e["type"] == "StageCompleted"]
    require(sum(e["name"] == "implement" for e in stages) >= 2, "missing implementation laps")
    require(sum(e["name"] == "review" and e["next"] == "sprints" for e in stages) >= 2,
            "review did not return both items to engine validation")
    validations = [e for e in events if e["type"] == "LoopValidated"]
    require(len(validations) >= 2, "missing engine validation laps")
    require([v["item"] for v in validations[0]["validations"]] == ITEMS[:1],
            "first sprint was not validated separately")
    final = validations[-1]["validations"]
    require([v["item"] for v in final] == ITEMS, "last lap did not revalidate both items")
    require(all(v["passed"] and v["exit_code"] == 0 and
                v.get("infer", {}).get("verdict") == "pass" for v in final),
            "commands and evidence judges did not pass both items")
    verdicts = [e["verdict"] for e in events if e["type"] == "LoopEvaluated"]
    require(verdicts[0] == "not_done" and verdicts[-1] == "done",
            f"goal evaluator did not distinguish partial from complete work: {verdicts}")
    ledger = workspace / "docs/sprints/ledger.md"
    require(ledger_contract(ledger) == ledger_contract(FIXTURE / "docs/sprints/ledger.md"),
            "the acceptance ledger changed")
    require(len(re.findall(r"^\s+done: true$", ledger.read_text(), re.M)) == 2,
            "both ledger items were not closed")
    for path in FIXTURE.rglob("*"):
        if path.is_file() and path.name not in ("quote.py", "README.md", "ledger.md"):
            require(sha(path) == sha(workspace / path.relative_to(FIXTURE)),
                    f"acceptance file changed: {path.relative_to(FIXTURE)}")
    changed = run("git", "diff", "--name-only", "seed", "HEAD", cwd=workspace).splitlines()
    require(changed == ["quote.py"], f"agent commits changed files outside the application: {changed}")
    oracle = runpy.run_path(str(FIXTURE / "check.py"))["check"]
    results = [record for mode in ("standard", "expedited") for record in oracle(workspace, mode)]
    save(root / "final-cli.json", results)
    require(all(r["passed"] for r in results), "independent final CLI invocations failed")
    return {"selected_items": selected, "goal_verdicts": verdicts,
            "engine_validations": validations,
            "review_routes": [e for e in stages if e["name"] == "review"],
            "final_cli": str(root / "final-cli.json")}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--binary", type=Path, help="exact candidate binary; otherwise build current source")
    parser.add_argument("--output", type=Path, help="new evidence directory (default: fresh temporary directory)")
    parser.add_argument("--timeout", type=int, default=1200, help="whole-run timeout in seconds (default: 1200)")
    args = parser.parse_args()
    require(args.timeout > 0, "timeout must be positive")
    root = args.output.resolve() if args.output else Path(tempfile.mkdtemp(prefix="tractor-build-proof-"))
    if args.output:
        root.mkdir(parents=True, exist_ok=False)
    print(f"Proof artifacts: {root}", flush=True)
    receipt = {"passed": False, "source_revision": run("git", "rev-parse", "HEAD"),
               "source_sha256": source_hash(), "started_at": time.time()}
    save(root / "result.json", receipt)
    try:
        for name in ("claude", "codex", "agy"):
            require(shutil.which(name), f"native harness missing: {name}")
        binary = args.binary.resolve() if args.binary else root / "tractor"
        if not args.binary:
            subprocess.run(["go", "build", "-trimpath", "-o", str(binary), "./cmd/tractor"],
                           cwd=REPO, check=True)
        receipt.update(binary=str(binary), binary_sha256=sha(binary))
        (root / "binary-build.txt").write_text(run("go", "version", "-m", str(binary)) + "\n")
        workflow = run(str(binary), "workflows", "show", "sprint-execute") + "\n"
        require(workflow == (REPO / "internal/workflows/sprint-execute.yaml").read_text(),
                "candidate's embedded sprint-execute differs from this checkout")
        (root / "workflow.yaml").write_text(workflow)
        receipt["workflow_sha256"] = sha(root / "workflow.yaml")
        workspace = root / "workspace"
        shutil.copytree(FIXTURE, workspace, ignore=shutil.ignore_patterns("README.md", "__pycache__"))
        run("git", "init", "-q", cwd=workspace)
        run("git", "config", "user.name", "Tractor proof", cwd=workspace)
        run("git", "config", "user.email", "tractor-proof@example.invalid", cwd=workspace)
        run("git", "add", ".", cwd=workspace)
        run("git", "-c", "core.hooksPath=/dev/null", "commit", "-qm", "Seed broken shipping CLI", cwd=workspace)
        run("git", "tag", "seed", cwd=workspace)
        oracle = runpy.run_path(str(FIXTURE / "check.py"))["check"]
        baseline = {mode: oracle(workspace, mode) for mode in ("standard", "expedited")}
        save(root / "baseline-cli.json", baseline)
        require([r["passed"] for r in baseline["standard"]] == [True, False, True] and
                not any(r["passed"] for r in baseline["expedited"]),
                "seed did not demonstrate the expected real application failures")
        command = [str(binary), "run", "sprint-execute", "--workdir", str(workspace),
                   "--logs", str(root / "run")]
        receipt["argv"] = command
        save(root / "result.json", receipt)
        print("Running shipped sprint-execute with real native agents…", flush=True)
        with (root / "stdout.log").open("w") as stdout, (root / "stderr.log").open("w") as stderr:
            process = subprocess.Popen(command, cwd=workspace, stdout=stdout, stderr=stderr,
                                       start_new_session=True)
            try:
                code = process.wait(timeout=args.timeout)
            except (subprocess.TimeoutExpired, KeyboardInterrupt):
                os.killpg(process.pid, signal.SIGINT)
                try:
                    process.wait(timeout=15)
                except subprocess.TimeoutExpired:
                    os.killpg(process.pid, signal.SIGKILL)
                    process.wait()
                raise RuntimeError("proof interrupted or exceeded its whole-run timeout") from None
        receipt["exit_code"] = code
        require(code == 0, f"Tractor exited {code}; inspect {root / 'stderr.log'}")
        receipt["evidence"] = verify_run(root)
        require(sha(binary) == receipt["binary_sha256"], "candidate binary changed during proof")
        require(source_hash() == receipt["source_sha256"], "source changed during proof; rerun it")
        receipt["passed"] = True
    except Exception as error:
        receipt["error"] = str(error)
        print(f"NOT PROVED: {error}", file=sys.stderr)
    finally:
        receipt["finished_at"] = time.time()
        save(root / "result.json", receipt)
    print(f"{'PROVED' if receipt['passed'] else 'NOT PROVED'}: {root / 'result.json'}", flush=True)
    return 0 if receipt["passed"] else 1


if __name__ == "__main__":
    sys.exit(main())
