# factory boot sequence v0.1

## Goal

Provide deterministic startup and run initialization for repeatable Level 4 operations.

## Sequence

1. **Profile Resolve**
- Load `FactoryProfile`.
- Resolve provider profile, execution environment, policy defaults.
- Fail hard if required keys/env/contracts are missing.

2. **Spec Resolve**
- Load immutable `PipelineSpec@version`.
- Load immutable `ScenarioSuite@version`.
- Validate compatibility (`pipeline_id`, major version, required capabilities).

3. **Preflight Validation**
- Run structural validation/lint against pipeline graph.
- Verify handler registry coverage for all node types in graph.
- Verify external dependency policy (real/twin) for this environment.

4. **Run Create**
- Create `Run` envelope with ids and timestamps.
- Set `status=booting`.
- Allocate run directory.
- Seed initial context and checkpoint.

5. **Execution Init**
- Set `status=executing`.
- Start traversal at start node.
- Emit `run.started` event.

6. **Runtime Gatepoints**
- On `wait.human`, set `status=awaiting_human`.
- On continue, return to `status=executing`.
- On terminal node, move to `status=evaluating`.

7. **Scenario Evaluation**
- Execute holdout `ScenarioSuite` outside agent-visible context.
- Collect pass/fail and failure signatures.

8. **Decision Gate**
- If scenarios pass and policy checks pass: allow `approved` decision.
- Else require `rejected` decision with remediation path.
- Persist `DecisionRecord`.

9. **Finalize**
- Emit `run.completed` event.
- Freeze run artifacts and logs.
- Publish run summary.

## Failure Domains

- `BOOT_FAILURE`: profile/spec/preflight failures
- `EXEC_FAILURE`: handler/runtime failures
- `EVAL_FAILURE`: scenario execution or scenario fails
- `DECISION_FAILURE`: missing or invalid decision record

## Recovery Rules

- `BOOT_FAILURE`: fix config/spec, create new run.
- `EXEC_FAILURE`: retry per policy or route to failure handler.
- `EVAL_FAILURE`: no approval path; must remediate and rerun.
- `DECISION_FAILURE`: block closure until valid decision recorded.

## Required Events (minimum)

- `run.created`
- `run.started`
- `stage.started`
- `stage.completed`
- `stage.failed`
- `run.awaiting_human`
- `run.evaluating`
- `run.approved`
- `run.rejected`
- `run.failed`
- `run.completed`

