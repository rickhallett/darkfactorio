# level 4 run 01-10 execution sheet v0.1

## Intent

Operate a minimum 10-run Level 4 window with deterministic metrics and machine-evaluated go/no-go.

## Fixed Inputs (lock before run 01)

- Window ID: `<window_id>`
- Factory profile: `<factory_profile_id>`
- Pipeline class A: `low_risk_feature`
- Pipeline class B: `medium_integration`
- Scenario suite versions pinned per pipeline class
- Evaluation command (baseline): `go run ./cmd/dfgatev01 -input runs/<window_id>.ndjson -window <window_id> -criteria profiles/level4-gate-v0.1-baseline.json -output json`

## Ownership

- Window owner:
- Operator primary:
- Operator backup:
- Spec owner:
- Scenario owner:
- Incident owner:

## Run Cadence and Mix

- Run count minimum: 10
- Required mix: at least 4 runs in each class, remaining 2 flexible
- Suggested cadence: 2 runs/day for 5 days

## Execution Table

| Run # | Planned Date | Pipeline ID | Pipeline Class | Pipeline Version | Scenario Suite Version | Operator | Outcome (A/R/F) | First Pass (Y/N) | Retries | Interventions | Scenario Passed/Total | Decision Reversed (Y/N) | Critical Incident (Y/N) | Run ID |
|---|---|---|---|---|---|---|---|---|---:|---:|---|---|---|---|
| 01 |  |  | low_risk_feature |  |  |  |  |  |  |  |  |  |  |  |
| 02 |  |  | medium_integration |  |  |  |  |  |  |  |  |  |  |  |
| 03 |  |  | low_risk_feature |  |  |  |  |  |  |  |  |  |  |  |
| 04 |  |  | medium_integration |  |  |  |  |  |  |  |  |  |  |  |
| 05 |  |  | low_risk_feature |  |  |  |  |  |  |  |  |  |  |  |
| 06 |  |  | medium_integration |  |  |  |  |  |  |  |  |  |  |  |
| 07 |  |  | low_risk_feature |  |  |  |  |  |  |  |  |  |  |  |
| 08 |  |  | medium_integration |  |  |  |  |  |  |  |  |  |  |  |
| 09 |  |  | low_risk_feature |  |  |  |  |  |  |  |  |  |  |  |
| 10 |  |  | medium_integration |  |  |  |  |  |  |  |  |  |  |  |

## NDJSON Capture Contract

For each run completion, append one JSON line to:

- `runs/<window_id>.ndjson`

Schema:
- `schemas/level4-eval-record-v0.1.json`

Example line:

```json
{"window_id":"w-2026-02-l4-01","run_id":"run-001","pipeline_id":"p-low-001","pipeline_class":"low_risk_feature","scenario_total":12,"scenario_passed":12,"first_pass_success":true,"retries":1,"interventions":1,"decision":"approved","decision_reversed":false,"critical_incident":false,"timestamp":"2026-02-18T20:00:00Z"}
```

## Hard Gate Command

```bash
go run ./cmd/dfgatev01 -input runs/<window_id>.ndjson -window <window_id> -criteria profiles/level4-gate-v0.1-baseline.json -output text
```

Adversarial replay:

```bash
go run ./cmd/dfgatev01 -input runs/<window_id>.ndjson -window <window_id> -criteria profiles/level4-gate-v0.1-adversarial.json -output text
```

Exit codes:
- `0`: pass
- `2`: fail thresholds
- `1`: invalid input/system error

## Thresholds (v0.1)

- Scenario pass rate >= 90%
- First-pass success >= 70%
- Mean retries <= 2.0
- Intervention trend stable/decreasing (second half avg <= first half avg)
- Decision reversal rate <= 5% (of approved runs)
- Critical incidents from approved runs = 0

## Decision Record Template

- Window ID:
- Date range:
- Gate result (pass/conditional/fail):
- Metrics snapshot:
- Top 3 blockers:
- Corrective actions:
- Next review date:

## Discipline Rules

- No manual diff review as acceptance substitute.
- No approval without scenario results.
- No rewriting historical run records.
- If schema validation fails, run does not count.
