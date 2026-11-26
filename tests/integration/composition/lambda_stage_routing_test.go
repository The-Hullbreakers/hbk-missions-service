package composition_test

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	missionlambda "github.com/hbk/hbk-missions-service/internal/missions/interface/lambdahttp"
)

// TEST-LAMBDA-STAGE-ROUTING-001
func TestGivenGatewayStageWhenInvokingSeasonRouteThenOnlyExactPrefixIsNormalized(t *testing.T) {
	// Arrange
	handler, err := missioncomposition.NewServer(missioncomposition.Config{UseExistingTable: true})
	require.NoError(t, err)
	proxy := missionlambda.NewProxy(handler)
	const route = "/api/v1/missions/seasons/current"
	for _, tc := range []struct {
		name, stage, rawPath, httpPath string
		want                           int
	}{
		{"named stage prefix", "stg", "/stg" + route, "/stg" + route, http.StatusUnauthorized},
		{"clean path", "stg", route, route, http.StatusUnauthorized},
		{"fallback path", "stg", "", "/stg" + route, http.StatusUnauthorized},
		{"similar prefix is not a stage", "stg", "/stg-other" + route, "/stg-other" + route, http.StatusNotFound},
		{"default stage", "$default", route, route, http.StatusUnauthorized},
		{"other stage is not stripped", "stg", "/prod" + route, "/prod" + route, http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			event := events.APIGatewayV2HTTPRequest{Version: "2.0", RawPath: tc.rawPath, RequestContext: events.APIGatewayV2HTTPRequestContext{Stage: tc.stage, DomainName: "stg-api.thehullbreakers.com", HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet, Path: tc.httpPath}}}
			// Act
			response, err := proxy(t.Context(), event)
			// Assert
			require.NoError(t, err)
			require.Equal(t, tc.want, response.StatusCode)
		})
	}
}
