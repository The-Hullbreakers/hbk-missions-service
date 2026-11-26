//go:build integration

package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TEST-HTTP-LANG-002 (list-contracts partition): explicit lang=es on a
// bilingual contract serves the Spanish baseline honestly — not just the
// omitted-query default.
func TestLangListContractsExplicitSpanish(t *testing.T) {
	// Arrange
	server, err := languageContractsServer(t, true, false)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/missions/contracts?season_id=season-1&status=all&lang=es", nil)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer any")

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
	assert.Equal(t, "contract-rescue", items[0].ContractID)
	assert.Equal(t, "es", items[0].ContentLanguage)
	assert.Equal(t, "Rescate", items[0].Offer.Title)
	require.Len(t, items[0].Terms, 1)
	assert.Equal(t, "Capturar", items[0].Terms[0].Name)
	assert.Equal(t, 10, items[0].Terms[0].MeritPerUnit)
}

// TEST-HTTP-LANG-002 (list-contracts partition): a legacy Spanish-only
// contract serves Spanish under lang=en with an honest label and stable
// economics.
func TestLangListContractsLegacyFallback(t *testing.T) {
	// Arrange
	server, err := languageContractsServer(t, false, true)
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
	assert.Equal(t, "es", response.Header.Get("Content-Language"))
	var items []langContractItem
	require.NoError(t, json.NewDecoder(response.Body).Decode(&items))
	require.Len(t, items, 1)
	assert.Equal(t, "contract-old", items[0].ContractID)
	assert.Equal(t, "es", items[0].ContentLanguage)
	assert.Equal(t, "Rescate antiguo", items[0].Offer.Title)
	require.Len(t, items[0].Terms, 1)
	assert.Equal(t, "Capturar", items[0].Terms[0].Name)
	assert.Equal(t, 10, items[0].Terms[0].MeritPerUnit)
	assert.Equal(t, 2, items[0].Terms[0].UnitLimit)
}

// TEST-HTTP-LANG-002 (list-contracts partition): a mixed roster serves
// each contract in its own language — no header on the heterogeneous
// page — and the outsider contract stays filtered out.
func TestLangListContractsMixedAndFiltered(t *testing.T) {
	// Arrange
	server, err := languageContractsServer(t, true, true)
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
	assert.Equal(t, "", response.Header.Get("Content-Language"))
	var raw []map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(response.Body).Decode(&raw))
	require.Len(t, raw, 2)
	for _, itemRaw := range raw {
		assert.NotContains(t, itemRaw, "ContentEn")
		var offer map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(itemRaw["Offer"], &offer))
		assert.NotContains(t, offer, "ContentEn")
	}
	var items []langContractItem
	require.NoError(t, json.Unmarshal(mustMarshal(t, raw), &items))
	byID := map[string]langContractItem{}
	for _, item := range items {
		byID[item.ContractID] = item
	}
	assert.NotContains(t, byID, "contract-other")
	require.Contains(t, byID, "contract-rescue")
	require.Contains(t, byID, "contract-old")
	assertRescueContractEnglish(t, byID["contract-rescue"])
	assert.Equal(t, "es", byID["contract-old"].ContentLanguage)
	assert.Equal(t, "Rescate antiguo", byID["contract-old"].Offer.Title)
	require.Len(t, byID["contract-old"].Terms, 1)
	assert.Equal(t, "Capturar", byID["contract-old"].Terms[0].Name)
	assert.Equal(t, 10, byID["contract-old"].Terms[0].MeritPerUnit)
}
