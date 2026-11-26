#!/usr/bin/env bash
# Creates season "season-0" starting now, with an automatic succession cycle:
#   ACTIVE for ACTIVE_DAYS days -> FREEZING (off-season) for OFF_DAYS days
#   -> scheduler auto-creates the next iteration (season-1, season-2, ...).
#
# Usage:
#   BASE_URL=http://localhost:8080 ADMIN_TOKEN=<jwt> ./scripts/start_season_0.sh
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
PREFIX="${PREFIX:-season}"
ITERATION="${ITERATION:-0}"
ACTIVE_DAYS="${ACTIVE_DAYS:-30}"
OFF_DAYS="${OFF_DAYS:-3}"
STARTS_AT="${STARTS_AT:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
OFFICIAL="${OFFICIAL:-true}"

if [[ -z "${ADMIN_TOKEN:-}" ]]; then
  echo "error: ADMIN_TOKEN env var required (admin JWT)" >&2
  exit 1
fi

body=$(cat <<EOF
{
  "starts_at": "${STARTS_AT}",
  "official_statistics": ${OFFICIAL},
  "schedule": {
    "active_days": ${ACTIVE_DAYS},
    "off_days": ${OFF_DAYS},
    "series": { "prefix": "${PREFIX}", "iteration": ${ITERATION} }
  }
}
EOF
)

echo "creating ${PREFIX}-${ITERATION} @ ${STARTS_AT} (active=${ACTIVE_DAYS}d off=${OFF_DAYS}d)"
curl -fsS -X POST "${BASE_URL}/api/v1/seasons" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "${body}"
echo
