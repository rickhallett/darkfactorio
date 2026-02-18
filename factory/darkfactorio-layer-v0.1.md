# darkfactorio layer v0.1

## Purpose

`darkfactorio layer v0.1` is the operational layer on top of Attractor for repeatable Level 4 execution.

Attractor already specifies:
- orchestration mechanics
- agent loop behavior
- unified provider contracts

This layer adds what production operations need:
- stable resource model (CRUD)
- deterministic boot sequence
- operator runbooks
- evaluation gates for go/no-go decisions

## Scope

In scope:
- Level 4 repeatable operation for feature delivery
- human approval on outcomes, not line diffs
- external scenario evaluation and decision logs

Out of scope:
- full autonomous promotion to production without human sign-off
- Level 5 no-human-review shipping
- brownfield migration automation

## Operating Model

Control-plane resources:
- `FactoryProfile`: runtime policy and environment bindings
- `PipelineSpec`: executable workflow definition (Attractor-compatible)
- `ScenarioSuite`: external holdout behavior checks
- `Run`: one execution instance of a pipeline
- `RunCheckpoint`: point-in-time run state for resume/audit
- `DecisionRecord`: explicit approve/reject with rationale

Data-plane artifacts:
- prompts/responses
- status files
- generated code artifacts
- scenario results
- incident logs

## CRUD Contract (v0.1)

### FactoryProfile
- Create: define environment, model profiles, policy defaults
- Read: inspect profile and effective policy
- Update: rotate model/env bindings or limits
- Delete: retire profile (deny if active runs exist)

### PipelineSpec
- Create: register graph + metadata + version
- Read: fetch exact version and validation result
- Update: create new immutable version
- Delete: soft-delete only when no active run references it

### ScenarioSuite
- Create: register holdout scenarios for a pipeline version
- Read: list scenarios and historical pass/fail
- Update: versioned update only
- Delete: archive only; never hard-delete evaluated suites

### Run
- Create: instantiate from `PipelineSpec@version + ScenarioSuite@version`
- Read: status, stage outcomes, artifacts, logs
- Update: limited to operator actions (pause/resume/cancel/intervene)
- Delete: forbidden in v0.1; runs are immutable audit records

### DecisionRecord
- Create: operator approval or rejection
- Read: decision history and rationale
- Update: append-only amendments
- Delete: forbidden in v0.1

## Interfaces

### Canonical run status values
- `queued`
- `booting`
- `executing`
- `awaiting_human`
- `evaluating`
- `approved`
- `rejected`
- `failed`
- `canceled`

### Canonical outcomes
- `success`
- `retry`
- `fail`
- `partial_success`

These align to Attractor `status.json` outcome semantics.

## Guardrails

- No deploy/promotion when ScenarioSuite has unresolved failures.
- No silent retries beyond configured policy.
- No run approval without DecisionRecord.
- All operator interventions are logged as events.

## Minimal Implementation Dependencies

- Attractor execution model (`attractor/attractor-spec.md`)
- Attractor agent loop model (`attractor/coding-agent-loop-spec.md`)
- Attractor unified LLM client model (`attractor/unified-llm-spec.md`)

## Exit Criteria for v0.1

v0.1 is considered shipped when:
- one pipeline can be run end-to-end through boot, execution, evaluation, decision
- run state is queryable by status and id
- scenario gate can block approval
- decision record is mandatory and persisted

