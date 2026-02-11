package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	playcamp "github.com/playcamp/playcamp-go-sdk"
)

// handleListCampaigns handles GET /api/campaigns
func (a *app) handleListCampaigns(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))

	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 20)

	result, err := sdk.Campaigns.List(r.Context(), &playcamp.PaginationOptions{
		Page:  playcamp.Int(page),
		Limit: playcamp.Int(limit),
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// handleGetCampaign handles GET /api/campaigns/{id}
func (a *app) handleGetCampaign(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))
	id := chi.URLParam(r, "id")

	campaign, err := sdk.Campaigns.Get(r.Context(), id)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

// handleGetCampaignCreators handles GET /api/campaigns/{id}/creators
func (a *app) handleGetCampaignCreators(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))
	id := chi.URLParam(r, "id")

	creators, err := sdk.Campaigns.GetCreators(r.Context(), id)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, creators)
}
