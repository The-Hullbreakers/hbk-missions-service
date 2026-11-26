//go:build integration

package http_test

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/composition"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

// langAdminSession is a UserSession double admitting an admin caller so the
// language boundary tests exercise the real role middleware path.
type langAdminSession struct{}

func (langAdminSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "admin-1", map[string]any{"roles": []any{"admin"}}, nil
}

func (langAdminSession) HasRole(_ map[string]any, role string) bool {
	return role == "admin"
}

// languageHTTPServer boots the real composition server as the admin caller.
func languageHTTPServer(t *testing.T, seed func(*missiondynamo.DynamoDBClient, string) error) (*httptest.Server, error) {
	t.Helper()
	return languageSessionServer(t, langAdminSession{}, seed)
}

// languageSessionServer boots the real composition server with the given
// session against a private DynamoDB table; seed runs against the same table
// before the server is served. Tables are unique per call — no shared state.
func languageSessionServer(t *testing.T, userSession session.UserSession, seed func(*missiondynamo.DynamoDBClient, string) error) (*httptest.Server, error) {
	t.Helper()
	endpoint := os.Getenv("DYNAMODB_TEST_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:18000"
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "http" || (parsed.Hostname() != "localhost" && parsed.Hostname() != "127.0.0.1") {
		return nil, fmt.Errorf("DYNAMODB_TEST_ENDPOINT must be loopback HTTP")
	}
	client, err := missiondynamo.NewDynamoDBClient(t.Context(), "us-east-1", endpoint)
	if err != nil {
		return nil, err
	}
	table := fmt.Sprintf("hbk-missions-lang-red-%d", time.Now().UnixNano())
	server, err := composition.NewServer(composition.Config{TableName: table, DynamoDB: client, UserSession: userSession})
	if err != nil {
		return nil, err
	}
	if seed != nil {
		if err := seed(client, table); err != nil {
			return nil, err
		}
	}
	ts := httptest.NewServer(server)
	t.Cleanup(ts.Close)
	return ts, nil
}

func languageDefinitionsServer(t *testing.T, seed bool) (*httptest.Server, error) {
	if !seed {
		return languageHTTPServer(t, nil)
	}
	return languageHTTPServer(t, func(client *missiondynamo.DynamoDBClient, table string) error {
		definition, err := domain.NewMissionDefinition("legacy-1", "Rescate", "Salvar al piloto", "BOUNTY_HUNTING", domain.MissionModeSolo, 0, nil, nil)
		if err != nil {
			return err
		}
		return missiondynamo.NewDefinitionStore(client, table).Save(definition)
	})
}

// languageOfferServer seeds one Spanish-only legacy offer; English-success
// coverage stays deferred until persistence maps ContentEn.
func languageOfferServer(t *testing.T) (*httptest.Server, error) {
	return languageHTTPServer(t, func(client *missiondynamo.DynamoDBClient, table string) error {
		offer := domain.NewMissionOffer("offer-legacy", "BOUNTY_HUNTING", "season-1")
		offer.Title = "Rescate"
		offer.Description = "Salvar al piloto"
		offer.CompletionReward = 50
		offer.Objectives = []domain.Objective{
			domain.NewObjectiveNamed("obj-1", "Capturar", "", "BOUNTY_HUNTING", 10, 0, 2),
		}
		return missiondynamo.NewOfferStore(client, table).Save(offer)
	})
}

// languageMixedOffersServer seeds one legacy Spanish-only offer and one
// bilingual offer; persistence now round-trips ContentEn, so the mixed
// page genuinely serves English for the bilingual entry.
func languageMixedOffersServer(t *testing.T) (*httptest.Server, error) {
	return languageHTTPServer(t, func(client *missiondynamo.DynamoDBClient, table string) error {
		store := missiondynamo.NewOfferStore(client, table)
		legacy := domain.NewMissionOffer("offer-legacy", "BOUNTY_HUNTING", "season-1")
		legacy.Title = "Rescate"
		legacy.Description = "Salvar al piloto"
		if err := store.Save(legacy); err != nil {
			return err
		}
		bilingual := domain.NewMissionOffer("offer-bilingual", "BOUNTY_HUNTING", "season-1")
		bilingual.Title = "Caceria"
		bilingual.Description = "Derrota al objetivo"
		bilingual.ContentEn = domain.LocalizedContent{Title: "Hunt", Description: "Defeat the target"}
		return store.Save(bilingual)
	})
}
