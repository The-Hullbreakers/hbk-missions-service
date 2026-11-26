package composition

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/application"

	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/membership"
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
	meritStore := missiondynamo.NewMeritLedgerStore(cfg.DynamoDB, cfg.TableName)
	overallCredits := missiondynamo.NewOverallCreditsReader(cfg.DynamoDB, cfg.TableName)
	overallStandingsApp := application.NewGetOverallStandingsHandler(seasonStore, overallCredits)

	policyStore := missiondynamo.NewPolicyStore(cfg.DynamoDB, cfg.TableName)

	maxBytes := cfg.EvidenceMaxBytes
	if maxBytes <= 0 {
		maxBytes = 50 << 20
	}

	confirmApp := application.NewConfirmObjectiveHandlerWithEvidence(contractStore, contractStore, seasonStore, cfg.Evidence)
	uploadApp := application.NewRequestEvidenceUploadHandler(contractStore, cfg.Evidence, maxBytes)
	readApp := application.NewRequestEvidenceReadHandler(contractStore, cfg.Evidence)
	publishApp := application.NewPublishMissionHandler(offerStore, definitionStore, seasonStore)
	var membershipReader application.TeamMembershipReader = membership.DenyAllReader{}
	if cfg.PartyManagerURL != "" {
		membershipReader = membership.NewPartyManagerReader(cfg.PartyManagerURL)
	}
	acceptApp := application.NewAcceptMissionHandler(policyStore, offerStore, contractStore, contractStore, definitionStore, seasonStore, contractStore, membershipReader)
	updateLimitApp := application.NewUpdateActiveLimitHandler(policyStore)
	startApp := application.NewStartContractHandler(contractStore, contractStore)
	submitApp := application.NewSubmitContractHandler(contractStore, contractStore)
	approveApp := application.NewApproveContractHandler(contractStore, contractStore)
	rejectApp := application.NewRejectContractHandler(contractStore, contractStore)
	abandonApp := application.NewAbandonContractHandler(contractStore, contractStore)
	leaveApp := application.NewLeaveContractHandler(contractStore, contractStore)
	confirmInviteApp := application.NewConfirmInvitationHandler(contractStore, contractStore)
	declineInviteApp := application.NewDeclineInvitationHandler(contractStore, contractStore)
	createSeasonApp := application.NewCreateSeasonHandler(seasonStore, seasonStore)
	createScheduledSeasonApp := application.NewCreateScheduledSeasonHandler(seasonStore, seasonStore)
	getOfferApp := application.NewGetOfferHandler(offerStore)
	currentSeasonApp := application.NewGetCurrentSeasonHandler(seasonStore)
	myStandingApp := application.NewGetMyStandingHandler(seasonStore, meritStore)
	historyApp := application.NewListStandingsHistoryHandler(seasonStore, meritStore, meritStore)
	statisticsApp := application.NewGetSeasonStatisticsHandler(seasonStore, seasonStore, meritStore, meritStore)
	listSeasonsApp := application.NewListSeasonsHandler(seasonStore)
	listLeaderboardApp := application.NewListLeaderboardHandler(seasonStore, meritStore)
	freezeApp := application.NewFreezeSeasonHandler(seasonStore, seasonStore)
	adjudicationLock := missiondynamo.NewAdjudicationLock(cfg.DynamoDB, cfg.TableName)
	snapshotReader := missiondynamo.NewAdjudicationSnapshotReader(cfg.DynamoDB, cfg.TableName)
	creditedRosters := missiondynamo.NewCreditedRosterReader(cfg.DynamoDB, cfg.TableName)
	factStager := missiondynamo.NewPlacementFactStager(cfg.DynamoDB, cfg.TableName)
	factVerifier := missiondynamo.NewStagedFactVerifier(cfg.DynamoDB, cfg.TableName)
	committer := missiondynamo.NewFinalizationCommitter(cfg.DynamoDB, cfg.TableName)
	finalizeApp := application.NewSeasonFinalizer(seasonStore, adjudicationLock, snapshotReader, creditedRosters, factStager, factVerifier, committer)
	createDefinitionApp := application.NewCreateMissionDefinitionHandler(definitionStore)
	publishDefinitionApp := application.NewPublishMissionDefinitionHandler(definitionStore, definitionStore)
	listDefinitionsApp := application.NewListDefinitionsHandler(definitionStore)
	listOffersApp := application.NewListOffersHandler(offerStore, contractStore, seasonStore)
	offerFacetsApp := application.NewListOfferFacetsHandler(offerStore, seasonStore)
	listContractsApp := application.NewListPlayerContractsHandler(contractStore)
	reviewQueueApp := application.NewListPendingReviewHandler(contractStore)
	contractDetailApp := application.NewGetContractHandler(contractStore)
	raiseDisputeApp := application.NewRaiseDisputeHandler(contractStore, contractStore)
	resolveDisputeApp := application.NewResolveDisputeHandler(contractStore, contractStore)
	listDisputesApp := application.NewListDisputesHandler(contractStore)

	var outbox application.PendingIntentDrainer
	if cfg.DomainEvents != nil {
		outbox = application.NewPlacementFactPublisher(seasonStore, seasonStore, cfg.DomainEvents, seasonStore, seasonStore)
	}
	scheduler := application.NewSeasonScheduler(seasonStore, application.NewRealClock(), freezeApp, finalizeApp, createScheduledSeasonApp, outbox)
	scheduler.WithOrphanSweep(adjudicationLock, orphanLogReporter{}, 24*time.Hour)
	scheduler.Start(cfg.ShutdownContext, cfg.SchedulerTick, func(err error) {
		log.Printf("season scheduler tick failed: %v", err)
	})

	auth := missionhttp.NewAuthMiddleware(cfg.UserSession)
	adminOnly := missionhttp.RequireRole(cfg.UserSession, "admin")
	adminOrModerator := missionhttp.RequireRole(cfg.UserSession, "admin", "moderator")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealth)
	mux.HandleFunc("GET /readyz", handleReady(cfg.DynamoDB, cfg.TableName))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/evidence", auth.Authenticate(missionhttp.ConfirmEvidenceHandler(confirmApp)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/evidence/upload-url", auth.Authenticate(missionhttp.RequestEvidenceUploadHandler(uploadApp)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/evidence/media-url", auth.Authenticate(missionhttp.RequestEvidenceReadHandler(readApp)))
	mux.Handle("POST /api/v1/missions/definitions", auth.Authenticate(adminOrModerator(missionhttp.CreateMissionDefinitionHandler(createDefinitionApp))))
	mux.Handle("POST /api/v1/missions/definitions/{definition_code}/publish", auth.Authenticate(adminOrModerator(missionhttp.PublishMissionDefinitionHandler(publishDefinitionApp))))
	mux.Handle("GET /api/v1/missions/definitions", auth.Authenticate(adminOrModerator(missionhttp.ListDefinitionsHandler(listDefinitionsApp))))
	mux.Handle("POST /api/v1/missions/offers", auth.Authenticate(adminOrModerator(missionhttp.PublishMissionHandler(publishApp))))
	mux.Handle("POST /api/v1/missions/offers/{offer_id}/accept", auth.Authenticate(missionhttp.AcceptMissionHandler(acceptApp)))
	mux.Handle("PUT /api/v1/missions/policies/active-contract-limit", auth.Authenticate(adminOrModerator(missionhttp.UpdateActiveLimitHandler(updateLimitApp))))
	mux.Handle("GET /api/v1/missions/offers", auth.Authenticate(missionhttp.ListOffersHandler(listOffersApp)))
	mux.Handle("GET /api/v1/missions/offers/facets", auth.Authenticate(missionhttp.ListOfferFacetsHandler(offerFacetsApp)))
	mux.Handle("GET /api/v1/missions/offers/{offer_id}", auth.Authenticate(missionhttp.GetOfferHandler(getOfferApp)))
	mux.Handle("GET /api/v1/missions/contracts", auth.Authenticate(missionhttp.ListContractsHandler(listContractsApp)))
	mux.Handle("GET /api/v1/missions/review-queue", auth.Authenticate(adminOrModerator(missionhttp.ReviewQueueHandler(reviewQueueApp))))
	mux.Handle("GET /api/v1/missions/contracts/{contract_id}", auth.Authenticate(missionhttp.ContractDetailHandler(contractDetailApp)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/start", auth.Authenticate(missionhttp.StartContractHandler(startApp)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/submit", auth.Authenticate(missionhttp.SubmitContractHandler(submitApp)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/approve", auth.Authenticate(adminOrModerator(missionhttp.ApproveContractHandler(approveApp))))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/reject", auth.Authenticate(adminOrModerator(missionhttp.RejectContractHandler(rejectApp))))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/abandon", auth.Authenticate(missionhttp.AbandonContractHandler(abandonApp)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/leave", auth.Authenticate(missionhttp.LeaveContractHandler(leaveApp)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/disputes", auth.Authenticate(missionhttp.RaiseDisputeHandler(raiseDisputeApp)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/disputes/{dispute_id}/resolve", auth.Authenticate(adminOrModerator(missionhttp.ResolveDisputeHandler(resolveDisputeApp))))
	mux.Handle("GET /api/v1/missions/contracts/{contract_id}/disputes", auth.Authenticate(adminOrModerator(missionhttp.ListDisputesHandler(listDisputesApp))))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/confirm-invitation", auth.Authenticate(missionhttp.ConfirmInvitationHandler(confirmInviteApp)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/decline-invitation", auth.Authenticate(missionhttp.DeclineInvitationHandler(declineInviteApp)))
	mux.Handle("GET /api/v1/seasons", auth.Authenticate(missionhttp.ListSeasonsHandler(listSeasonsApp)))
	mux.Handle("GET /api/v1/missions/seasons/current", auth.Authenticate(missionhttp.GetCurrentSeasonHandler(currentSeasonApp)))
	mux.Handle("POST /api/v1/seasons", auth.Authenticate(adminOnly(missionhttp.CreateSeasonHandler(createSeasonApp, createScheduledSeasonApp))))
	mux.Handle("GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}", auth.Authenticate(missionhttp.LeaderboardHandler(listLeaderboardApp)))
	mux.Handle("GET /api/v1/missions/seasons/{season_id}/leaderboards/{category}/me", auth.Authenticate(missionhttp.GetMyStandingHandler(myStandingApp)))
	mux.Handle("GET /api/v1/missions/seasons/{season_id}/standings/overall", auth.Authenticate(missionhttp.OverallStandingsHandler(overallStandingsApp)))
	mux.Handle("GET /api/v1/missions/seasons/{season_id}/standings/history", auth.Authenticate(missionhttp.StandingsHistoryHandler(historyApp)))
	mux.Handle("GET /api/v1/missions/seasons/{season_id}/statistics", auth.Authenticate(missionhttp.SeasonStatisticsHandler(statisticsApp)))
	mux.Handle("POST /api/v1/missions/seasons/{id}/freeze", auth.Authenticate(adminOnly(missionhttp.FreezeSeasonHandler(freezeApp))))
	mux.Handle("POST /api/v1/missions/seasons/{id}/finalize", auth.Authenticate(adminOnly(missionhttp.FinalizeSeasonHandler(finalizeApp))))

	return mux, nil
}
