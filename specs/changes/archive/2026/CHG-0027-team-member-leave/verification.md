---
change_id: CHG-0027
artifact: verification
status: verified
updated: 2026-10-09
---

# Verification — Team Member Leave

## RED → GREEN Evidence

- **Domain (T-01/T-02)** — `contract_leave_test.go`: 7 scenarios RED as
  compile errors (`Leave`, `IsDeparted`, `ActiveMembers` undefined) → GREEN
  after `departed` set + `Leave` + `CanSubmitEvidence` exclusion.
- **Application (T-03)** — `leave_contract_test.go`: 3 scenarios RED
  (`NewLeaveContractHandler` undefined) → GREEN with
  `internal/missions/application/leave_contract.go`.
- **HTTP boundary (T-04)** — `contract_lifecycle_leave_test.go`: 200 alive,
  403 outsider, last-member → ABANDONED via the shared `transition` adapter.
- **Persistence (T-05)** — `contract_store_leave_test.go`: `Departed`
  round-trip through real DynamoDB-local + `ListActiveByPlayer` no longer
  counts the leaver's contract (slot freed); survivor still counts it.
- **Composition/E2E (T-06/T-07)** — `team_leave_e2e_test.go`: seeded
  2-member team contract → player-1 leaves over HTTP → departed evidence
  `403` → survivor confirms (`201`) and completes (`200`) → TEAM
  leaderboard shows `team-9` with evidence merit.

## Gates

- **Test-first** — every layer's RED classified correctly before its GREEN.
- **BLUE** — all touched files `< 150` physical lines; gofmt/vet clean.
- **Security** — `security-review.md`: no Critical/High findings.
- **Validation** — `tools/validate-sdd-change.sh` PASSED on 2026-10-09:
  unit + integration suites green under race detector; coverage
  domain 91.4%, application 91.9%, dynamodb 88.1%.

## Notes

- HTTP `accept` always mints a single-member roster, so the e2e journey
  seeds the 2-member contract directly in DynamoDB — documented in the
  test header. Multi-member join is a separate capability.
