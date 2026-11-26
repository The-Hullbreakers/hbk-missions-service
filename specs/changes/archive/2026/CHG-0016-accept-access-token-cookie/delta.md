# Delta — mission-contract-fulfillment

## Requirement Changes

- **MODIFIED — REQ-MC-AUTH**: `/api/v1/**` endpoints accept the OIDC token via `Authorization: Bearer <token>` **or** the HttpOnly `access_token` cookie issued by HBK Identity. The `Authorization` header takes precedence when both are present. `401 Unauthorized` still applies when neither carries a valid token.

## Scenario Changes

- **REG-001 — Cookie-authenticated request**: Given a request with a valid `access_token` cookie and no `Authorization` header, when it reaches an `/api/v1/**` route, then it is authenticated and the caller identity comes from the cookie token.
- **REG-002 — Header precedence**: Given both a Bearer header and an `access_token` cookie, when authenticating, then the header token is used.
- **REG-003 — Invalid cookie token**: Given an `access_token` cookie with an invalid token, when authenticating, then the response is `401`.
