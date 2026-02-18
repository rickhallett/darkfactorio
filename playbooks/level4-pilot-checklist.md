# Level 4 Pilot Checklist

## Pilot Definition
- Target workflow:
- Pilot scope:
- Success metric:
- Abort criteria:

## Preconditions
- [ ] Spec template completed with explicit non-behaviors
- [ ] Holdout scenario suite defined outside agent context
- [ ] Integration boundaries documented
- [ ] Safety rollback path defined
- [ ] Owner assigned for outcome evaluation

## Run Steps
1. Freeze spec and scenario suite for this run.
2. Execute autonomous implementation pass.
3. Evaluate against holdout scenarios only.
4. Record pass/fail and observed deviations.
5. Approve/reject outcome with decision log.

## Decision Log
| Date | Build/Commit | Outcome | Reason | Next Action |
|---|---|---|---|---|
| YYYY-MM-DD | <sha> | approve/reject | <why> | <next> |

## Exit Criteria (Pilot Success)
- [ ] Outcome pass rate meets threshold
- [ ] Human diff-reading time reduced materially
- [ ] No critical regression escaped scenario suite
- [ ] Spec quality improved between runs

## Failure Patterns to Watch
- Ambiguous specs creating “technically correct, practically wrong” output
- Agent overfitting to visible tests instead of behavior
- Humans quietly returning to full diff review due to trust gaps
