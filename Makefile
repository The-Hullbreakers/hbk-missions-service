.PHONY: build-MissionsAPIFunction build clean unit-tests integration-tests

GOOS ?= linux
GOARCH ?= amd64
ARTIFACTS_DIR ?= .aws-sam/build/MissionsAPIFunction

build: build-MissionsAPIFunction

build-MissionsAPIFunction:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="-s -w" -o $(ARTIFACTS_DIR)/bootstrap cmd/lambda/main.go

clean:
	rm -rf .aws-sam/

unit-tests:
	go test ./tests/unit/...

integration-tests:
	go test -p=1 ./tests/integration/...

.PHONY: seed-tests seed-integration-tests

seed-tests:
	python3 -m unittest discover -s tests/seeds -p 'test_seed_missions.py' -v
	bash tools/test-workflow-contract.sh

seed-integration-tests:
	@test -n "$(DYNAMODB_SEED_TEST_ENDPOINT)" || (echo 'Set DYNAMODB_SEED_TEST_ENDPOINT to DynamoDB Local'; exit 1)
	DYNAMODB_SEED_TEST_ENDPOINT="$(DYNAMODB_SEED_TEST_ENDPOINT)" python3 -m unittest discover -s tests/seeds -p 'test_seed_integration.py' -v
