#!/usr/bin/env python3
"""Inspect resolved defaults across consumers through the actual CLI."""
import json
import pathlib
import subprocess
import sys

binary = str(pathlib.Path(sys.argv[1]).resolve())
reports = []


def inspect(label, graph, expected):
    run = subprocess.run([binary, "inspect-models", "--json", json.dumps(graph)],
                         capture_output=True, text=True, timeout=20)
    assert run.returncode == 0, (label, run.stderr)
    resolved = json.loads(run.stdout)
    lookup = {(r["node_id"], r["role"]): r for r in resolved}
    for key, values in expected.items():
        actual = lookup[key]
        for field, value in values.items():
            assert actual[field] == value, (label, key, field, actual[field], value)
        assert actual["source"], (label, key, "missing provenance")
    reports.append({"case": label, "resolved": resolved, "passed": True})


flash = {"native_model": "gemini-3.8-flash-medium", "effective_effort": "medium",
         "provider": "gemini", "harness": "agy"}
fable5 = {"native_model": "claude-fable-5", "effective_effort": "low",
          "provider": "anthropic", "harness": "claude", "authored_version": "5"}
sol = {"native_model": "gpt-5.6-sol", "effective_effort": "medium",
       "provider": "openai", "harness": "codex"}
inspect("ordinary-and-supervisor-atomic-replacement", {
    "start": "work", "defaults": {"model": {"name": "fable", "version": "5", "effort": "low"}},
    "nodes": [{"id": "work", "type": "agent", "model": {"name": "flash"},
               "edges": [{"to": "success"}]},
              {"id": "watch", "type": "supervisor", "prompt": "Observe.",
               "supervises": ["work"], "model": {"name": "flash"}}]},
    {("work", "agent"): flash, ("watch", "supervisor"): flash})
inspect("loop-role-defaults", {
    "start": "items", "defaults": {"model": {"name": "fable", "version": "5", "effort": "low"}},
    "nodes": [{"id": "items", "type": "loop", "checklist": "work.md",
               "edges": {"loop": "work", "exit": "success"}},
              {"id": "work", "type": "agent", "edges": [{"to": "items"}]}]},
    {("work", "agent"): fable5, ("items", "item_judge"): flash,
     ("items", "goal_evaluator"): fable5})
inspect("branch-template-and-override", {
    "start": "fan", "defaults": {"fidelity": "none", "model": {"name": "gpt-5.6-sol", "effort": "medium"}},
    "nodes": [{"id": "fan", "type": "fan_out", "model": {"name": "fable", "version": "5", "effort": "low"},
               "branches": [{"id": "left", "artifacts": ["REPORT.md"],
                             "agent": {"model": {"name": "flash"}}},
                            {"id": "right", "artifacts": ["REPORT.md"]}],
               "branch_edges": [{"to": "join"}]},
              {"id": "join", "type": "fan_in", "edges": [{"to": "success"}]}]},
    {("left", "branch_agent"): flash, ("right", "branch_agent"): fable5,
     ("fan", "branch_agent_template"): fable5, ("join", "fan_in"): sol})
print(json.dumps({"cases": reports, "passed": len(reports)}, indent=2))
