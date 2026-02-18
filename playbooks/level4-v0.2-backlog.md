# level 4 backlog v0.2 (first-principles, data-shaped)

## Why this exists

You now have:

- a versioned run contract (`schemas/level4-eval-record-v0.1.json`)
- deterministic gate criteria (`profiles/level4-gate-v0.1-*.json`)
- a scriptable evaluator (`cmd/dfgatev01`)
- an enforced learning log gate (`cmd/dflearn` + CI workflow)

That is enough to stop guessing and start steering by evidence.

## Current Signal (from data on hand)

Source records:

- `runs/examples/window-sample.ndjson` (2 runs)
- gate replay against baseline profile

Observed metrics:

- run_count: `2` (required: `>=10`) -> fail
- class coverage: `1 low_risk_feature / 1 medium_integration` (required: `>=4 each`) -> fail
- scenario pass rate: `95%` -> pass threshold (`>=90%`)
- first-pass rate: `100%` -> pass threshold (`>=70%`)
- mean retries: `1.5` -> pass threshold (`<=2.0`)
- decision reversals: `0%` -> pass threshold (`<=5%`)
- approved critical incidents: `0` -> pass threshold (`=0`)

Interpretation:

- Quality shape looks promising.
- Evidence volume is non-existent.
- We are data-poor, not gate-poor.

## First Principles

1. No metric, no claim.
2. No scenario result, no approval.
3. No learning entry, no merge.
4. Advance autonomy only when stability is demonstrated across windows, not single runs.

## v0.2 Objective

Make Level 4 repeatable enough that a single operator can run windows without heroics, and reviewers can trust the decision record without opening diffs.

## Priority Backlog (ranked)

## 1) Build a real 10-run baseline window

Type: Technical + Process  
Difficulty: Medium  
Why first: current blocker is sample size, not quality.

Deliverables:

- `runs/w-<date>-l4-02.ndjson` with >=10 valid records
- class mix satisfied (`>=4` each class)
- baseline gate pass artifact captured in `learning/journal/*`

Hard exit criteria:

- `dfgatev01` baseline exits `0`
- no schema-invalid lines
- all runs have timestamp + decision

Target delta:

- run_count from `2` -> `>=10`
- class coverage from `1/1` -> `>=4/>=4`

## 2) Add schema validation in ingestion path

Type: Technical  
Difficulty: Medium  
Why second: garbage records will poison every metric.

Deliverables:

- `dfgatev01` (or adjacent thin validator) rejects malformed records pre-eval
- clear error report by line number

Hard exit criteria:

- malformed record test fixture fails fast
- valid fixture still evaluates unchanged

Target delta:

- invalid-record acceptance from unknown -> `0%`

## 3) Split interventions into typed causes

Type: Specification Quality + Process  
Difficulty: Low  
Why third: intervention count alone is weak; cause taxonomy makes it actionable.

Deliverables:

- extend run record with optional intervention tags:
  - `spec_ambiguity`
  - `scenario_gap`
  - `integration_instability`
  - `operator_override`
- add simple rollup per window

Hard exit criteria:

- >=90% of interventions classified
- top 2 intervention causes visible per window

Target delta:

- unknown intervention cause share -> `<10%`

## 4) Promote learning log from narrative to queryable summary

Type: Process + Tooling  
Difficulty: Low  
Why fourth: decision memory should compile into ops decisions, not archaeology.

Deliverables:

- per-window learning summary template:
  - blockers
  - metric drift
  - corrective action + owner + due date
- one summary file per completed window

Hard exit criteria:

- every window has summary within 24h of close
- every corrective action links to evidence path

Target delta:

- missing closeout summaries -> `0`

## 5) Add adversarial replay as mandatory second gate for promotion

Type: Process + Risk Control  
Difficulty: Medium  
Why fifth: prevents local overfitting to baseline thresholds.

Deliverables:

- rule: baseline pass required for continue, adversarial pass required for promotion
- promotion decision rubric documented in playbook

Hard exit criteria:

- each completed window has both gate outputs archived
- promotion blocked if adversarial gate exits non-zero

Target delta:

- promotion decisions without adversarial evidence -> `0`

## What we have vs what we need next

What we have:

- deterministic evaluation skeleton
- versioned profiles
- enforceable learning memory

What we need next:

- statistically meaningful windows
- record hygiene guarantees
- failure-cause observability
- explicit promotion policy

## Promotion rule for v0.2 -> v0.3

Only promote if all are true across two consecutive windows:

- baseline gate pass in both windows
- adversarial gate pass in both windows
- no approved critical incidents
- intervention trend stable/decreasing
- top intervention cause has corrective action closed

If one fails, do not “discuss exceptions.” Re-run with fixes.

## Suggested cadence

- Week 1: backlog items 1 + 2
- Week 2: backlog items 3 + 4
- Week 3: backlog item 5 + two-window promotion decision

Yes, this is conservative. That is why it is useful.
