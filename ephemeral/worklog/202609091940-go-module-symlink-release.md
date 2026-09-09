# Go module archive repair

friction: Go module archives omit symbolic links. Gimble's generated editor package imported three symlinked Go files, so v0.10.0 could not be installed from the public module proxy despite working from a checkout.
decision: Materialize only the three generated Go link files as regular source files and publish v0.10.1; verify using an isolated go install of the public tag.
