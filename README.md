# darkfactorio

`darkfactorio` is a working notebook for people who have stopped pretending that faster keystrokes are the same thing as software leverage.

If your current loop is “AI writes code, I read every diff, feel productive, ship slower,” welcome. You are not uniquely cursed. You are in the statistical majority.

This repo exists to push from Level 2-3 behavior into Level 4 operations, with a credible path to Level 5.

## What This Is

A practical operating model for autonomous software delivery where:

- humans specify what should exist
- agents implement
- humans evaluate outcomes
- nobody plays hero by manually polishing diffs at 2am

In plain terms: we are trying to build a software factory, not an autocomplete addiction.

## What This Is Not

- a prompt scrapbook
- a vibe-coding highlight reel
- a theological argument about whether AI is "good"
- a cosplay version of Level 5 where you still review every line and call it autonomous

## The Core Thesis

The bottleneck has moved.

Old bottleneck: implementation throughput.
New bottleneck: specification clarity + behavioral evaluation quality.

If the spec is mush, the software is mush at machine speed.

## The Levels (Operationally, Not Marketing)

- Level 0: spicy autocomplete
- Level 1: coding intern (single tasks)
- Level 2: junior dev (multi-file changes, human reads diffs)
- Level 3: manager mode (AI implements, human still diff-bound)
- Level 4: PM mode (spec in, outcomes out; no routine diff reading)
- Level 5: dark factory (specification to shippable artifacts, autonomously)

`darkfactorio` assumes most teams sit at Level 2/3 and overestimate their altitude.

## Non-Negotiable Principles

1. **Code must not be written by humans (as default workflow).**
2. **Code must not be reviewed by humans (as default workflow).**
3. **Humans own specification quality and outcome approval.**
4. **Evaluation must be behavior-first, not test-gaming-friendly.**
5. **If it cannot be specified clearly, it is not ready for autonomous implementation.**

Yes, there are exceptions for emergencies. No, exceptions are not the operating model.

## Why Most AI Adoption Underperforms

Because teams bolt new engines onto old transmissions.

Symptoms:

- larger diffs, higher review fatigue
- “almost right” code that burns debugging time
- perceived speed gains, measured slowdowns
- rituals designed for human implementation still running full blast

If your process still assumes humans are the primary code producers, your AI stack is mostly expensive theater.

## Operational Requirements

To move toward Level 4/5, these must exist and be healthy:

- **Spec discipline**: structured, unambiguous markdown specs
- **Scenario architecture**: external behavioral checks (holdout-style), isolated from agent context
- **Execution environment**: safe integration surface (digital twins/simulated dependencies when needed)
- **Gates**: outcome acceptance criteria tied to user-visible behavior, not pretty diffs
- **Decision logs**: why we approved/rejected outcomes

No amount of prompt wizardry substitutes for missing structure.

## Spec Standard (Minimum Viable)

Every implementation spec should answer, concretely:

- problem and user impact
- in-scope / out-of-scope behavior
- invariants and constraints
- integration expectations
- failure modes and edge cases
- acceptance scenarios (observable behavior)
- rollback or mitigation conditions

If a spec cannot survive contact with ambiguity, it is not a spec yet. It is a wish.

## Scenario Standard (Anti-Gaming)

Scenarios should be:

- behavior-first (black-box)
- external to implementation context when possible
- difficult to overfit via obvious test leakage
- mapped to explicit pass/fail criteria

The point is not “tests passed.”
The point is “the intended behavior exists under realistic conditions.”

## Migration Reality (for Legacy Systems)

You do not dark-factory a legacy monolith by declaration.

Practical sequence:

1. Use Level 2/3 workflows where they already help.
2. Extract and document actual system behavior (not aspirational docs).
3. Build scenario suites that encode current and desired behavior.
4. Upgrade CI/CD gates for high-volume AI-generated change.
5. Shift greenfield/new modules to Level 4 patterns first.
6. Expand autonomy boundary only when quality signals hold.

This takes years. Anyone promising quarters is selling a keynote.

## Org Design Implications

Coordination-heavy structures become drag when implementation becomes abundant.

Value shifts from:

- task coordination -> system articulation
- diff review -> outcome judgment
- throughput management -> specification quality

The hard skill is no longer typing faster.
It is being precise enough that machines can execute without hallucinating your intent.

## What Success Looks Like Here

For `darkfactorio`, success is:

- less human time spent reading diffs
- more human time spent defining and evaluating outcomes
- tighter specs, stronger scenarios, fewer ambiguity loops
- expanding zones where autonomous implementation is trustworthy

If we are still debating line-level style in giant generated PRs, we are not succeeding.

## Usage

Use this repository as:

- an operating handbook
- a source of reusable spec templates and scenario patterns
- a log of what actually worked versus what sounded smart on X

Template index: `templates.md`

## darkfactorio layer v0.1

- Architecture: `factory/darkfactorio-layer-v0.1.md`
- Boot sequence: `factory/boot-sequence-v0.1.md`
- Operations manual: `manuals/level4-operations-manual-v0.1.md`
- Evaluation gate: `playbooks/level4-evaluation-gate-v0.1.md`
- Run 01-10 sheet: `playbooks/level4-run-01-10-execution-sheet-v0.1.md`
- Run schema: `schemas/run-envelope-v0.1.json`
- Eval record schema: `schemas/level4-eval-record-v0.1.json`

Scriptable gate evaluator (Go, versioned):

- baseline: `go run ./cmd/dfgatev01 -input runs/<window_id>.ndjson -window <window_id> -criteria profiles/level4-gate-v0.1-baseline.json -output text`
- adversarial replay: `go run ./cmd/dfgatev01 -input runs/<window_id>.ndjson -window <window_id> -criteria profiles/level4-gate-v0.1-adversarial.json -output text`

## Final Note

The dark factory does not remove the need for good engineers.

It removes hiding places for mediocre thinking.

Welcome to the fun part.

## Author

- Kai (`kai@oceanheart.ai`)
- Rick Hallett (`rickhallett@icloud.com`)
- <https://www.oceanheart.ai>
