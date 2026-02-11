package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	playcamp "github.com/playcamp/playcamp-go-sdk"
)

// handleGetSponsor handles GET /api/sponsors/{userId}
func (a *app) handleGetSponsor(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))
	userID := chi.URLParam(r, "userId")

	sponsors, err := sdk.Sponsors.GetByUser(r.Context(), userID)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sponsors)
}

// handleCreateSponsor handles POST /api/sponsors
func (a *app) handleCreateSponsor(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID     string  `json:"userId"`
		CreatorKey string  `json:"creatorKey"`
		CampaignID *string `json:"campaignId,omitempty"`
		IsTest     *bool   `json:"isTest,omitempty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	isTest := body.IsTest != nil && *body.IsTest
	sdk := a.getSDK(isTest)

	sponsor, err := sdk.Sponsors.Create(r.Context(), playcamp.CreateSponsorParams{
		UserID:     body.UserID,
		CreatorKey: body.CreatorKey,
		CampaignID: body.CampaignID,
		IsTest:     body.IsTest,
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sponsor)
}

// handleUpdateSponsor handles PUT /api/sponsors/{userId}
func (a *app) handleUpdateSponsor(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	var body struct {
		CampaignID    *string `json:"campaignId,omitempty"`
		NewCreatorKey string  `json:"newCreatorKey"`
		IsTest        *bool   `json:"isTest,omitempty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	isTest := body.IsTest != nil && *body.IsTest
	sdk := a.getSDK(isTest)

	sponsor, err := sdk.Sponsors.Update(r.Context(), userID, playcamp.UpdateSponsorParams{
		CampaignID:    body.CampaignID,
		NewCreatorKey: body.NewCreatorKey,
		IsTest:        body.IsTest,
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sponsor)
}

// handleDeleteSponsor handles DELETE /api/sponsors/{userId}
func (a *app) handleDeleteSponsor(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))
	userID := chi.URLParam(r, "userId")

	var opts *playcamp.DeleteSponsorOptions
	if campaignID := r.URL.Query().Get("campaignId"); campaignID != "" {
		opts = &playcamp.DeleteSponsorOptions{CampaignID: playcamp.String(campaignID)}
	}

	if err := sdk.Sponsors.Delete(r.Context(), userID, opts); err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// handleGetSponsorHistory handles GET /api/sponsors/{userId}/history
func (a *app) handleGetSponsorHistory(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))
	userID := chi.URLParam(r, "userId")

	opts := &playcamp.GetSponsorHistoryOptions{
		Page:  playcamp.Int(parsePositiveInt(r.URL.Query().Get("page"), 1)),
		Limit: playcamp.Int(parsePositiveInt(r.URL.Query().Get("limit"), 20)),
	}
	if campaignID := r.URL.Query().Get("campaignId"); campaignID != "" {
		opts.CampaignID = playcamp.String(campaignID)
	}

	result, err := sdk.Sponsors.GetHistory(r.Context(), userID, opts)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
