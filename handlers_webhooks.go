package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	playcamp "github.com/playcamp/playcamp-go-sdk"
	"github.com/playcamp/playcamp-go-sdk/webhookutil"
)

// webhookStore is a thread-safe in-memory store for received webhooks.
type webhookStore struct {
	mu       sync.Mutex
	webhooks []receivedWebhook
	counter  int
	maxSize  int
}

type receivedWebhook struct {
	ID         string          `json:"id"`
	Valid      bool            `json:"valid"`
	Error      string          `json:"error,omitempty"`
	Events     []webhookEvent  `json:"events"`
	ReceivedAt string          `json:"receivedAt"`
	RawBody    json.RawMessage `json:"rawBody,omitempty"`
}

type webhookEvent struct {
	Event     string          `json:"event"`
	Timestamp string          `json:"timestamp"`
	IsTest    *bool           `json:"isTest,omitempty"`
	Data      json.RawMessage `json:"data"`
}

func newWebhookStore(maxSize int) *webhookStore {
	return &webhookStore{maxSize: maxSize}
}

func (s *webhookStore) add(wh receivedWebhook) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counter++
	wh.ID = fmt.Sprintf("wh_%d", s.counter)
	wh.ReceivedAt = time.Now().UTC().Format(time.RFC3339)

	// Prepend (newest first).
	s.webhooks = append([]receivedWebhook{wh}, s.webhooks...)

	if len(s.webhooks) > s.maxSize {
		s.webhooks = s.webhooks[:s.maxSize]
	}
}

func (s *webhookStore) list() []receivedWebhook {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]receivedWebhook, len(s.webhooks))
	copy(result, s.webhooks)
	return result
}

func (s *webhookStore) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.webhooks = nil
}

// --- Webhook Management Handlers ---

// handleListWebhooks handles GET /api/webhooks
func (a *app) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	webhooks, err := a.server.Webhooks.List(r.Context())
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, webhooks)
}

// handleCreateWebhook handles POST /api/webhooks
func (a *app) handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EventType  playcamp.WebhookEventType `json:"eventType"`
		URL        string                    `json:"url"`
		RetryCount *int                      `json:"retryCount,omitempty"`
		TimeoutMs  *int                      `json:"timeoutMs,omitempty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	webhook, err := a.server.Webhooks.Create(r.Context(), playcamp.CreateWebhookParams{
		EventType:  body.EventType,
		URL:        body.URL,
		RetryCount: body.RetryCount,
		TimeoutMs:  body.TimeoutMs,
	})
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, webhook)
}

// handleUpdateWebhook handles PUT /api/webhooks/{id}
func (a *app) handleUpdateWebhook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook ID")
		return
	}

	var body playcamp.UpdateWebhookParams
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	webhook, err := a.server.Webhooks.Update(r.Context(), id, body)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, webhook)
}

// handleDeleteWebhook handles DELETE /api/webhooks/{id}
func (a *app) handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook ID")
		return
	}

	if err := a.server.Webhooks.Delete(r.Context(), id); err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// handleGetWebhookLogs handles GET /api/webhooks/{id}/logs
func (a *app) handleGetWebhookLogs(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook ID")
		return
	}

	logs, err := a.server.Webhooks.GetLogs(r.Context(), id)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

// handleTestWebhook handles POST /api/webhooks/{id}/test
func (a *app) handleTestWebhook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook ID")
		return
	}

	result, err := a.server.Webhooks.Test(r.Context(), id)
	if err != nil {
		handleSDKError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// --- Webhook Receiver ---

// handleWebhookReceiver handles POST /webhooks/playcamp
func (a *app) handleWebhookReceiver(w http.ResponseWriter, r *http.Request) {
	body, err := readRawBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	signature := r.Header.Get("X-Webhook-Signature")

	result := webhookutil.Verify(webhookutil.VerifyOptions{
		Payload:   body,
		Signature: signature,
		Secret:    a.webhookSecret,
	})

	wh := receivedWebhook{
		Valid:   result.Valid,
		RawBody: body,
	}

	if !result.Valid {
		wh.Error = result.Error
	}

	if result.Payload != nil {
		for _, evt := range result.Payload.Events {
			wh.Events = append(wh.Events, webhookEvent{
				Event:     string(evt.Event),
				Timestamp: evt.Timestamp,
				IsTest:    evt.IsTest,
				Data:      evt.Data,
			})
		}
	}

	a.receivedWebhooks.add(wh)

	// Log webhook reception.
	if result.Valid {
		var events []string
		for _, evt := range wh.Events {
			events = append(events, evt.Event)
		}
		log.Printf("[webhook] received valid webhook: events=[%s]", strings.Join(events, ", "))
	} else {
		log.Printf("[webhook] received invalid webhook: %s", result.Error)
	}

	writeJSON(w, http.StatusOK, map[string]bool{"received": true})
}

// --- In-Memory Webhook Store Endpoints ---

// handleGetReceivedWebhooks handles GET /api/webhooks/received
func (a *app) handleGetReceivedWebhooks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.receivedWebhooks.list())
}

// handleClearReceivedWebhooks handles DELETE /api/webhooks/received
func (a *app) handleClearReceivedWebhooks(w http.ResponseWriter, r *http.Request) {
	a.receivedWebhooks.clear()
	writeJSON(w, http.StatusOK, map[string]bool{"cleared": true})
}

// handleSimulateWebhook handles POST /api/webhooks/simulate
func (a *app) handleSimulateWebhook(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Event string          `json:"event"`
		Data  json.RawMessage `json:"data"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	wh := receivedWebhook{
		Valid: true,
		Events: []webhookEvent{
			{
				Event:     body.Event,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Data:      body.Data,
			},
		},
	}

	a.receivedWebhooks.add(wh)
	writeJSON(w, http.StatusOK, map[string]bool{"simulated": true})
}
