# CHG-0056 - Rollout and rollback

status: pending

1. Approve this delta and each test-first layer gate; do not deploy
   an endpoint before its authorized Go core, HTTP and persistence
   checks are green.
2. Verify Cloud SAM event and the separately owned local SAM route
   select the same authenticated handler. Exercise the local preview
   without deleting or altering its tables.
3. Deploy the additive endpoint before the web client consumes it;
   leave the category leaderboard, category `/me`, statistics and
   history routes intact for older clients.
4. Validate off-season, fresh-start, mid-season, missing auth,
   category isolation and request cost. If identity aliases are not
   approved, frontend uses a non-identifying rival label.
5. Roll back the new route/client on failure. No data migration or
   automatic table cleanup is planned. Existing APIs must keep
   working throughout.
