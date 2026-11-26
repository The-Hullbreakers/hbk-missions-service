# CHG-0056 — Season overall standings: mutation evidence

mutation_scope: domain ranking policy, application use
case, HTTP error mapping, DynamoDB boundary filter.
mutation_tool: manual selected-mutant campaign (no
native Go mutation tool installed; repo precedent
CHG-0055). Each mutant applied to the real file,
focused suite executed, file restored from backup.

## Campaign (2026-11-26) — 10 mutants, all killed

| Mutant | Scope | Result |
| --- | --- | --- |
| drop individual-track filter | domain | killed |
| drop season filter | domain | killed |
| replace merit aggregation | domain | killed |
| count raw credits as objectives | domain | killed |
| count every contract as completed | domain | killed* |
| reverse ranking order | domain | killed |
| first entry becomes my standing | application | killed |
| my standing rank off-by-one | application | killed |
| season error maps to 500 not 404 | http | killed |
| drop track filter in credit query | dynamodb | killed |

*First flagged SURVIVED: the existing fixtures marked
every contract completed, so counting all contracts
was data-equivalent. New pin
`TestOverallStandings_InProgressContractDoesNotCountCompletion`
asserts an in-progress contract adds no completion —
the mutant is now killed. The earlier reformulated
mutants (perl delimiter failures) were rerun
properly; only real mutants count.

status: passed — no unexplained survivor.
