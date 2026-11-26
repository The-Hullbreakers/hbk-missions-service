//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type langContractItem struct {
	ContractID      string `json:"ContractID"`
	ContentLanguage string `json:"ContentLanguage"`
	Offer           struct {
		Title        string   `json:"Title"`
		Description  string   `json:"Description"`
		Rules        []string `json:"Rules"`
		MeritSummary struct {
			TotalMerits int `json:"TotalMerits"`
		} `json:"MeritSummary"`
	} `json:"Offer"`
	Terms []struct {
		ObjectiveID  string `json:"ObjectiveID"`
		Name         string `json:"Name"`
		Description  string `json:"Description"`
		MeritPerUnit int    `json:"MeritPerUnit"`
		UnitLimit    int    `json:"UnitLimit"`
	} `json:"Terms"`
}

func assertRescueContractEnglish(t *testing.T, item langContractItem) {
	t.Helper()
	assert.Equal(t, "contract-rescue", item.ContractID)
	assert.Equal(t, "en", item.ContentLanguage)
	assert.Equal(t, "Rescue", item.Offer.Title)
	assert.Equal(t, "Save the pilot", item.Offer.Description)
	assert.Equal(t, []string{"No aids"}, item.Offer.Rules)
	assert.Equal(t, 70, item.Offer.MeritSummary.TotalMerits)
	require.Len(t, item.Terms, 1)
	assert.Equal(t, "obj-1", item.Terms[0].ObjectiveID)
	assert.Equal(t, "Capture", item.Terms[0].Name)
	assert.Equal(t, "Capture the pilot", item.Terms[0].Description)
	assert.Equal(t, 10, item.Terms[0].MeritPerUnit)
	assert.Equal(t, 2, item.Terms[0].UnitLimit)
}

// TEST-HTTP-LANG-002 (list-contracts partition): a bilingual contract
// serves English under lang=en with one contract-level label and the
// standard Content-Language header — ContentEn never exposed.
func TestLangListContractsHomogeneousEnglish(t *testing.T) {
	// Arrange
	server, err := languageContractsServer(t, true, false)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/contracts?season_id=season-1&status=all&lang=en", nil)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer any")

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "en", response.Header.Get("Content-Language"))
	var raw []map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(response.Body).Decode(&raw))
	require.Len(t, raw, 1)
	assert.NotContains(t, raw[0], "ContentEn")
	var offer map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw[0]["Offer"], &offer))
	assert.NotContains(t, offer, "ContentEn")
	var item langContractItem
	require.NoError(t, json.Unmarshal(mustMarshal(t, raw[0]), &item))
	assertRescueContractEnglish(t, item)
}

func mustMarshal(t *testing.T, raw any) []byte {
	t.Helper()
	b, err := json.Marshal(raw)
	require.NoError(t, err)
	return b
}

// TEST-HTTP-LANG-002 (list-contracts partition): omitted lang defaults to
// Spanish and ignores Accept-Language.
func TestLangListContractsSpanishDefaultIgnoresAcceptLanguage(t *testing.T) {
	// Arrange
	server, err := languageContractsServer(t, true, false)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/contracts?season_id=season-1&status=all", nil)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer any")
	request.Header.Set("Accept-Language", "en")

	// Act
	response, err := server.Client().Do(request)

	// Assert
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "es", response.Header.Get("Content-Language"))
	var items []langContractItem
	require.NoError(t, json.NewDecoder(response.Body).Decode(&items))
	require.Len(t, items, 1)
	assert.Equal(t, "es", items[0].ContentLanguage)
	assert.Equal(t, "Rescate", items[0].Offer.Title)
	require.Len(t, items[0].Terms, 1)
	assert.Equal(t, "Capturar", items[0].Terms[0].Name)
}
