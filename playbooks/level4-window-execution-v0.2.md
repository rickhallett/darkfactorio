# level 4 window execution v0.2

## Purpose

Run a repeatable two-window evidence sprint that proves whether `darkfactorio` can sustain Level 4 behavior under hard gates.

This is intentionally boring. Boring is how reliability happens.

## Scope

- Duration: 14 days
- Windows: 2
- Runs per window: 10 minimum
- Required class mix per window:
  - `low_risk_feature >= 4`
  - `medium_integration >= 4`

## Fixed Inputs (lock on day 1)

- Baseline criteria: `profiles/level4-gate-v0.1-baseline.json`
- Adversarial criteria: `profiles/level4-gate-v0.1-adversarial.json`
- Record schema: `schemas/level4-eval-record-v0.1.json`
- Evaluator: `cmd/dfgatev01`
- Learning gate: `cmd/dflearn`

No threshold/profile/schema edits inside an active window.

## Roles

- Window owner: accountable for go/no-go
- Operator: executes runs + appends records
- Spec owner: owns input specification quality
- Scenario owner: owns external behavioral scenarios
- Incident owner: owns any approved-run incident follow-up

One person can hold multiple roles for small teams.

## Artifacts per window

- Run data: `runs/<window_id>.ndjson`
- Baseline gate output: captured in learning journal entry
- Adversarial gate output: captured in learning journal entry
- Closeout summary: blockers + corrective actions + next review date

## Day-by-day runbook

## Day 1: Lock and stage

1. Choose two window IDs:
   - `w-YYYY-MM-l4-02`
   - `w-YYYY-MM-l4-03`
2. Create empty files:
   - `runs/w-YYYY-MM-l4-02.ndjson`
   - `runs/w-YYYY-MM-l4-03.ndjson`
3. Confirm tools:
   - `make test`
4. Add learning entry:
   - `go run ./cmd/dflearn touch --source-project darkfactorio --summary "Locked v0.2 execution sprint inputs" --decision "Freeze schema/profiles for active windows"`

Checkpoint:

- Inputs locked, files staged, tests green.

## Days 2-6: Execute window 1 (10 runs)

For each run:

1. Append one valid JSON line to `runs/w-YYYY-MM-l4-02.ndjson`.
2. Run baseline gate:

```bash
go run ./cmd/dfgatev01 \
  -input runs/w-YYYY-MM-l4-02.ndjson \
  -window w-YYYY-MM-l4-02 \
  -criteria profiles/level4-gate-v0.1-baseline.json \
  -output text
```

3. If malformed line error occurs, fix immediately before next run.
4. If intervention occurred, add `dflearn touch` entry with cause and action.

Checkpoint (end of day 6):

- 10 valid runs captured
- class mix satisfied
- no unresolved schema errors

## Day 7: Close window 1

1. Run baseline gate and capture result.
2. Run adversarial replay and capture result:

```bash
go run ./cmd/dfgatev01 \
  -input runs/w-YYYY-MM-l4-02.ndjson \
  -window w-YYYY-MM-l4-02 \
  -criteria profiles/level4-gate-v0.1-adversarial.json \
  -output text
```

3. Write closeout learning entry:
   - pass/fail both gates
   - top 3 blockers
   - corrective actions with owner + due date

Checkpoint:

- both gate outputs recorded
- closeout exists

## Days 8-13: Execute window 2 (10 runs)

Repeat Days 2-6 using `w-YYYY-MM-l4-03`, applying only corrective actions from window 1.

No new process changes allowed unless incident-level justified.

Checkpoint (end of day 13):

- second 10-run window complete
- both classes satisfy minimums

## Day 14: Promotion decision

Run both gates for window 2 and make decision using rules below.

## Promotion Rules (v0.2 -> v0.3)

Promote only if true across both windows:

1. Baseline gate pass in both windows.
2. Adversarial gate pass in both windows.
3. Approved-run critical incidents = 0 in both windows.
4. Intervention trend stable/decreasing in both windows.
5. Blockers from window 1 have explicit disposition (closed, accepted risk, or deferred with owner/date).

If any condition fails: no promotion. Roll into remediation window.

## Command set

Baseline gate:

```bash
go run ./cmd/dfgatev01 -input runs/<window_id>.ndjson -window <window_id> -criteria profiles/level4-gate-v0.1-baseline.json -output text
```

Adversarial gate:

```bash
go run ./cmd/dfgatev01 -input runs/<window_id>.ndjson -window <window_id> -criteria profiles/level4-gate-v0.1-adversarial.json -output text
```

Corpus adversarial replay (multi-window):

```bash
go run ./cmd/dfcorpusv01 \
  --inputs runs/w-2026-02-l4-02.ndjson,runs/w-2026-02-l4-03.ndjson \
  --criteria profiles/level4-gate-v0.1-adversarial.json \
  --output text
```

Optional CI variant:

- GitHub Actions workflow: `corpus-promotion-check` (manual dispatch)

Learning entry:

```bash
go run ./cmd/dflearn touch --source-project darkfactorio --source-ref "window:<window_id>" --summary "<what changed>" --decision "<decision>" --evidence "<path>" --next-action "<next action>"
```

Learning gate check:

```bash
go run ./cmd/dflearn check --base HEAD~1 --head HEAD
```

Autonomous batch advance (append runs + replay gates + learning log):

```bash
go run ./cmd/dfwindowv01 --window <window_id> --append 2
```

Quality remediation mode (forces perfect scenario outcomes in appended runs):

```bash
go run ./cmd/dfwindowv01 --window <window_id> --append 2 --quality high --quality-reason "<explicit justification>"
```

## Failure protocol

If gate fails mid-window:

1. Do not alter thresholds.
2. Record failure reason in learning log.
3. Apply one corrective action at a time.
4. Re-run and measure impact.

If ingestion fails:

1. Fix malformed record at the source line.
2. Re-run gate.
3. Record root cause once per failure class (not per typo).

## Anti-patterns (disallowed)

- “Passing run but no record written”
- “Adversarial replay skipped because baseline passed”
- “Profile tweaked to rescue a failing window”
- “Diff review used as acceptance substitute”

## Exit artifact template (copy per window)

- Window ID:
- Date range:
- Baseline gate result:
- Adversarial gate result:
- Run count and class mix:
- Top blockers:
- Corrective actions:
- Promotion recommendation:
