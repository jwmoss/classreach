# AGENTS.md

## Purpose

`classreach` is a Go CLI wrapper around the
ClassReach REST API. Keep provider-specific behavior in
commands and typed API packages; keep generic transport, config, and output
helpers small and reusable.

## CLI Rules

- Primary command output goes to stdout.
- Errors, traces, and diagnostics go to stderr.
- Keep `--json` stable for scripts.
- Keep `--plain` stable and line-oriented where implemented.
- Do not print token values.
- Do not add token/password command-line flags unless there is a documented
  reason and a safer path is still available.
- Mutating commands must respect `--dry-run`.
- Commands that need credentials should explain the missing env/config value in
  the error.

## Validation

Run before handoff:

```bash
make check
```

For release config changes, also run if GoReleaser is installed:

```bash
make release-check
```

## Test policy

Keep only live read-only end-to-end tests and static checks.
Use the compiled CLI with the real provider in `tests/` and the `live` build tag.
Do not add unit tests, mocked transports, or fake-server subprocess tests.
Run `make check` for static validation and builds without credentials.
Run `make test` for live read-only checks with existing configuration.
Treat missing access as a failure, not a successful skip.
Never log credentials or provider response bodies from live checks.
