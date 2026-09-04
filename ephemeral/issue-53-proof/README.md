# Issue 53 independent proof

Prepared by the parent reviewer, separately from implementation ownership.
The fixture is a real Python shipping-quote CLI with a deliberate inclusive
threshold bug and missing expedited mode. A real worker first captures the
faulty application output without repairing it. The item judge must reject
that evidence, a later worker turn must repair the threshold, and the goal
evaluator must dispatch missing expedited work despite the first item passing.

`capture.py` records actual commands, exit statuses, and outputs. It supplies no
pass/fail opinion. `snapshot` preserves each worker's software and evidence
before loop validation; `verify_quote.py` is a separate final behavioral oracle
outside the worker's disposable workspace. The fixture demonstrates these
specific decisions and routing behavior, not model judgment quality in general.

Copy `fixture/` into a disposable directory before executing the candidate
Tractor CLI. Do not run a harness in this source/proof directory. Runtime results,
native selections, tested revision, and editor proof will be recorded after
execution. This document currently describes the proof design, not a pass.
