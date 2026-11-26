package composition

import (
	"context"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

// Config holds the dependencies needed to build the HTTP server.
type Config struct {
	TableName string
	DynamoDB  *missiondynamo.DynamoDBClient
	// UseExistingTable leaves table provisioning to the infrastructure owner.
	UseExistingTable bool
	UserSession      session.UserSession
	ShutdownContext  context.Context
	SchedulerTick    time.Duration
	Evidence         application.EvidenceBlobStore
	EvidenceMaxBytes int64
	// PartyManagerURL enables the real Party Manager membership adapter; an
	// empty value keeps the fail-closed DenyAllReader.
	PartyManagerURL string
	// DomainEvents publishes outbox intents to the shared bus; nil disables
	// the scheduler's outbox drain (e.g. when no topic ARN is configured).
	DomainEvents application.DomainEventPublisher
}
