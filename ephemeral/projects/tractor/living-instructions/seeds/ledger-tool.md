# ledger

A command-line tool, written in Go as one module at the root of this
repository, that keeps a markdown ledger of items in `LEDGER.md` in the
current directory. Expected to plan as LARGE with two chapters: the
ledger commands, then validation and export.

- `ledger add <text>` appends `- [ ] <text>` to `LEDGER.md`, creating
  the file if needed.
- `ledger list` prints every item numbered from 1 in file order, open
  items as `[ ]` and done items as `[x]`.
- `ledger done <n>` marks item n done; an out-of-range n is exit status
  1 with a message on stderr.
- `ledger validate` exits 0 when every non-empty line of `LEDGER.md` is
  an item in the form above, and exits 1 naming the first bad line
  otherwise.
- `ledger export --json` prints the items as a JSON array of objects
  with `text` and `done` fields, in file order.
- `README.md` at the repository root says how to build and run it.

Acceptance examples. The finished program must produce exactly this,
starting in an empty directory, in this order:

```
$ ledger add "buy milk" && ledger add "call Ada" && ledger list
1. [ ] buy milk
2. [ ] call Ada
$ ledger done 1 && ledger list
1. [x] buy milk
2. [ ] call Ada
$ ledger done 9; echo "exit=$?"
ledger: no item 9
exit=1
$ ledger validate; echo "exit=$?"
exit=0
$ ledger export --json
[{"text":"buy milk","done":true},{"text":"call Ada","done":false}]
$ echo "garbage" >> LEDGER.md && ledger validate; echo "exit=$?"
LEDGER.md:3: not an item: garbage
exit=1
```
