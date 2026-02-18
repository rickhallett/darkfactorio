# Specification: <initiative>

## System Overview
- What this is:
- Who it serves:
- Why it exists:

## Behavioral Contract

### Primary Flows
- When <condition>, the system <observable behavior>.
- When <condition>, the system <observable behavior>.

### Error Flows
- When <dependency failure>, the system <observable behavior>.
- When <invalid input>, the system <observable behavior>.

### Boundary Conditions
- When <limit condition>, the system <observable behavior>.
- When <edge input>, the system <observable behavior>.

## Explicit Non-Behaviors
- The system must not <behavior> because <reason>.
- The system must not <behavior> because <reason>.

## Integration Boundaries
| External System | Data In | Data Out | Contract | Failure Behavior | Real vs Twin |
|---|---|---|---|---|---|
| <system> | <in> | <out> | <format + guarantees> | <timeout/retry/fallback> | <real/twin> |

## Behavioral Scenarios (Holdout)

### Happy Path
1. **Scenario HP-1: <name>**
- Setup:
- Action:
- Expected outcomes:

2. **Scenario HP-2: <name>**
- Setup:
- Action:
- Expected outcomes:

3. **Scenario HP-3: <name>**
- Setup:
- Action:
- Expected outcomes:

### Error
4. **Scenario ER-1: <name>**
- Setup:
- Action:
- Expected outcomes:

5. **Scenario ER-2: <name>**
- Setup:
- Action:
- Expected outcomes:

### Edge Case
6. **Scenario EC-1: <name>**
- Setup:
- Action:
- Expected outcomes:

7. **Scenario EC-2: <name>**
- Setup:
- Action:
- Expected outcomes:

## Ambiguity Warnings
- Ambiguity:
- Likely agent assumption:
- Decision required:

## Implementation Constraints
- Required language/framework (if any):
- Architectural constraints (if any):
- Security/compliance constraints (if any):
