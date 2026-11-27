#!/usr/bin/env bash
# test-workflow-contract.sh verifies the CI/CD workflow contract for CHG-0014.
set -euo pipefail

cd "$(dirname "$0")/.."

WORKFLOW=".github/workflows/ci-cd.yml"

fail() {
	echo "CONTRACT FAIL: $1" >&2
	exit 1
}

[ -f "$WORKFLOW" ] || fail "$WORKFLOW does not exist"

require() {
	grep -q "$1" "$WORKFLOW" || fail "missing: $1"
}

# Triggers
require 'pull_request:'
require 'staging'
require 'main'
require 'workflow_dispatch:'

# Safety defaults
require 'contents: read'
require 'cancel-in-progress: true'
require 'persist-credentials: false'

# Canonical jobs
require 'sdd-policy:'
require 'unit-tests:'
require 'integration-tests:'
require 'tools/validate-sdd-change.sh'
require 'go test ./tests/unit/...'
require 'go test -p=1 ./tests/integration/...'
require 'docker-compose.test.yml'
require 'DYNAMODB_TEST_ENDPOINT'

# Deploy jobs gated on tests and environments
require 'deploy-staging:'
require 'deploy-production:'
require 'needs: \[unit-tests, integration-tests, build\]'
require 'environment: staging'
require 'environment: production'
require 'sam deploy'
require 'test_seed_missions.py'
require 'test_seed_integration.py'
require 'SeedDeploymentRoleName'
require 'scripts/seed_missions.py --environment staging'
require 'scripts/seed_missions.py --environment production'
require 'table hbk-missions-stg --region us-east-2 --apply'
require 'table hbk-missions-prod --region us-east-2 --apply'
require 'GITHUB_STEP_SUMMARY'

# SAM template wires the shared API Gateway
TEMPLATE="template.yaml"
[ -f "$TEMPLATE" ] || fail "$TEMPLATE does not exist"
grep -q 'hbk-shared-infrastructure-${Environment}-ApiGatewayId' "$TEMPLATE" || fail "template missing shared ApiGatewayId import"
grep -q 'AWS::ApiGatewayV2::Integration' "$TEMPLATE" || fail "template missing ApiGatewayV2 integration"
grep -q 'AWS::ApiGatewayV2::Route' "$TEMPLATE" || fail "template missing ApiGatewayV2 routes"

# TEST-STG-CURRENT-SEASON-ROUTE-001: the current-season GET reaches the shared Lambda integration.
sed -n '/^  MissionsCurrentSeasonRoute:$/,+4p' "$TEMPLATE" |
  grep -Fxq '      RouteKey: GET /api/v1/missions/seasons/current' || fail 'template missing current-season GET route'

# TEST-STG-SEASON-STATISTICS-ROUTE-001: Gateway exposes the authenticated season statistics route.
statistics_route=$(sed -n '/^  MissionsSeasonStatisticsRoute:$/,+4p' "$TEMPLATE")
grep -Fxq '    Type: AWS::ApiGatewayV2::Route' <<< "$statistics_route" &&
  grep -Fxq '      <<: *RouteProps' <<< "$statistics_route" &&
  grep -Fxq '      RouteKey: GET /api/v1/missions/seasons/{season_id}/statistics' <<< "$statistics_route" ||
  fail 'template missing season statistics GET route or shared integration'

# TEST-STG-OFFER-FACETS-ROUTE-001: Gateway exposes the authenticated offer facets route.
facets_route=$(sed -n '/^  MissionsOfferFacetsRoute:$/,+4p' "$TEMPLATE")
grep -Fxq '    Type: AWS::ApiGatewayV2::Route' <<< "$facets_route" &&
  grep -Fxq '      <<: *RouteProps' <<< "$facets_route" &&
  grep -Fxq '      RouteKey: GET /api/v1/missions/offers/facets' <<< "$facets_route" ||
  fail 'template missing offer facets GET route or shared integration'

# No deploy secrets on pull requests is enforced by if-conditions on push/dispatch
require "github.ref == 'refs/heads/staging'"
require "github.ref == 'refs/heads/main'"

# TEST-STG-EVENTS-001: staging passes the exported topic to SAM or stops before deploy.
# Arrange
staging_deploy=$(sed -n '/^  deploy-staging:$/,/^  deploy-production:$/p' "$WORKFLOW" |
  sed -n '/^      - name: SAM deploy$/,/^      - name: Set up Python /p' |
  sed '1,/^        run: |$/d; /^      - name: /d; s/\${{[^}]*}}/fixture/g')
[ -n "$staging_deploy" ] || fail 'staging SAM deploy body is missing'
aws() {
  [ "$*" = "cloudformation list-exports --region us-east-2 --query Exports[?Name=='hbk-shared-infrastructure-stg-EventsTopicArn'].Value | [0] --output text" ] || return 64
  printf '%s\n' "$TOPIC_RESULT"
}
sam() { printf '%s\n' "$@"; }
export -f aws sam
expected_topic='arn:aws:sns:us-east-2:123456789012:hbk-domain-events-stg'
# Act
success=$(TOPIC_RESULT="$expected_topic" bash -ec "$staging_deploy")
# Assert
grep -Fxq "ParameterKey=SharedEventsTopicArn,ParameterValue=$expected_topic" <<< "$success" ||
  fail 'staging SAM deploy does not receive the exported events topic ARN'

# TEST-STG-CATALOG-IAM-001: staging must not ask the stack to manage its own deploy role.
if grep -q '^ParameterKey=SeedDeploymentRoleName,' <<< "$success"; then
  fail 'staging SAM deploy still asks CloudFormation to manage the deployment role policy'
fi

# Arrange
export TOPIC_RESULT=None
# Act
status=0
failure=$(bash -ec "$staging_deploy" 2>&1) || status=$?
# Assert
[ "$status" -eq 1 ] && [ "$failure" = 'Missing or invalid staging events topic export' ] ||
  fail 'staging must stop before SAM when the events topic export is missing'

echo "Workflow contract OK"
