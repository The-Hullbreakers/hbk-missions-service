---
change_id: CHG-0041
artifact: design
updated: 2026-10-14
---

# Design — Contract Disputes

## Open Questions

- Dispute on what exactly: contract rejection only, or also per-credit
  rejection? [NEEDS CLARIFICATION]
- Does resolving a dispute re-open adjudication or only record an outcome?
- Dispute storage: attribute on contract item vs separate `SK` children
  (children preferred — disputes are a collection).

## Guards

`leave`/`abandon` must check open disputes on the same atomic write — same
consistency model as `CommitAcceptance`.
