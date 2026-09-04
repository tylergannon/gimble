---
title: Graph editor
description: Open a pipeline file in the browser, edit nodes and edges with inline lint, and keep sharing the same YAML with the agent that is writing it.
eyebrow: Operator guide
order: 5
sourceLabel: Browse the editor source
sourceUrl: https://github.com/tylergannon/tractor/tree/main/web/editor
---

A pipeline is a YAML file, and the editor is a view of that file. `tractor
edit` serves a page that draws the graph, lets you change it, and writes the
result straight back to disk. Nothing is stored anywhere else.

## Start it

```sh
tractor edit examples/loops/bake-off.yaml
```

The command listens on loopback only, prints the URL, opens your browser, and
keeps running until you interrupt it. Pass `--addr 127.0.0.1:0` (the default)
to pick a port, or a fixed one such as `--addr 127.0.0.1:7331`; any address
that is not loopback is refused. `--no-open` prints the URL without opening a
browser, which is what you want over SSH.

The pipeline must end in `.yaml` or `.yml`. The page always writes YAML, so a
`.json` pipeline is rejected rather than silently rewritten.

## What the page does

The canvas shows every node and edge. Selecting a node opens an inspector for
its type: `agent`, `fan_out`, `fan_in`, `command`, `supervisor`, and `loop`
each get the fields the schema gives them, with inherited defaults shown
greyed. Selecting nothing shows the graph settings and file-level `defaults`.

Lint is the same validator `tractor validate` runs. Its diagnostics appear on
the node they belong to, and graph-wide problems appear beside the settings,
so what the page accepts is what `start_run` will accept.

Each change is saved about half a second after you make it. Edits go back into
the parsed document at the changed path, so comments, key order, and the
wrapping of prose you did not touch all survive. The header shows when a save
is in flight.

## Sharing the file with an agent

The server watches the file. When anything else writes it (an agent working
in the repository, another editor, `git checkout`), the page reloads and shows
the new graph. An agent can keep editing a pipeline while a person has it open
and both see one file.

If the file changes on disk while the page has edits that have not been saved
yet, the page keeps the on-disk version, drops its own, and says so in a
banner. The disk always wins; the window is short because saves are quick.

A YAML syntax error on disk puts the page in read-only mode. The canvas shows
the error and its line, every field is disabled, and the page returns to
normal on the next on-disk change that parses.

## Where positions live

Node positions are not part of the pipeline. They are stored beside it in
`<stem>.layout.json`, so `bake-off.yaml` gets `bake-off.layout.json`. A
pipeline with no sidecar is laid out automatically: columns by distance from
`start`, supervisors in a row beneath, terminals on the right. The sidecar is
written the first time you move or resize a node; commit it or ignore it.

## Rebuilding the page

The page is a SvelteKit app in `web/editor`, compiled into the binary from
`internal/editor/dist`. After changing anything under `web/editor/src`, run:

```sh
pnpm --dir web/editor build
```

and rebuild `tractor`. The bundle uses a constant version string instead of a
build timestamp, so rebuilding unchanged source produces byte-identical files
and the committed `dist` shows no spurious diff.
