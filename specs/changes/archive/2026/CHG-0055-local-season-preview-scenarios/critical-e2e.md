# CHG-0055 — Create local season preview scenarios: critical E2E

critical_e2e_id: TEST-0055-006
scenario_id: SCN-0055-006
status: passed_with_identity_simulation

Given the persistent mid-season preview in DynamoDB Local on
`http://localhost:8000` and an authenticated member whose ID comes
from its contract roster, when they visit the public missions routes,
then they see the active unofficial season, offers, four leaderboard
buckets, a populated PvP Air ranking and at least one of their own
contracts. A request without authentication is rejected.

`go test -tags=integration ./tests/integration/http -run
'^TestSeasonPreviewJourney$' -count=1` passed six subtests through
the real composition root, HTTP handlers, application and DynamoDB
Local. The external OIDC session is simulated; this does not prove
login, browser rendering, or actual screenshot capture. The test
reads an existing preview table, does not modify identity data, and
does not delete any table. See `verification.md` for detailed output.
