1. Yes. During a real `implement` turn, replace each open sprint item’s contents and check while preserving its name, then implement only the known acceptance examples. The engine reloads and validates the current ledger after the turn (`engine/loop.go:68-96`), while P10 compares only A0 names and disappearance/renaming (`design.md:64-72`). The infer judge never compares approved and executed item contents (`design.md:78-90`). All events remain genuine, but the approved package was not what ran.

2. Yes. The claim “the approved package was what ran” depends on package identity the design does not capture. It byte-compares the pre-approval and final planning packages (`design.md:57-59`), but execution binding is reduced to matching item names. Same-name changes to commands, checklist text, or sprint documents therefore cannot distinguish faithful execution from replacement.

3. Yes. `validate-plan` requires the canonical `Next` value to be exactly `tractor workflow run medium|large --project <project>` (`workflow/artifacts.go:248-262`). The design simultaneously requires executing that line verbatim and adding a check-named `--logs <dir>` (`design.md:23-25,60-61`). A correct package with the required canonical handoff must fail one of those mutually incompatible checks.

ROUTE: fail
