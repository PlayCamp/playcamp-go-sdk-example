package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	playcamp "github.com/playcamp/playcamp-go-sdk"
)

// handleValidateCoupon handles POST /api/coupons/validate
func (a *app) handleValidateCoupon(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CouponCode string `json:"couponCode"`
		UserID     string `json:"userId"`
		IsTest     *bool  `json:"isTest,omitempty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	isTest := body.IsTest != nil && *body.IsTest
	sdk := a.getSDK(isTest)

	result, err := sdk.Coupons.Validate(r.Context(), playcamp.ValidateCouponServerParams{
		CouponCode: body.CouponCode,
		UserID:     body.UserID,
		IsTest:     body.IsTest,
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// handleRedeemCoupon handles POST /api/coupons/redeem
func (a *app) handleRedeemCoupon(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CouponCode   string  `json:"couponCode"`
		UserID       string  `json:"userId"`
		GameUserUUID *string `json:"gameUserUuid,omitempty"`
		CallbackID   string  `json:"callbackId,omitempty"`
		IsTest       *bool   `json:"isTest,omitempty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	isTest := body.IsTest != nil && *body.IsTest
	sdk := a.getSDK(isTest)

	result, err := sdk.Coupons.Redeem(r.Context(), playcamp.RedeemCouponParams{
		CouponCode:   body.CouponCode,
		UserID:       body.UserID,
		GameUserUUID: body.GameUserUUID,
		CallbackID:   body.CallbackID,
		IsTest:       body.IsTest,
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// handleGetCouponHistory handles GET /api/coupons/user/{userId}
func (a *app) handleGetCouponHistory(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))
	userID := chi.URLParam(r, "userId")

	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 20)

	result, err := sdk.Coupons.GetUserHistory(r.Context(), userID, &playcamp.PaginationOptions{
		Page:  playcamp.Int(page),
		Limit: playcamp.Int(limit),
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
