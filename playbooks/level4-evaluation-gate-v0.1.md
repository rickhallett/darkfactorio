# level 4 evaluation gate v0.1

## Purpose

Decide if `darkfactorio layer v0.1` is stable enough for broader Level 4 use, before any Level 5 discussion.

## Evaluation Window

- Minimum: 10 full runs across at least 2 distinct pipeline specs
- Recommended: 20+ runs across normal and high-change weeks

## Required Metrics

- Scenario pass rate
- First-pass run success rate
- Mean retries per run
- Human intervention rate (`awaiting_human` events per run)
- Decision reversal rate (approved then rolled back/rejected downstream)
- Incident count linked to approved runs

## Acceptance Thresholds (v0.1)

- Scenario pass rate: >= 90%
- First-pass success: >= 70%
- Mean retries per run: <= 2.0
- Human interventions: stable or decreasing trend over window
- Decision reversals: <= 5%
- Critical incidents from approved runs: 0

## Qualitative Review

- Are rejects mostly due to known ambiguity classes?
- Are operators making consistent decisions across similar outcomes?
- Are runbooks sufficient without ad-hoc heroics?

## Exit Outcomes

1. **Pass Level 4 Gate**
- Continue with expansion to more pipelines.
- Keep Level 5 in evaluation-only discussion mode.

2. **Conditional Pass**
- Expand only under constraints (limited pipeline classes).
- Close top failure classes first.

3. **Fail Gate**
- Freeze expansion.
- Run remediation sprint on spec quality, scenario quality, or runtime policy.

## Required Decision Record

- Window dates
- metrics snapshot
- gate outcome (pass/conditional/fail)
- top blockers
- next review date

