package sns

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
)

// publishClient is the SNS Publish surface the adapter needs.
type publishClient interface {
	Publish(ctx context.Context, input *awssns.PublishInput, optFns ...func(*awssns.Options)) (*awssns.PublishOutput, error)
}

// Publisher implements application.DomainEventPublisher: it sends the
// persisted outbox envelope to the shared topic unchanged, with `event_type`
// and `source` message attributes for subscription-side filtering.
type Publisher struct {
	client   publishClient
	topicARN string
}

// NewPublisher creates a publisher bound to an explicit endpoint (local or
// test wiring) with static credentials.
func NewPublisher(endpointURL, topicARN string) *Publisher {
	client := awssns.NewFromConfig(aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("test", "test", ""),
	}, func(o *awssns.Options) {
		if endpointURL != "" {
			o.BaseEndpoint = aws.String(endpointURL)
		}
	})
	return NewPublisherWithClient(client, topicARN)
}

// NewPublisherWithClient creates a publisher on an existing client, e.g. one
// built from the ambient AWS credential chain in composition.
func NewPublisherWithClient(client *awssns.Client, topicARN string) *Publisher {
	return &Publisher{client: client, topicARN: topicARN}
}

// Publish sends the intent's envelope to the shared topic. A failure is
// returned so the intent stays pending and the next drain retries it with
// the same event id.
func (p *Publisher) Publish(ctx context.Context, intent application.PendingIntent) error {
	_, err := p.client.Publish(ctx, &awssns.PublishInput{
		TopicArn: aws.String(p.topicARN),
		Message:  aws.String(intent.Envelope),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"event_type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(intent.EventType),
			},
			"source": {
				DataType:    aws.String("String"),
				StringValue: aws.String(application.EventSource),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to publish event %s: %w", intent.EventID, err)
	}
	return nil
}
