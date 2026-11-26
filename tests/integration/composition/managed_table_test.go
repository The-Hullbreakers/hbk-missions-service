package composition_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/stretchr/testify/require"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// TEST-STG-MANAGED-TABLE-001
func TestGivenManagedTableWhenSeasonRequestArrivesThenLambdaStartsWithoutProvisioning(t *testing.T) {
	// Arrange
	var calls atomic.Int32
	dynamo := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/x-amz-json-1.0")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"__type":"com.amazonaws.dynamodb.v20120810#AccessDeniedException","message":"table provisioning is not permitted"}`))
	}))
	t.Cleanup(dynamo.Close)
	db, err := missiondynamo.NewDynamoDBClient(t.Context(), "us-east-2", dynamo.URL)
	require.NoError(t, err)
	cfg := missioncomposition.Config{TableName: "hbk-missions-stg", DynamoDB: db, UseExistingTable: true}
	req := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/missions/seasons/current", RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet, Path: "/api/v1/missions/seasons/current"}}}
	handler, initErr := missioncomposition.NewServer(cfg)
	require.NoError(t, initErr)
	adapter := httpadapter.NewV2(handler)
	// Act
	resp, err := adapter.ProxyWithContext(t.Context(), req)
	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Zero(t, calls.Load())
}
