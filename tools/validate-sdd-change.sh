#!/usr/bin/env bash
set -euo pipefail

# validate-sdd-change.sh verifies the minimum quality gates for an SDD change.

cd "$(dirname "$0")/.."

COMPOSE_FILE="docker-compose.test.yml"
DYNAMODB_TEST_PORT="${DYNAMODB_TEST_PORT:-18000}"
export DYNAMODB_TEST_ENDPOINT="${DYNAMODB_TEST_ENDPOINT:-http://localhost:${DYNAMODB_TEST_PORT}}"

start_dynamodb() {
	if ! command -v docker >/dev/null 2>&1; then
		return
	fi
	if ! docker info >/dev/null 2>&1; then
		return
	fi
	if [ ! -f "$COMPOSE_FILE" ]; then
		return
	fi

	docker compose -f "$COMPOSE_FILE" down --remove-orphans 2>/dev/null || true
	DYNAMODB_TEST_PORT="$DYNAMODB_TEST_PORT" docker compose -f "$COMPOSE_FILE" up -d
	trap 'docker compose -f "$COMPOSE_FILE" down --remove-orphans 2>/dev/null || true' EXIT

	for _ in $(seq 1 30); do
		if curl --silent --max-time 1 "$DYNAMODB_TEST_ENDPOINT" >/dev/null 2>&1; then
			return
		fi
		sleep 1
	done

	echo "DynamoDB test container did not become ready at $DYNAMODB_TEST_ENDPOINT" >&2
	exit 1
}

start_dynamodb

echo "=== gofmt ==="
gofmt -l .

echo "=== go vet ==="
go vet ./...

echo "=== unit tests ==="
go test ./tests/unit/...

echo "=== integration tests ==="
go test -p=1 ./tests/integration/...

echo "=== race detector ==="
go test -race -p=1 ./tests/unit/... ./tests/integration/...

echo "=== coverage (domain, application, affected infrastructure) ==="
go test -coverpkg=./internal/missions/domain -coverprofile=coverage-domain.out ./tests/unit/missions/domain
go test -coverpkg=./internal/missions/application -coverprofile=coverage-app.out ./tests/unit/missions/application
go test -coverpkg=./internal/missions/infrastructure/dynamodb -coverprofile=coverage-infra.out ./tests/integration/infrastructure/dynamodb

echo "=== SDD validation passed ==="
