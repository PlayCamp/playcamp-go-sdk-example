package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	playcamp "github.com/playcamp/playcamp-go-sdk"
)

// app holds the SDK instances and shared state.
type app struct {
	server           *playcamp.Server
	testServer       *playcamp.Server
	webhookSecret    string
	receivedWebhooks *webhookStore
}

// getSDK returns the appropriate SDK instance based on test mode.
func (a *app) getSDK(isTest bool) *playcamp.Server {
	if isTest {
		return a.testServer
	}
	return a.server
}

func main() {
	// Load .env file (ignore error if not present).
	_ = godotenv.Load()

	// Read required config.
	apiKey := os.Getenv("SERVER_API_KEY")
	if apiKey == "" {
		log.Fatal("SERVER_API_KEY environment variable is required")
	}

	webhookSecret := os.Getenv("WEBHOOK_SECRET")
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	// Build SDK options.
	var opts []playcamp.Option

	env := os.Getenv("SDK_ENVIRONMENT")
	if env != "" {
		opts = append(opts, playcamp.WithEnvironment(playcamp.Environment(env)))
	}

	if baseURL := os.Getenv("SDK_API_URL"); baseURL != "" {
		opts = append(opts, playcamp.WithBaseURL(baseURL))
	}

	if strings.EqualFold(os.Getenv("SDK_DEBUG"), "true") {
		opts = append(opts, playcamp.WithDebug(playcamp.DebugOptions{
			Enabled:         true,
			LogRequestBody:  true,
			LogResponseBody: true,
		}))
	}

	// Create normal SDK instance.
	server, err := playcamp.NewServer(apiKey, opts...)
	if err != nil {
		log.Fatalf("Failed to create SDK server: %v", err)
	}

	// Create test-mode SDK instance.
	testOpts := append([]playcamp.Option{playcamp.WithTestMode(true)}, opts...)
	testServer, err := playcamp.NewServer(apiKey, testOpts...)
	if err != nil {
		log.Fatalf("Failed to create test SDK server: %v", err)
	}

	a := &app{
		server:           server,
		testServer:       testServer,
		webhookSecret:    webhookSecret,
		receivedWebhooks: newWebhookStore(50),
	}

	// Router setup.
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// --- Campaigns ---
	r.Get("/api/campaigns", a.handleListCampaigns)
	r.Get("/api/campaigns/{id}", a.handleGetCampaign)
	r.Get("/api/campaigns/{id}/creators", a.handleGetCampaignCreators)

	// --- Creators (literal path before parameterized) ---
	r.Get("/api/creators/search", a.handleSearchCreators)
	r.Get("/api/creators/{key}", a.handleGetCreator)
	r.Get("/api/creators/{key}/coupons", a.handleGetCreatorCoupons)

	// --- Coupons ---
	r.Post("/api/coupons/validate", a.handleValidateCoupon)
	r.Post("/api/coupons/redeem", a.handleRedeemCoupon)
	r.Get("/api/coupons/user/{userId}", a.handleGetCouponHistory)

	// --- Sponsors ---
	r.Post("/api/sponsors", a.handleCreateSponsor)
	r.Get("/api/sponsors/{userId}", a.handleGetSponsor)
	r.Put("/api/sponsors/{userId}", a.handleUpdateSponsor)
	r.Delete("/api/sponsors/{userId}", a.handleDeleteSponsor)
	r.Get("/api/sponsors/{userId}/history", a.handleGetSponsorHistory)

	// --- Payments (literal path before parameterized) ---
	r.Post("/api/payments", a.handleCreatePayment)
	r.Get("/api/payments/user/{userId}", a.handleGetUserPayments)
	r.Get("/api/payments/{transactionId}", a.handleGetPayment)
	r.Post("/api/payments/{transactionId}/refund", a.handleRefundPayment)

	// --- Webhooks (literal paths before parameterized) ---
	r.Get("/api/webhooks", a.handleListWebhooks)
	r.Post("/api/webhooks", a.handleCreateWebhook)
	r.Get("/api/webhooks/received", a.handleGetReceivedWebhooks)
	r.Delete("/api/webhooks/received", a.handleClearReceivedWebhooks)
	r.Post("/api/webhooks/simulate", a.handleSimulateWebhook)
	r.Get("/api/webhooks/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Not a standard endpoint, but route exists for completeness.
		writeError(w, http.StatusNotFound, "use /api/webhooks/:id/logs or /api/webhooks/:id/test")
	})
	r.Put("/api/webhooks/{id}", a.handleUpdateWebhook)
	r.Delete("/api/webhooks/{id}", a.handleDeleteWebhook)
	r.Get("/api/webhooks/{id}/logs", a.handleGetWebhookLogs)
	r.Post("/api/webhooks/{id}/test", a.handleTestWebhook)

	// --- Webhook Receiver ---
	r.Post("/webhooks/playcamp", a.handleWebhookReceiver)

	// --- Static files ---
	fileServer := http.FileServer(http.Dir("public"))
	r.Handle("/*", fileServer)

	// Print startup banner.
	effectiveAPIURL := os.Getenv("SDK_API_URL")
	if effectiveAPIURL == "" {
		effectiveAPIURL = playcamp.EnvironmentURL(playcamp.Environment(env))
		if effectiveAPIURL == "" {
			effectiveAPIURL = playcamp.EnvironmentURL(playcamp.EnvironmentLive)
		}
	}

	envInfo := fmt.Sprintf("Environment: %s", env)
	if os.Getenv("SDK_API_URL") != "" {
		envInfo = fmt.Sprintf("Custom: %s", effectiveAPIURL)
	}
	if env == "" {
		envInfo = "Environment: live"
	}

	debugStatus := "Debug: OFF"
	if strings.EqualFold(os.Getenv("SDK_DEBUG"), "true") {
		debugStatus = "Debug: ON"
	}

	fmt.Printf(`
╔═══════════════════════════════════════════════════╗
║     PlayCamp SDK Example Server (Go)              ║
╠═══════════════════════════════════════════════════╣
║  Server: http://localhost:%s
║  SDK API: %s
║  %s
║  %s
╚═══════════════════════════════════════════════════╝

API Endpoints:

[Campaigns]
   GET  /api/campaigns              - List campaigns
   GET  /api/campaigns/:id          - Get campaign
   GET  /api/campaigns/:id/creators - Get campaign creators

[Creators]
   GET  /api/creators/search        - Search creators
   GET  /api/creators/:key          - Get creator
   GET  /api/creators/:key/coupons  - Get creator coupons

[Coupons]
   POST /api/coupons/validate       - Validate coupon
   POST /api/coupons/redeem         - Redeem coupon
   GET  /api/coupons/user/:userId   - Get user coupon history

[Sponsors]
   GET  /api/sponsors/:userId          - Get sponsor
   POST /api/sponsors                  - Create sponsor
   PUT  /api/sponsors/:userId          - Update sponsor
   DELETE /api/sponsors/:userId        - Delete sponsor
   GET  /api/sponsors/:userId/history  - Get sponsor history

[Payments]
   POST /api/payments                    - Create payment
   GET  /api/payments/:txnId             - Get payment
   GET  /api/payments/user/:userId       - Get user payments
   POST /api/payments/:txnId/refund      - Refund payment

[Webhooks]
   GET  /api/webhooks             - List webhooks
   POST /api/webhooks             - Create webhook
   PUT  /api/webhooks/:id         - Update webhook
   DELETE /api/webhooks/:id       - Delete webhook
   GET  /api/webhooks/:id/logs    - Get webhook logs
   POST /api/webhooks/:id/test    - Test webhook

[Webhook Receiver]
   POST /webhooks/playcamp        - Receive webhooks
   GET  /api/webhooks/received    - Get received webhooks
   DELETE /api/webhooks/received  - Clear received webhooks
   POST /api/webhooks/simulate    - Simulate webhook
`, port, effectiveAPIURL, envInfo, debugStatus)

	log.Fatal(http.ListenAndServe(":"+port, r))
}

// corsMiddleware adds CORS headers for the Web UI.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
