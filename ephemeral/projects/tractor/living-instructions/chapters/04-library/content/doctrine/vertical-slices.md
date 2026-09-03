# Vertical slices

Every slice crosses the stack thinly and yields something exercisable
when it is done: a command that runs, a page that renders, a promise
you can demonstrate. Reject horizontal plans (all the storage, then all
the logic, then all the interface). That is the default model tic, and
it produces sprints that finish with nothing to show.

Sequence slices so each demonstrates at least one promise or de-risks
a seam a promise crosses. If you cannot say what is exercisable at the
end of a slice, it is not a slice yet; cut it differently.

Sizing:

- A sprint is one agent turn: one fresh context, one commit, one
  demonstration. If it needs two turns, it is two sprints.
- A chapter is a run of sprints toward one exercisable state, proven
  at exit by a verifier who operates the result.

Every slice serves a promise, and every promise reaches a slice. Route
them explicitly: a slice that serves no promise is speculative scope;
a promise no slice serves is unbuilt. Both are defects a reviewer will
name.

Docs, cleanup, and closeout are not slices on their own. Fold them
into the slice whose result they document, and give that slice a
demonstration (an agent follows the docs and does the thing).

Source: sources/tractor-design/wisdom.md (rule 4, from slice-design); decisions.md 35, 44.
