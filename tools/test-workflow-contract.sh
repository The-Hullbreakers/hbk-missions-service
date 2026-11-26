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

# No deploy secrets on pull requests is enforced by if-conditions on push/dispatch
require "github.ref == 'refs/heads/staging'"
require "github.ref == 'refs/heads/main'"

echo "Workflow contract OK"
