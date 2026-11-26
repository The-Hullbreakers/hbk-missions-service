package lambdahttp

import (
	"context"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

// NewProxy adapts HTTP API Gateway v2 events to the service's HTTP router.
func NewProxy(handler http.Handler) func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	proxy := httpadapter.NewV2(handler)
	return func(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		stage := event.RequestContext.Stage
		if stage != "" && stage != "$default" {
			path := event.RawPath
			if path == "" {
				path = event.RequestContext.HTTP.Path
			}
			prefix := "/" + stage
			switch {
			case path == prefix:
				event.RawPath = "/"
			case strings.HasPrefix(path, prefix+"/"):
				event.RawPath = strings.TrimPrefix(path, prefix)
			}
		}
		return proxy.ProxyWithContext(ctx, event)
	}
}
