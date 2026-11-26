package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

// NewS3Client creates an S3 client. When endpointURL is non-empty it uses
// static test credentials and the provided endpoint (e.g. localstack).
func NewS3Client(ctx context.Context, region, endpointURL string) (*awss3.Client, error) {
	if endpointURL != "" {
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "testtest", "")),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load test config: %w", err)
		}
		return awss3.NewFromConfig(cfg, func(o *awss3.Options) {
			o.BaseEndpoint = aws.String(endpointURL)
			o.UsePathStyle = true
		}), nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	return awss3.NewFromConfig(cfg), nil
}
