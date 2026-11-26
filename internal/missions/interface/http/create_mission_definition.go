package http

import (
	"encoding/json"
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// createMissionDefinitionRequest carries business fields only; caller
// identity is always derived from the authenticated session.
type createMissionDefinitionRequest struct {
	DefinitionCode   string                              `json:"definition_code"`
	Title            string                              `json:"title"`
	Description      string                              `json:"description"`
	Category         string                              `json:"category"`
	Mode             string                              `json:"mode"`
	CompletionReward int                                 `json:"completion_reward"`
	Objectives       []objectiveRuleRequest              `json:"objectives"`
	Rules            []string                            `json:"rules"`
	Translations     createMissionDefinitionTranslations `json:"translations"`
}

// objectiveRuleRequest is one objective rule inside a mission definition.
type objectiveRuleRequest struct {
	ObjectiveID  string `json:"objective_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	MeritPerUnit int    `json:"merit_per_unit"`
	Bonus        int    `json:"bonus"`
	UnitLimit    int    `json:"unit_limit"`
}

// CreateMissionDefinitionHandler returns an HTTP handler for creating mission definitions.
func CreateMissionDefinitionHandler(app *application.CreateMissionDefinitionHandler) nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("POST /api/v1/missions/definitions", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		var req createMissionDefinitionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(nethttp.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "BAD_REQUEST"})
			return
		}

		mode := domain.MissionMode(req.Mode)
		if mode == "" {
			mode = domain.MissionModeSolo
		}

		_, idErr := CallerIdentityFromContext(r.Context())
		if idErr != nil {
			w.WriteHeader(nethttp.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UNAUTHORIZED"})
			return
		}

		objectives := make([]domain.Objective, 0, len(req.Objectives))
		for _, o := range req.Objectives {
			category := o.Category
			if category == "" {
				category = req.Category
			}
			objectives = append(objectives, domain.NewObjectiveNamed(o.ObjectiveID, o.Name, o.Description, category, o.MeritPerUnit, o.Bonus, o.UnitLimit))
		}

		definition, err := app.Handle(application.CreateMissionDefinitionRequest{
			DefinitionCode:   req.DefinitionCode,
			Title:            req.Title,
			Description:      req.Description,
			Category:         req.Category,
			Mode:             mode,
			CompletionReward: req.CompletionReward,
			Objectives:       objectives,
			Rules:            req.Rules,
			ContentEn:        req.Translations.englishContent(),
		})
		if err != nil {
			w.WriteHeader(nethttp.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": PublicCode(err)})
			return
		}

		w.WriteHeader(nethttp.StatusCreated)
		_ = json.NewEncoder(w).Encode(definition)
	})
	return mux
}
