package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	playcamp "github.com/playcamp/playcamp-go-sdk"
)

// handleCreatePayment handles POST /api/payments
func (a *app) handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID           string                    `json:"userId"`
		TransactionID    string                    `json:"transactionId"`
		ProductID        string                    `json:"productId"`
		ProductName      *string                   `json:"productName,omitempty"`
		Amount           float64                   `json:"amount"`
		Currency         string                    `json:"currency"`
		Platform         playcamp.PaymentPlatform  `json:"platform"`
		DistributionType *playcamp.DistributionType `json:"distributionType,omitempty"`
		PurchasedAt      *string                   `json:"purchasedAt,omitempty"`
		Receipt          *string                   `json:"receipt,omitempty"`
		CampaignID       *string                   `json:"campaignId,omitempty"`
		CreatorKey       *string                   `json:"creatorKey,omitempty"`
		CallbackID       string                    `json:"callbackId,omitempty"`
		IsTest           *bool                     `json:"isTest,omitempty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	isTest := body.IsTest != nil && *body.IsTest
	sdk := a.getSDK(isTest)

	// Parse purchasedAt or default to now.
	purchasedAt := time.Now().UTC()
	if body.PurchasedAt != nil && *body.PurchasedAt != "" {
		parsed, err := time.Parse(time.RFC3339, *body.PurchasedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid purchasedAt format, expected RFC3339")
			return
		}
		purchasedAt = parsed
	}

	payment, err := sdk.Payments.Create(r.Context(), playcamp.CreatePaymentParams{
		UserID:           body.UserID,
		TransactionID:    body.TransactionID,
		ProductID:        body.ProductID,
		ProductName:      body.ProductName,
		Amount:           body.Amount,
		Currency:         body.Currency,
		Platform:         body.Platform,
		DistributionType: body.DistributionType,
		PurchasedAt:      purchasedAt,
		Receipt:          body.Receipt,
		CampaignID:       body.CampaignID,
		CreatorKey:       body.CreatorKey,
		CallbackID:       body.CallbackID,
		IsTest:           body.IsTest,
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, payment)
}

// handleGetPayment handles GET /api/payments/{transactionId}
func (a *app) handleGetPayment(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))
	txnID := chi.URLParam(r, "transactionId")

	payment, err := sdk.Payments.Get(r.Context(), txnID)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payment)
}

// handleGetUserPayments handles GET /api/payments/user/{userId}
func (a *app) handleGetUserPayments(w http.ResponseWriter, r *http.Request) {
	sdk := a.getSDK(isTestFromQuery(r))
	userID := chi.URLParam(r, "userId")

	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 20)

	result, err := sdk.Payments.ListByUser(r.Context(), userID, &playcamp.PaginationOptions{
		Page:  playcamp.Int(page),
		Limit: playcamp.Int(limit),
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// handleRefundPayment handles POST /api/payments/{transactionId}/refund
func (a *app) handleRefundPayment(w http.ResponseWriter, r *http.Request) {
	txnID := chi.URLParam(r, "transactionId")

	var body struct {
		CallbackID string `json:"callbackId,omitempty"`
		IsTest     *bool  `json:"isTest,omitempty"`
	}
	// Body is optional for refund.
	_ = decodeJSON(r, &body)

	isTest := body.IsTest != nil && *body.IsTest
	sdk := a.getSDK(isTest)

	var opts *playcamp.RefundPaymentOptions
	if body.IsTest != nil || body.CallbackID != "" {
		opts = &playcamp.RefundPaymentOptions{
			CallbackID: body.CallbackID,
			IsTest:     body.IsTest,
		}
	}

	payment, err := sdk.Payments.Refund(r.Context(), txnID, opts)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, payment)
}
