# Promotion Decision: v0.2 -> v0.3

## Context

Two-window execution cycle completed with autonomous tooling and corpus-level adversarial replay.

Evidence sources:

- `runs/w-2026-02-l4-02.ndjson`
- `runs/w-2026-02-l4-03.ndjson`
- `profiles/level4-gate-v0.1-baseline.json`
- `profiles/level4-gate-v0.1-adversarial.json`

## Gate Results

- Window `w-2026-02-l4-02` baseline: pass
- Window `w-2026-02-l4-03` baseline: pass
- Corpus adversarial (combined windows): pass

Corpus metrics snapshot:

- records: 32
- class mix: low_risk_feature=16, medium_integration=16
- scenario pass rate: 95.18%
- first-pass rate: 100%
- mean retries: 1.00
- decision reversal rate: 0%
- approved critical incidents: 0

## Decision

Promote to `v0.3` operating state.

Rationale:

- Baseline stability proven across both windows.
- Adversarial criteria passed at corpus level.
- No incident/reversal regressions.

## Guardrails for `quality=high` Mode

`quality=high` is allowed only when all are true:

1. The sole failing metric is scenario pass rate.
2. Run count/class minima are already met or actively targeted in same cycle.
3. A learning entry records why high mode was used.

`quality=high` is disallowed when:

1. Critical incidents exist in approved runs.
2. Decision reversal rate is non-zero and worsening.
3. It is being used to mask class/run-count deficits.

## Next Actions

1. Add policy enforcement in `dfwindowv01` to require explicit `--quality high` justification text.
2. Add corpus replay to CI as optional promotion check job.
3. Start first `v0.3` window with standard mode default and quality-high only by exception.
