# CHG-0056 - Security review plan

status: pending
role: authenticated member
trust_boundary: session -> HTTP -> overall Application query -> local ledger

Risk: a request body/query must never select another member's
`my_standing`. Only `domain.CallerIdentity` extracted from the
authenticated context identifies the caller. Missing auth returns
401; missing merits returns `my_standing: null`. Do not return emails,
handles, token claims, media/evidence, contract IDs, private user
metadata or raw ledger rows. Team credits never enter the individual
overall result. Bound limit/cursor, avoid unbounded response payloads,
and report read amplification without logging recipient identities.

Checks pending: caller isolation, correct 400/401/404 codes,
error content safety, local/cloud credentials separation, query
latency/cost, and no Critical/High findings. The identity service
owns public handle exposure; this backend does not read user records.
