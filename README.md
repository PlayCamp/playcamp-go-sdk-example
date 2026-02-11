# PlayCamp Go SDK Example Server

An example server that exposes all PlayCamp Go SDK Server APIs as HTTP endpoints.

## Quick Start

### 1. Configure environment variables

```bash
cp .env.example .env
```

Open `.env` and fill in your actual values:

```
SERVER_API_KEY=ak_server_your_key_id:your_secret
WEBHOOK_SECRET=your_webhook_secret_hex
SDK_ENVIRONMENT=sandbox
```

### 2. Run the server

```bash
go run .
```

### 3. Access

- Web UI: http://localhost:4000
- API: http://localhost:4000/api/campaigns

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/campaigns | List campaigns |
| GET | /api/campaigns/:id | Get campaign |
| GET | /api/campaigns/:id/creators | Get campaign creators |
| GET | /api/creators/search | Search creators |
| GET | /api/creators/:key | Get creator |
| GET | /api/creators/:key/coupons | Get creator coupons |
| POST | /api/coupons/validate | Validate coupon |
| POST | /api/coupons/redeem | Redeem coupon |
| GET | /api/coupons/user/:userId | Get coupon history |
| GET | /api/sponsors/:userId | Get sponsor |
| POST | /api/sponsors | Create sponsor |
| PUT | /api/sponsors/:userId | Update sponsor |
| DELETE | /api/sponsors/:userId | Delete sponsor |
| GET | /api/sponsors/:userId/history | Get sponsor history |
| POST | /api/payments | Create payment |
| GET | /api/payments/:txnId | Get payment |
| GET | /api/payments/user/:userId | Get user payments |
| POST | /api/payments/:txnId/refund | Refund payment |
| GET | /api/webhooks | List webhooks |
| POST | /api/webhooks | Create webhook |
| PUT | /api/webhooks/:id | Update webhook |
| DELETE | /api/webhooks/:id | Delete webhook |
| GET | /api/webhooks/:id/logs | Get webhook logs |
| POST | /api/webhooks/:id/test | Test webhook |
| POST | /webhooks/playcamp | Receive webhooks |
| GET | /api/webhooks/received | Get received webhooks |
| DELETE | /api/webhooks/received | Clear received webhooks |
| POST | /api/webhooks/simulate | Simulate webhook |

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| SERVER_API_KEY | Yes | Server API key (`keyId:secret` format) |
| WEBHOOK_SECRET | No | Webhook signature verification secret |
| SDK_ENVIRONMENT | No | `sandbox` or `live` (default: `live`) |
| SDK_API_URL | No | Custom API URL (overrides environment) |
| SDK_DEBUG | No | Enable debug logging (`true`/`false`) |
| PORT | No | Server port (default: `4000`) |

## Test Mode

Use the Test Mode toggle in the Web UI or add `?isTest=true` query parameter to make API calls in test mode.
For POST requests, include `"isTest": true` in the JSON body.
