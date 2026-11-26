# Tasks — CHG-0016

## Wave 1 — Interface (RED)

- **T-1 — Add cookie-auth regression tests** (`tests/unit/missions/interface/http/auth_middleware_test.go`): REG-001 cookie authenticates, REG-002 header precedence, REG-003 invalid cookie → 401.
  - Owner: agent. Depends on: none. Verification: `go test` shows new cases FAIL for the expected reason (401 before cookie support). Next: Gate 3.

## Wave 2 — Interface (GREEN + BLUE)

- **T-2 — Implement cookie fallback in `AuthMiddleware.Authenticate`** (`internal/missions/interface/http/middleware.go`): header first, else `access_token` cookie → `Bearer` token → existing session port.
  - Owner: agent. Depends on: T-1, Gate 3. Verification: REG-001..003 green + existing tests green. Next: BLUE refactor while green.

## Wave 3 — Evidence and docs

- **T-3 — Update docs and consolidate**: `docs/API-GUIDE-FRONTEND.md` auth line; run unit + composition suites; verify-spec → consolidate into `mission-contract-fulfillment` spec → archive.
  - Owner: agent. Depends on: T-2. Verification: suites green, index updated, archive complete.
