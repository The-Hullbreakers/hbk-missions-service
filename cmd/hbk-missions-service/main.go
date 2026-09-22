package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	missioncomposition "github.com/hbk/hbk-missions-service/internal/missions/composition"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	tableName := os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "hbk-missions"
	}

	issuer := os.Getenv("OIDC_ISSUER")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	var audiences []string
	if v := os.Getenv("OIDC_ALLOWED_AUDIENCES"); v != "" {
		audiences = strings.Split(v, ",")
	}

	ctx := context.Background()
	db, err := missiondynamo.NewDynamoDBClient(ctx, region, endpoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create DynamoDB client: %v\n", err)
		os.Exit(1)
	}

	userSession := session.NewJWTUserSession(issuer, clientID, audiences)

	schedulerTick := time.Minute
	if v := os.Getenv("SCHEDULER_TICK_MINUTES"); v != "" {
		if minutes, err := strconv.Atoi(v); err == nil && minutes > 0 {
			schedulerTick = time.Duration(minutes) * time.Minute
		}
	}

	server, err := missioncomposition.NewServer(missioncomposition.Config{
		TableName:       tableName,
		DynamoDB:        db,
		UserSession:     userSession,
		ShutdownContext: ctx,
		SchedulerTick:   schedulerTick,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create server: %v\n", err)
		os.Exit(1)
	}

	addr := ":" + port
	fmt.Printf("Listening on %s\n", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		fmt.Fprintf(os.Stderr, "server failed: %v\n", err)
		os.Exit(1)
	}
}
