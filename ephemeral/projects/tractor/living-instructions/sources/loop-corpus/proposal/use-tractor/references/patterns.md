# Small Tractor pipeline patterns

These examples use Tractor's native JSON graph rather than upstream DOT. Replace
placeholder verification commands with the repository's real commands. Keep the
visit budgets unless the task justifies a different explicit bound.

## Contents

- [Implement and check](#implement-and-check)
- [Plan, implement, and check](#plan-implement-and-check)
- [Diagnose, fix, and reproduce](#diagnose-fix-and-reproduce)
- [Draft, review, and revise](#draft-review-and-revise)
- [Authoring notes](#authoring-notes)

## Implement and check

Use this by default for a change with a trustworthy automated check. The worker
owns the whole coherent task; the graph owns the exit evidence and retry bound.

```json
{
  "name": "implement-check",
  "goal": "Implement the requested change and prove it with the repository's focused test command",
  "defaults": {
    "timeout": "15m"
  },
  "start": "implement",
  "nodes": [
    {
      "id": "implement",
      "type": "codergen",
      "label": "Implement or correct",
      "prompt": "Advance $goal. Read .tractor-check-output.txt if it exists and address the reported failure. Keep the change focused; do not weaken the check. When the implementation is ready, route to verification.",
      "max_visits": 4,
      "edges": [
        { "to": "verify" }
      ]
    },
    {
      "id": "verify",
      "type": "tool",
      "label": "Focused verification",
      "tool_command": "./run_tests.sh > .tractor-check-output.txt 2>&1",
      "on_success": "success",
      "on_error": "implement",
      "timeout": "10m"
    }
  ]
}
```

## Plan, implement, and check

Use this when the approach spans meaningful boundaries or deserves reviewable
state. The plan is a workspace artifact. Red tests normally return to
implementation; the implementer chooses replanning only when the approach has
proved wrong.

```json
{
  "name": "plan-implement-check",
  "goal": "Design, implement, and verify the requested repository change",
  "defaults": {
    "timeout": "15m"
  },
  "start": "plan",
  "nodes": [
    {
      "id": "plan",
      "type": "codergen",
      "label": "Plan",
      "prompt": "Plan $goal against the current repository. Write .tractor-plan.md containing the chosen approach, files and integration points, risks, and exact verification. Keep it concise enough for another agent to execute.",
      "max_visits": 2,
      "edges": [
        { "to": "implement" }
      ]
    },
    {
      "id": "implement",
      "type": "codergen",
      "label": "Implement",
      "prompt": "Read .tractor-plan.md and implement $goal. If .tractor-check-output.txt exists, use it as correction evidence. Choose verification when the approach remains sound, replanning only when repository facts invalidate the plan, or failure when the goal cannot be completed as stated.",
      "max_visits": 4,
      "edges": [
        { "to": "verify", "condition": "The implementation is ready for the planned verification" },
        { "to": "plan", "condition": "Repository evidence invalidated the approach and the plan must change" },
        { "to": "failure", "condition": "The goal is impossible or unsafe as stated" }
      ]
    },
    {
      "id": "verify",
      "type": "tool",
      "label": "Verify",
      "tool_command": "./run_tests.sh > .tractor-check-output.txt 2>&1",
      "on_success": "success",
      "on_error": "implement",
      "timeout": "10m"
    }
  ]
}
```

## Diagnose, fix, and reproduce

Use this for a bug whose symptom can be reproduced mechanically. Diagnosis
writes the current theory; the check refreshes the evidence on every lap.

```json
{
  "name": "diagnose-fix-reproduce",
  "goal": "Find the root cause of the reported bug, implement the minimal fix, and prove the original symptom is gone",
  "defaults": {
    "timeout": "15m"
  },
  "start": "diagnose",
  "nodes": [
    {
      "id": "diagnose",
      "type": "codergen",
      "label": "Diagnose",
      "prompt": "Investigate $goal. Read .tractor-reproduction.txt if present. Write .tractor-diagnosis.md with the observed symptom, relevant evidence, current root-cause hypothesis, and the next discriminating change or check. Do not patch blindly.",
      "max_visits": 3,
      "edges": [
        { "to": "fix" }
      ]
    },
    {
      "id": "fix",
      "type": "codergen",
      "label": "Fix",
      "prompt": "Read .tractor-diagnosis.md and implement the smallest fix for $goal. Add or preserve a regression check for the original symptom. When ready, route to reproduction.",
      "max_visits": 3,
      "edges": [
        { "to": "reproduce" }
      ]
    },
    {
      "id": "reproduce",
      "type": "tool",
      "label": "Reproduce and test",
      "tool_command": "./reproduce_bug.sh > .tractor-reproduction.txt 2>&1",
      "on_success": "success",
      "on_error": "diagnose",
      "timeout": "10m"
    }
  ]
}
```

## Draft, review, and revise

Use this for a document, design, plan, or code change whose acceptance includes
judgment. The reviewer starts fresh on each visit and writes findings for the
reviser. Give the reviewer a more specific mandate than this generic template
whenever possible.

```json
{
  "name": "draft-review-revise",
  "goal": "Produce an artifact that fully satisfies the requested requirements",
  "defaults": {
    "timeout": "15m"
  },
  "start": "draft",
  "nodes": [
    {
      "id": "draft",
      "type": "codergen",
      "label": "Draft",
      "prompt": "Create the artifact required by $goal in the workspace. Make the result self-contained and ready for independent review.",
      "edges": [
        { "to": "review" }
      ]
    },
    {
      "id": "review",
      "type": "codergen",
      "label": "Independent requirements review",
      "prompt": "Independently inspect the current workspace artifact against $goal. Write .tractor-review.md with only material, actionable findings and their evidence. Choose success when there are no material findings, revision when specific changes are needed, or failure when the requirements are internally impossible.",
      "fidelity": "none",
      "max_visits": 4,
      "edges": [
        { "to": "success", "condition": "No material findings remain" },
        { "to": "revise", "condition": "Material findings can be corrected" },
        { "to": "failure", "condition": "The requirements cannot be satisfied as stated" }
      ]
    },
    {
      "id": "revise",
      "type": "codergen",
      "label": "Revise",
      "prompt": "Read .tractor-review.md, verify each finding against the artifact and $goal, and correct every supported material issue. Preserve correct work and keep the revision focused.",
      "max_visits": 3,
      "edges": [
        { "to": "review" }
      ]
    }
  ]
}
```

## Authoring notes

- Call `get_pipeline_schema` before authoring; these examples are guidance, not
  a substitute for the installed Tractor version's schema.
- Replace `./run_tests.sh` or `./reproduce_bug.sh` with a command proven to work
  from the run's work directory.
- Plain redirection preserves the command's exit status while making failures
  visible to the next agent turn.
- `max_retries` handles retryable execution failures such as provider errors;
  `max_visits` bounds graph-level corrective laps.
- Add parallel fan-out only after identifying two or more independent branches
and a concrete fan-in comparison or synthesis task.
