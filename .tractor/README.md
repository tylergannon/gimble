# `.tractor/`

## `run`

How this repository starts the software under test. One shell command; its
contents go to `/bin/sh`, the same as any command node.

```sh
# .tractor/run
exec go run ./cmd/server
```

The engine allocates a free port, exports `PORT` and `TRACTOR_URL`, starts
this in its own process group, and waits for the port to accept a connection
before validating anything. Every validation command in the lap runs against
that one instance, and the group is signalled away afterwards.

Tractor holds one handle and never looks inside it: whether that command is a
single binary, `overmind start`, or `docker compose up` is not something the
engine can tell. Anything a process supervisor does belongs in there, not
here. A tool with its own spelling for the port is adapted in this file —

```sh
OVERMIND_PORT=$PORT exec overmind start
```

There is no `ready:`, `env:`, `depends_on:`, or `port:`. The port accepting is
the readiness probe, and a port you chose is a port something else can already
be sitting on.

A repository without this file validates exactly as it always has.
