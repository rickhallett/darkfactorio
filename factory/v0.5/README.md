# darkfactorio v0.5 validation layer

v0.5 extends v0.4 with first-principles checks for remaining dark-factory gaps:

1. spec execution evidence (real command + artifact)
2. holdout provenance integrity (sha256 attested)
3. twin-vs-real drift budgets
4. deploy+rollback execution evidence
5. runtime SLO gate input
6. provider-vs-internal economics reconciliation
7. red-team detection channel
8. tamper-evident policy hash chain
9. portfolio scheduling priority correctness

Run:

```bash
make factory-v05-validate
```

Exit codes:

- `0`: all checks pass
- `2`: one or more checks fail
- `1`: invalid input/runtime error
