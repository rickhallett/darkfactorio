# level 4 operations manual v0.1

## Operator Mission

Run pipelines safely and repeatedly without human code review, using behavior outcomes as the acceptance boundary.

## Daily Workflow

1. Queue run from pinned `PipelineSpec@version` and `ScenarioSuite@version`.
2. Monitor stage events and runtime status transitions.
3. Resolve human gates quickly with explicit rationale.
4. Trigger scenario evaluation on terminal execution.
5. Issue explicit approve/reject decision.
6. Log deviations and open remediation tasks.

## Runbook: Start a Run

- Confirm profile/environment target.
- Confirm spec and scenario versions are pinned.
- Verify no active incident on shared dependencies.
- Create run and monitor boot sequence.

## Runbook: Handle Awaiting Human

- Read context + edge options.
- Select path using policy, not intuition alone.
- Record why this branch was selected.
- Resume run.

## Runbook: Failed Stage

- Classify error: retryable, terminal, pipeline.
- Apply retry policy only if retryable.
- If terminal, route to fail path and continue to evaluation summary.
- Open incident entry if failure is systemic.

## Runbook: Scenario Failures

- Block approval immediately.
- Map failure to spec ambiguity, implementation defect, or environment issue.
- Create remediation spec delta.
- Rerun with new versions; do not mutate old run artifacts.

## Decision Policy

Approve only when all are true:
- holdout scenarios passed
- no unresolved critical failures
- policy checks passed
- operator can state user-visible behavior outcome clearly

Reject when any are true:
- scenario failure
- unresolved ambiguity with user impact
- inconsistent outcome across repeated runs
- missing audit trail or missing decision rationale

## Anti-Patterns (forbidden)

- approving based on "looks good" without scenario evidence
- bypassing scenario gate to hit timeline
- modifying historical run artifacts
- diff-driven manual patching as a hidden workflow

## Handoff Artifacts

Every run handoff must include:
- run summary
- scenario report
- decision record
- open risks and next actions

