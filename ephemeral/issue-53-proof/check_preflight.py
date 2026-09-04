#!/usr/bin/env python3
"""Exercise real CLI preflight; trap any native launch for invalid input."""
import copy
import json
import os
import pathlib
import subprocess
import sys
import tempfile

binary = str(pathlib.Path(sys.argv[1]).resolve())
root = pathlib.Path(tempfile.mkdtemp(prefix="tractor-issue53-preflight-"))
traps = root / "bin"
traps.mkdir()
marker = root / "unexpected-native-launch"
for name in ("claude", "agy", "codex"):
    trap = traps / name
    trap.write_text('#!/bin/sh\nprintf "%s\\n" "$0" >> "$ISSUE53_TRAP_MARKER"\nexit 97\n')
    trap.chmod(0o755)
environment = dict(os.environ, PATH=str(traps) + os.pathsep + os.environ["PATH"],
                   ISSUE53_TRAP_MARKER=str(marker))
base = {"start": "work", "defaults": {"model": {"name": "fable", "version": "5", "effort": "low"}},
        "nodes": [{"id": "work", "type": "agent", "prompt": "No real turn should run in this test.",
                   "edges": [{"to": "success"}]}]}
cases = []


def selection_case(label, selection, valid):
    graph = copy.deepcopy(base)
    graph["nodes"][0]["model"] = selection
    cases.append((label, graph, valid))


cases.append(("inherited-whole-selection", copy.deepcopy(base), True))
selection_case("name-only-flash", {"name": "flash"}, True)
selection_case("fable-version-5", {"name": "fable", "version": "5"}, True)
selection_case("fable-version-5.1", {"name": "fable", "version": "5.1"}, True)
selection_case("native-model", {"name": "gpt-5.6-sol", "effort": "medium"}, True)
selection_case("native-fixed-effort", {"name": "gemini-3.8-flash-medium"}, True)
selection_case("unknown-provider", {"name": "unrecognized-proof-model"}, False)
selection_case("authored-provider", {"name": "fable", "provider": "anthropic"}, False)
selection_case("numeric-version", {"name": "fable", "version": 5.1}, False)
selection_case("unknown-version", {"name": "fable", "version": "999999"}, False)
selection_case("effort-only", {"effort": "low"}, False)
selection_case("version-only", {"version": "5.1"}, False)
selection_case("empty-model", {}, False)
selection_case("blank-name", {"name": " "}, False)
selection_case("empty-version", {"name": "fable", "version": ""}, False)
selection_case("empty-effort", {"name": "fable", "effort": ""}, False)
selection_case("native-plus-version", {"name": "gpt-5.6-sol", "version": "5.6"}, False)
selection_case("pinned-alias-plus-version", {"name": "fable-5.1", "version": "5.1"}, False)
selection_case("native-effort-conflict", {"name": "gemini-3.8-flash-medium", "effort": "high"}, False)
selection_case("string-shorthand", "flash", False)
selection_case("null-model", None, False)
for field, value in (("llm_model", "flash"), ("llm_provider", "gemini"), ("reasoning_effort", "low")):
    graph = copy.deepcopy(base)
    graph["nodes"][0][field] = value
    cases.append(("legacy-agent-" + field, graph, False))

for role in ("item_judge", "goal_evaluator"):
    graph = {"start": "items", "nodes": [
        {"id": "items", "type": "loop", "checklist": "work.md",
         role: {"model": {"name": "unrecognized-proof-model"}},
         "edges": {"loop": "work", "exit": "success"}},
        {"id": "work", "type": "agent", "edges": [{"to": "items"}]}]}
    cases.append(("hidden-role-" + role, graph, False))
graph = copy.deepcopy(base)
graph["nodes"].append({"id": "watch", "type": "supervisor", "prompt": "Observe.",
                       "supervises": ["work"], "model": {"name": "unrecognized-proof-model"}})
cases.append(("supervisor-preflight", graph, False))
graph = {"start": "fan", "defaults": {"fidelity": "none"}, "nodes": [
    {"id": "fan", "type": "fan_out", "branches": [
        {"id": "left", "artifacts": ["REPORT.md"],
         "agent": {"model": {"name": "unrecognized-proof-model"}}},
        {"id": "right", "artifacts": ["REPORT.md"]}],
     "branch_edges": [{"to": "join"}]},
    {"id": "join", "type": "fan_in", "prompt": "Compare.",
     "edges": [{"to": "success"}]}]}
cases.append(("synthesized-branch-preflight", graph, False))
graph = copy.deepcopy(base)
graph["defaults"]["model"] = {"name": "unrecognized-proof-model"}
graph["nodes"][0]["model"] = {"name": "fable"}
cases.append(("invalid-overridden-default", graph, False))

report = []
for number, (label, graph, valid) in enumerate(cases):
    serialized = json.dumps(graph)
    validation = subprocess.run([binary, "validate", "--json", serialized],
                                env=environment, cwd=root, text=True, capture_output=True, timeout=20)
    assert (validation.returncode == 0) == valid, (label, validation.stdout, validation.stderr)
    record = {"case": label, "expected_valid": valid, "validate_exit": validation.returncode,
              "stdout": validation.stdout, "stderr": validation.stderr}
    if not valid:
        logs = root / f"logs-{number}"
        execution = subprocess.run([binary, "run", "--json", serialized,
                                    "--workdir", str(root), "--logs", str(logs)],
                                   env=environment, cwd=root, text=True, capture_output=True, timeout=20)
        assert execution.returncode != 0, label
        assert not marker.exists(), (label, "native harness was launched")
        assert not logs.exists(), (label, "run logs were created before preflight rejection")
        record.update(run_exit=execution.returncode, run_stderr=execution.stderr,
                      native_launch=False, logs_created=False)
    report.append(record)
print(json.dumps({"cases": report, "passed": len(report), "temporary_root": str(root)}, indent=2))
