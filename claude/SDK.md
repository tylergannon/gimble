# Upstream revision

Gimble depends on the forked module:

```text
github.com/tylergannon/claude-agent-sdk-go
v1.1.1-0.20260912021749-9a4ffeca77cc
revision 9a4ffeca77cca476e5f9e987c4baec71b3d4d2ce
```

The fork starts from `github.com/roasbeef/claude-agent-sdk-go` revision
`efdbecd88a98`. It retains the upstream license and removes the upstream
repository's tracked `.cache/` and `scratch/` artifacts from the published
module.

Local change:

- `Options.RawMessageObserver` and `WithRawMessageObserver` synchronously expose
  an owned copy of each complete, non-empty stdout JSON message immediately
  before `ParseMessage`.

The fork source patch is 163 added lines across `options.go`, `transport.go`,
and `transport_test.go` (17, 31, and 115 lines respectively).
