# shadow-pack v0.1

`shadow-pack` enforces separation between implementation and QA feedback loops.

Core rule:

- `candidate_producer` must be different from `holdout_producer`.

The evaluator compares:

- overlap coverage
- outcome mismatch rate
- p95 latency drift

Run:

```bash
make shadow-pack
```

or:

```bash
go run ./cmd/dfshadowv01 --manifest shadowpacks/examples/manifest.json --output text
```

This is a scaffold for independent holdout packs from external repos/teams.

## Onboarding a real project

Scaffold:

```bash
make onboard-project PROJECT=tspit
```

Validate generated artifact contracts:

```bash
make onboard-validate PROJECT=tspit
```

Then replace `shadowpacks/tspit/candidate.json` and `shadowpacks/tspit/holdout.json` with real outputs from independent producers.
