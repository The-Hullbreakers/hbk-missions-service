# CHG-0055 — Create local season preview scenarios: mutation evidence

mutation_scope: local endpoint guard, conflict protection, active-user
minimum, per-unit merit calculation, and completion reward.
mutation_tool: Python stdlib temporary-source-copy runner (no native
mutation tool was installed).
mutation_result: 5 selected mutants killed / 5 executed.
status: selected_scope_passed

The lead ran the normal 10-test Python unit suite first. Each mutant
was applied to a fresh copy of the four scenario modules and the
unit test in a temporary directory, then executed with
`python3 -B -m unittest discover -s <temp>/tests/unit/seeds -p
test_season_scenarios.py`; the source in the repository never changed.
Each modified anchor was asserted unique, and import errors were
excluded as kill evidence. All five failed on behavioral assertions:

| Mutant | Result | Observed failure |
| --- | --- | --- |
| Remove empty-userinfo rejection | killed | 1 assertion failure |
| Ignore divergent rows | killed | 1 assertion failure |
| Permit only two active users | killed | 1 assertion failure |
| Subtract rather than add per-unit bonus | killed | 1 assertion failure |
| Omit completed-contract reward | killed | 1 assertion failure |

The bonus and completion-reward assertions were added to
`tests/unit/seeds/test_season_scenarios.py` for the existing approved
credit semantics; the final suite passed 10/10 and the test file is
142 physical lines. No runner-wide mutation score or Python-wide
production coverage is claimed: this is a risk-selected five-mutant
sample, not an exhaustive automated campaign. No selected mutant
survived or timed out.
