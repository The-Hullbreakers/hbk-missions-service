package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	missions3 "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/s3"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
	missionsns "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/sns"
)

func main() {
	handler, err := newHandler()
	if err != nil {
		panic("failed to initialize lambda handler: " + err.Error())
	}

	lambda.Start(httpadapter.NewV2(handler).ProxyWithContext)
}

func newHandler() (http.Handler, error) {
	ctx := context.Background()

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-2"
	}

	endpoint := os.Getenv("DYNAMODB_ENDPOINT_URL")
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		tableName = "hbk-missions"
	}

	db, err := missiondynamo.NewDynamoDBClient(ctx, region, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to create DynamoDB client: %w", err)
	}

	var audiences []string
	if v := os.Getenv("OIDC_ALLOWED_AUDIENCES"); v != "" {
		audiences = strings.Split(v, ",")
	}
	userSession := session.NewJWTUserSession(os.Getenv("OIDC_ISSUER"), os.Getenv("OIDC_CLIENT_ID"), audiences)

	var evidence application.EvidenceBlobStore
	if bucket := os.Getenv("EVIDENCE_BUCKET_NAME"); bucket != "" {
		s3Client, err := missions3.NewS3Client(ctx, region, os.Getenv("S3_ENDPOINT_URL"))
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 client: %w", err)
		}
		evidence = missions3.NewEvidenceStore(bucket, s3Client)
	}

	var domainEvents application.DomainEventPublisher
	if topicARN := os.Getenv("SNS_TOPIC_ARN"); topicARN != "" {
		snsClient, err := missionsns.NewClient(ctx, region, os.Getenv("SNS_ENDPOINT_URL"))
		if err != nil {
			return nil, fmt.Errorf("failed to create SNS client: %w", err)
		}
		domainEvents = missionsns.NewPublisherWithClient(snsClient, topicARN)
	}

	maxBytes, _ := strconv.ParseInt(os.Getenv("EVIDENCE_MAX_BYTES"), 10, 64)

	return missioncomposition.NewServer(missioncomposition.Config{
		TableName:        tableName,
		DynamoDB:         db,
		UserSession:      userSession,
		ShutdownContext:  ctx,
		SchedulerTick:    time.Minute,
		Evidence:         evidence,
		EvidenceMaxBytes: maxBytes,
		PartyManagerURL:  os.Getenv("PARTY_MANAGER_URL"),
		DomainEvents:     domainEvents,
	})
}
