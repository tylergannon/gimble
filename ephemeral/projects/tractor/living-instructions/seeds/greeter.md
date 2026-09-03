# greeter

A command-line greeter, written in Go as one module at the root of this
repository. Expected to plan as MEDIUM.

- `greeter <name>` prints `Hello, <name>!`.
- `greeter --shout <name>` prints the same greeting in upper case.
- `greeter --names <file>` greets each non-empty line of the file, one
  greeting per line, in file order; `--shout` applies to each.
- No arguments: a usage message on stderr and exit status 2.
- `README.md` at the repository root says how to build and run it.

Acceptance examples. The finished program must produce exactly this:

```
$ greeter Ada
Hello, Ada!
$ greeter --shout Ada
HELLO, ADA!
$ printf 'Ada\n\nGrace\n' > names.txt && greeter --names names.txt
Hello, Ada!
Hello, Grace!
$ greeter; echo "exit=$?"
usage: greeter [--shout] [--names FILE] [NAME]
exit=2
```
