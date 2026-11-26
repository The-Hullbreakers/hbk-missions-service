// Package sns provides the AWS SNS adapter for the shared domain-events bus.
// The outbox envelope is published verbatim; `event_type` and `source`
// message attributes carry the routing/filter metadata.
package sns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
)

// NewClient creates an SNS client. When endpointURL is non-empty it uses
// static test credentials and the provided endpoint (e.g. localstack or a
// local fake); otherwise the ambient AWS credential chain is used.
func NewClient(ctx context.Context, region, endpointURL string) (*awssns.Client, error) {
	if endpointURL != "" {
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load test config: %w", err)
		}
		return awssns.NewFromConfig(cfg, func(o *awssns.Options) {
			o.BaseEndpoint = aws.String(endpointURL)
		}), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return awssns.NewFromConfig(cfg), nil
}
