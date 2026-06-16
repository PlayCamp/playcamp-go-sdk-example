package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	playcamp "github.com/playcamp/playcamp-go-sdk"
)

// handleWebviewToken handles POST /webview/token
func (a *app) handleWebviewToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID     string `json:"userId"`
		CampaignID string `json:"campaignId,omitempty"`
		CallbackID string `json:"callbackId,omitempty"`
		IsTest     *bool  `json:"isTest,omitempty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.UserID == "" {
		writeError(w, http.StatusBadRequest, "userId is required")
		return
	}

	isTest := body.IsTest != nil && *body.IsTest
	sdk := a.getSDK(isTest)

	result, err := sdk.Webview.CreateOTT(r.Context(), playcamp.WebviewOttParams{
		UserID:     body.UserID,
		CampaignID: body.CampaignID,
		CallbackID: body.CallbackID,
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}

	// Build the full WebView URL from the configured SDK API host so it works
	// regardless of SDK_API_URL (local, sandbox, live, custom).
	base := strings.TrimRight(a.apiBaseURL, "/")
	webviewURL := base + "/webview/?ott=" + url.QueryEscape(result.OTT)
	if body.CampaignID != "" {
		webviewURL += "&tabs=sponsor,coupon"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": result, "webviewUrl": webviewURL})
}
