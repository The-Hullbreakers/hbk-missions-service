# HBK agent routing

- Work only in this repository and the approved task scope. Read README.md, specs/index.md when present, and applicable local rules before editing.
- For L1-L3 behavior changes, use `/common-sdd-spec` and `/common-sdd-change-lifecycle` as the parent lifecycle, then one Go/React implementation primary. Obtain one human approval for the complete plan and tests before spec writes or RED; pass scoped technical Gate 3 before GREEN, then obtain one final human review after all checks pass and before consolidation. L0 docs and agent configuration use native checks. Load supporting rules only for affected boundaries.
- Preserve business language, TDD RED/GREEN/BLUE, Clean Architecture dependency direction and the repository test conventions. Do not create speculative layers or tests that mirror implementation.
- For identity, tenant, authorization, secrets, public API, CI or cloud permissions, request an independent portfolio-security review. For PRs, request portfolio-pr-reviewer.
- Preserve existing changes. Report actual verification commands and results; never invent approval or deploy, publish or merge without authorization.
