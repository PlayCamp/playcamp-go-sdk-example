package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	playcamp "github.com/playcamp/playcamp-go-sdk"
)

// handleCreatePlaytimeSession handles POST /api/playtime-sessions
func (a *app) handleCreatePlaytimeSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SessionID       string                 `json:"sessionId"`
		UserID          string                 `json:"userId"`
		DurationSeconds int                    `json:"durationSeconds"`
		StartedAt       string                 `json:"startedAt"`
		EndedAt         string                 `json:"endedAt"`
		Metadata        map[string]interface{} `json:"metadata,omitempty"`
		CallbackID      string                 `json:"callbackId,omitempty"`
		IsTest          *bool                  `json:"isTest,omitempty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	startedAt, err := time.Parse(time.RFC3339, body.StartedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid startedAt format, expected RFC3339")
		return
	}
	endedAt, err := time.Parse(time.RFC3339, body.EndedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid endedAt format, expected RFC3339")
		return
	}

	isTest := body.IsTest != nil && *body.IsTest
	sdk := a.getSDK(isTest)

	session, err := sdk.PlaytimeSessions.Create(r.Context(), playcamp.CreatePlaytimeSessionParams{
		SessionID:       body.SessionID,
		UserID:          body.UserID,
		DurationSeconds: body.DurationSeconds,
		StartedAt:       startedAt,
		EndedAt:         endedAt,
		Metadata:        body.Metadata,
		CallbackID:      body.CallbackID,
		IsTest:          body.IsTest,
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

// handleCreateBulkPlaytimeSession handles POST /api/playtime-sessions/bulk
func (a *app) handleCreateBulkPlaytimeSession(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Sessions   []json.RawMessage `json:"sessions"`
		CallbackID string            `json:"callbackId,omitempty"`
		IsTest     *bool             `json:"isTest,omitempty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if len(body.Sessions) == 0 {
		writeError(w, http.StatusBadRequest, "sessions array is required and must not be empty")
		return
	}

	isTest := body.IsTest != nil && *body.IsTest
	sdk := a.getSDK(isTest)

	var sessions []playcamp.CreatePlaytimeSessionParams
	for i, raw := range body.Sessions {
		var s struct {
			SessionID       string                 `json:"sessionId"`
			UserID          string                 `json:"userId"`
			DurationSeconds int                    `json:"durationSeconds"`
			StartedAt       string                 `json:"startedAt"`
			EndedAt         string                 `json:"endedAt"`
			Metadata        map[string]interface{} `json:"metadata,omitempty"`
		}
		if err := json.Unmarshal(raw, &s); err != nil {
			writeError(w, http.StatusBadRequest, "invalid session at index "+strconv.Itoa(i))
			return
		}

		startedAt, err := time.Parse(time.RFC3339, s.StartedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid startedAt at index "+strconv.Itoa(i))
			return
		}
		endedAt, err := time.Parse(time.RFC3339, s.EndedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid endedAt at index "+strconv.Itoa(i))
			return
		}

		sessions = append(sessions, playcamp.CreatePlaytimeSessionParams{
			SessionID:       s.SessionID,
			UserID:          s.UserID,
			DurationSeconds: s.DurationSeconds,
			StartedAt:       startedAt,
			EndedAt:         endedAt,
			Metadata:        s.Metadata,
		})
	}

	result, err := sdk.PlaytimeSessions.CreateBulk(r.Context(), playcamp.CreateBulkPlaytimeSessionParams{
		Sessions:   sessions,
		CallbackID: body.CallbackID,
		IsTest:     body.IsTest,
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}
