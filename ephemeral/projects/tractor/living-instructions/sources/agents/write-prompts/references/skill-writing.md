# Skill Writing

Use this reference when the output is a new or edited agent skill.

## Design Rules

1. Scope the skill to one repeatable job.
2. Make the frontmatter description front-loaded and trigger-specific.
3. Keep `SKILL.md` short enough to read in one pass.
4. Put details in one-level-deep `references/` files.
5. Add scripts only for deterministic or repeatedly generated work.
6. Include a definition of done and closeout evidence.
7. Do not add auxiliary docs such as `README.md` inside the skill directory
   unless the host format requires them.

## Description Shape

Use this structure:

```md
---
name: short-hyphen-name
description: >
  Do one specific capability. Use when the user says X, asks for Y, edits Z,
  or needs this workflow. Do not use when the task is unrelated.
---
```

The description is the routing surface. It should include the words users and
agents will actually say.

## Body Shape

Prefer this structure:

```md
# Skill Name

One paragraph: what job this skill performs.

## Workflow

1. Orient.
2. Do the work.
3. Verify.
4. Report.

## References

- Read [x.md](references/x.md) when ...

## Closeout

Report the exact evidence.
```

## Token Discipline

Good skills respect the intelligence of the receiving model. Avoid:

- restating generic coding advice;
- explaining concepts the model already knows;
- huge lists of prohibitions;
- project-specific lore unless the skill is project-specific;
- mixing examples, policy, scripts, and troubleshooting into one long file.

## Validation Ideas

- Run the repo's skill lister and confirm the new skill appears.
- Ask a fresh agent to use the skill on a small representative task.
- Add a focused test if the skill ships scripts or linker metadata.
- Keep early versions local and iterate after real use.
