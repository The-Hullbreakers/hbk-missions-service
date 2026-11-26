---
change_id: CHG-0032
artifact: tasks
status: complete
updated: 2026-10-12
---

# Tasks — CHG-0032 Review Queue

Execution mode: sequential, one agent. No domain RED — the change adds a
query over an existing state; domain vocabulary is already complete.

| ID | Task | work_type | Status | Wave |
| --- | --- | --- | --- | --- |
| T-032-001 | `ListPendingReviewHandler` + `PendingReviewLister` port; reviewer guard | application-query | done | application |
| T-032-002 | `ContractStore.ListPendingReview` — state-filtered query | persistence-adapter | done | infrastructure |
| T-032-003 | `ReviewQueueHandler` + `GET /review-queue` route + DTO | rest-endpoint | done | interface |
| T-032-004 | Composition wiring + docs + converge | composition-wiring | done | composition |

## Requirement → Test → Production Map

- **REQ-MCF-021** → app unit tests (reviewer/member/empty), infra query
  test, boundary 403/200 tests →
  `application/list_pending_review.go`,
  `infrastructure/dynamodb/contract_store.go`,
  `interface/http/review_queue.go`,
  `composition/server.go`.

## RED Plan

| Task | Test | Layer | Expected RED |
| --- | --- | --- | --- |
| T-032-001 | `tests/unit/missions/application/list_pending_review_test.go` | unit | compile: `NewListPendingReviewHandler` undefined |
| T-032-002 | `tests/integration/infrastructure/dynamodb/contract_store_pending_review_test.go` | infra | compile: `ListPendingReview` undefined |
| T-032-003 | `tests/integration/http/review_queue_test.go` | boundary | compile: `ReviewQueueHandler` undefined |
