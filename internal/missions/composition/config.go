package composition

import (
	"context"
	"time"

	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

// Config holds the dependencies needed to build the HTTP server.
type Config struct {
	TableName       string
	DynamoDB        *missiondynamo.DynamoDBClient
	UserSession     session.UserSession
	ShutdownContext context.Context
	SchedulerTick   time.Duration
}
