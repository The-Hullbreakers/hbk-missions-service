package dynamodb_test

import "os"

func dynamodbTestEndpoint() string {
	if endpoint := os.Getenv("DYNAMODB_TEST_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	return "http://localhost:18000"
}
