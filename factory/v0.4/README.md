# darkfactorio v0.4 validation layer

This layer adds concrete contract checks for the seven missing dark-factory aspects:

1. spec-to-implementation contract quality
2. external holdout scenario quality
3. digital twin health/contract readiness
4. release + rollback gate completeness
5. policy/compliance attestations with evidence
6. economic budget adherence
7. multi-agent orchestration structure

## command

```bash
make factory-v04-validate
```

or:

```bash
go run ./cmd/dffactoryv04 --bundle factory/v0.4/examples/bundle.json --output text
```

Exit codes:

- `0`: all seven layers validated
- `2`: one or more layer checks failed
- `1`: invalid input/runtime issue

## what this is / isn't

- It **is** an enforceable infrastructure contract for dark-factory readiness.
- It **is not** a full autonomous code-generation runtime.
- It gives you a deterministic scaffold to evolve toward Level 5 without hand-waving.
