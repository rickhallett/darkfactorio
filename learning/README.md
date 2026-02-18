# Learning Record

This directory is the append-only operating memory for `darkfactorio`.

It is intentionally project-agnostic: entries can reference any external system, repo, run window, or incident.

## Structure

- `learning/journal/YYYY/YYYY-MM-DD.md`: timestamped event log entries.
- `learning/decisions/`: explicit ADR-style records when a decision needs stand-alone traceability.

## Gate Rule

Substantive repo changes must include at least one update under:

- `learning/journal/`
- `learning/decisions/`

If no substantive files changed, the gate passes without requiring a new entry.

Current substantive filter:

- Includes most files.
- Ignores `.gitignore` only.

## CLI

Add an entry:

```bash
go run ./cmd/dflearn touch \
  --source-project tspit \
  --source-ref "run:w-2026-02-l4-01" \
  --summary "Ran level4 gate against baseline profile" \
  --decision "Keep current scenario envelope until adversarial replay fails" \
  --evidence "runs/w-2026-02-l4-01.ndjson" \
  --next-action "Replay against adversarial profile"
```

Check gate:

```bash
go run ./cmd/dflearn check --base origin/main --head HEAD
```

## Workflow

1. Do real work.
2. Record what changed, why, and what happens next using `dflearn touch`.
3. Run `dflearn check` before push/PR.
4. CI enforces the same gate.

If this feels strict, good. That is the point while the POC is hot.
