package composition

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// NewServer wires the application and returns an HTTP handler.
func NewServer(cfg Config) (http.Handler, error) {
	ctx := context.Background()

	if err := cfg.DynamoDB.CreateMissionsTable(ctx, cfg.TableName); err != nil {
		return nil, fmt.Errorf("failed to create missions table: %w", err)
	}

	contractStore := missiondynamo.NewContractStore(cfg.DynamoDB, cfg.TableName)
	offerStore := missiondynamo.NewOfferStore(cfg.DynamoDB, cfg.TableName)
	definitionStore := missiondynamo.NewDefinitionStore(cfg.DynamoDB, cfg.TableName)
	seasonStore := missiondynamo.NewSeasonStore(cfg.DynamoDB, cfg.TableName)
	objectiveStore := missiondynamo.NewObjectiveStore(cfg.DynamoDB, cfg.TableName)
	meritStore := missiondynamo.NewMeritLedgerStore(cfg.DynamoDB, cfg.TableName)

	policy := domain.NewActiveLimitPolicy(5, nil)

	confirmApp := application.NewConfirmObjectiveHandler(
		contractStore,
		contractStore,
		seasonStore,
		objectiveStore,
		meritStore,
	)
	publishApp := application.NewPublishMissionHandler(offerStore, definitionStore, seasonStore)
	acceptApp := application.NewAcceptMissionHandler(policy, offerStore, contractStore, contractStore)
	createSeasonApp := application.NewCreateSeasonHandler(seasonStore, seasonStore)
	createScheduledSeasonApp := application.NewCreateScheduledSeasonHandler(seasonStore, seasonStore)
	listSeasonsApp := application.NewListSeasonsHandler(seasonStore)
	listLeaderboardApp := application.NewListLeaderboardHandler(seasonStore, meritStore)
	freezeApp := application.NewFreezeSeasonHandler(seasonStore, seasonStore)
	finalizeApp := application.NewFinalizeSeasonHandler(seasonStore, seasonStore)
	createDefinitionApp := application.NewCreateMissionDefinitionHandler(definitionStore)
	publishDefinitionApp := application.NewPublishMissionDefinitionHandler(definitionStore, definitionStore)

	scheduler := application.NewSeasonScheduler(seasonStore, application.NewRealClock(), freezeApp, finalizeApp, createScheduledSeasonApp)
	scheduler.Start(cfg.ShutdownContext, cfg.SchedulerTick)

	auth := missionhttp.NewAuthMiddleware(cfg.UserSession)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /readyz", handleReady(cfg.DynamoDB, cfg.TableName))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/evidence", auth.Authenticate(missionhttp.ConfirmEvidenceHandler(confirmApp)))
	mux.Handle("POST /api/v1/missions/definitions", auth.Authenticate(missionhttp.CreateMissionDefinitionHandler(createDefinitionApp)))
	mux.Handle("POST /api/v1/missions/definitions/{definition_id}/publish", auth.Authenticate(missionhttp.PublishMissionDefinitionHandler(publishDefinitionApp)))
	mux.Handle("POST /api/v1/missions/offers", auth.Authenticate(missionhttp.PublishMissionHandler(publishApp)))
	mux.Handle("POST /api/v1/missions/offers/{offer_id}/accept", auth.Authenticate(missionhttp.AcceptMissionHandler(acceptApp)))
	mux.Handle("GET /api/v1/seasons", auth.Authenticate(missionhttp.ListSeasonsHandler(listSeasonsApp)))
	mux.Handle("POST /api/v1/seasons", auth.Authenticate(missionhttp.CreateSeasonHandler(createSeasonApp, createScheduledSeasonApp)))
	mux.Handle("GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}", auth.Authenticate(missionhttp.LeaderboardHandler(listLeaderboardApp)))
	mux.Handle("POST /api/v1/missions/seasons/{id}/freeze", auth.Authenticate(missionhttp.FreezeSeasonHandler(freezeApp)))
	mux.Handle("POST /api/v1/missions/seasons/{id}/finalize", auth.Authenticate(missionhttp.FinalizeSeasonHandler(finalizeApp)))

	return mux, nil
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func handleReady(client *missiondynamo.DynamoDBClient, tableName string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, err := client.Client().DescribeTable(context.Background(), &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "DYNAMODB_UNAVAILABLE"})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	}
}
