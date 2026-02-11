package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	playcamp "github.com/playcamp/playcamp-go-sdk"
)

// handleSearchCreators handles GET /api/creators/search
func (a *app) handleSearchCreators(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))

	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		writeError(w, http.StatusBadRequest, "keyword query parameter is required")
		return
	}

	params := playcamp.SearchCreatorsParams{Keyword: keyword}

	if campaignID := r.URL.Query().Get("campaignId"); campaignID != "" {
		params.CampaignID = playcamp.String(campaignID)
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		params.Limit = playcamp.Int(parsePositiveInt(limitStr, 20))
	}

	creators, err := sdk.Creators.Search(r.Context(), params)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, creators)
}

// handleGetCreator handles GET /api/creators/{key}
func (a *app) handleGetCreator(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))
	key := chi.URLParam(r, "key")

	creator, err := sdk.Creators.Get(r.Context(), key)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, creator)
}

// handleGetCreatorCoupons handles GET /api/creators/{key}/coupons
func (a *app) handleGetCreatorCoupons(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))
	key := chi.URLParam(r, "key")

	coupons, err := sdk.Creators.GetCoupons(r.Context(), key)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, coupons)
}
