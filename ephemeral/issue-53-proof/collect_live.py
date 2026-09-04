#!/usr/bin/env python3
"""Collect a bounded, prompt-free receipt from a completed native proof run."""
import json
import pathlib
import sys

root = pathlib.Path(sys.argv[1])


def lines(path):
    return [json.loads(line) for line in path.read_text().splitlines() if line]


metadata = json.loads((root / "metadata.json").read_text())
assert metadata["exit_code"] == metadata["oracle_exit_code"] == 0, metadata
timeline = lines(root / "logs/timeline.jsonl")
validations = [e for e in timeline if e["type"] == "LoopValidated"]
evaluations = [e for e in timeline if e["type"] == "LoopEvaluated"]
assert [e["passed"] for e in validations] == [False, True, True], validations
assert [e["verdict"] for e in evaluations] == ["not_done", "done"], evaluations
assert timeline[-1]["type"] == "PipelineCompleted", timeline[-1]
assert len(validations[-1]["validations"]) == 2
native = lines(root / "native-selections.jsonl")
logged = sorted([event for path in (root / "logs/events").glob("*.jsonl")
                 for event in lines(path) if event.get("type") == "model_selection"],
                key=lambda event: event["ts"])
expected = {
    "agent": ("claude", "claude-fable-5-1", "high"),
    "item_judge": ("agy", "gemini-3.8-flash-medium", "medium"),
    "goal_evaluator": ("codex", "gpt-5.6-sol", "medium"),
}
assert {e["role"] for e in logged} == set(expected)
for event in logged:
    route, model, effort = expected[event["role"]]
    assert (event["harness"], event["native_model"], event["effective_effort"]) == (route, model, effort)
    matching = [e for e in native if e["harness"] == route and e["selection"].get("model") == model]
    if route != "agy":
        matching = [e for e in matching if e["selection"].get("effort") == effort]
    assert matching, (event, "No matching actual native request")

workspace = root / "workspace"
snapshots = {path.name: path.read_text() for path in sorted((workspace / "snapshots").iterdir())}
assert len(snapshots) == 6, snapshots.keys()
oracle = lines(root / "final-oracle.jsonl")
assert len(oracle) == 6 and all(e["passed"] for e in oracle)
report = {
    "metadata": metadata,
    "observed_sequence": "judge fail; repair; judge pass; evaluator not_done; appended work dispatched; both items pass; evaluator done; pipeline completed",
    "timeline": [e for e in timeline if e["type"].startswith("Loop") or e["type"] == "PipelineCompleted"],
    "native_requests": native,
    "logged_model_selections": logged,
    "snapshots": snapshots,
    "final_program": (workspace / "quote.py").read_text(),
    "final_checklist": (workspace / "work.md").read_text(),
    "independent_final_oracle": oracle,
    "passed": True,
}
print(json.dumps(report, indent=2))
