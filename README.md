# Tech-Nova — Universal Real-Time Helpdesk, AI Assistant & Telegram Relay Microservice

A high-performance, universal multi-tenant customer support and live chat microservice built with Go. Features an automated AI first-responder pipeline, media attachment handling (images & videos), bi-directional WebSocket broadcasting, and seamless human escalation via Telegram Supergroup Forum Topics.

---

## Tech Stack

- **Language:** Go 1.26+
- **Framework:** [Gin Gonic](https://github.com/gin-gonic/gin)
- **Real-Time Communication:** [Gorilla WebSocket](https://github.com/gorilla/websocket)
- **Database:** PostgreSQL (`github.com/lib/pq`)
- **AI Integrations:** OpenAI / OpenRouter / Groq / Google Gemini APIs
- **Telegram Bot API:** Native Telegram Bot HTTP API (Forum Topics & Webhooks)
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate)

---

## Project Structure

```
.
├── cmd/
│   └── main.go                      # Application entrypoint & dependency wiring
├── internal/
│   ├── config/
│   │   └── config.go                # Environment variable configuration
│   ├── database/
│   │   └── postgres.go              # PostgreSQL connection setup & connection pooling
│   ├── middleware/
│   │   ├── auth.go                  # API key & tenant verification middleware
│   │   ├── cors.go                  # Cross-Origin Resource Sharing (CORS) handler
│   │   └── logger.go                # HTTP request logger middleware
│   ├── tenant/
│   │   ├── handler.go               # Tenant HTTP handlers
│   │   ├── service.go               # Tenant business logic & registration
│   │   ├── repository.go            # Tenant database queries
│   │   ├── model.go                 # Tenant entity model
│   │   └── dto.go                   # Tenant request & response DTOs
│   ├── ticket/
│   │   ├── handler.go               # Ticket HTTP handlers (Init, Rate)
│   │   ├── service.go               # Ticket lifecycle & topic naming logic
│   │   ├── repository.go            # Ticket database queries & thread ID lookups
│   │   ├── model.go                 # Ticket entity model
│   │   └── dto.go                   # Ticket request & response DTOs
│   ├── chat/
│   │   ├── handler.go               # Message & attachment HTTP handlers
│   │   ├── service.go               # Chat orchestrator, media processing, & WS dispatch
│   │   ├── repository.go            # Message database queries
│   │   ├── model.go                 # Message entity model
│   │   └── dto.go                   # Message request & response DTOs
│   ├── ai/
│   │   ├── service.go               # AI LLM prompt orchestration & vision analyzer
│   │   └── dto.go                   # AI context & request DTOs
│   ├── telegram/
│   │   ├── handler.go               # Telegram webhook callback query & message handler
│   │   ├── service.go               # Telegram Forum Topic, media upload & message dispatcher
│   │   └── dto.go                   # Telegram webhook update DTOs
│   └── websocket/
│       ├── handler.go               # WebSocket upgrade handler (`/ws`)
│       ├── hub.go                   # Room manager & bi-directional event dispatcher
│       └── client.go                # WebSocket client read/write pumps & ping-pong heartbeat
├── migrations/                      # Versioned SQL migration files (.up.sql & .down.sql)
├── scripts/
│   ├── migrate.sh                   # Helper script for migration operations
│   └── test_e2e.go                  # Automated end-to-end integration test simulation
├── uploads/                         # Storage directory for image and video attachments
├── API_SPEC.md                      # Detailed REST, WebSocket, & Webhook protocol spec
├── FLOW.md                          # Full architectural lifecycle and sequence flow
├── .env                             # Environment variables
├── go.mod
└── go.sum
```

---

## Environment Variables

Create a `.env` file in the root directory:

```env
# App Configuration
APP_NAME=tech-nova
APP_PORT=8080
APP_ENV=development
BASE_URL=http://localhost:8080

# Database Configuration (PostgreSQL)
DATABASE_URL=postgres://postgres:postgres@127.0.0.1:5432/technova_db?sslmode=disable
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=technova_db
DB_SSLMODE=disable

# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=YOUR_TELEGRAM_BOT_TOKEN
TELEGRAM_CHAT_ID=YOUR_TELEGRAM_SUPERGROUP_ID

# AI Configuration (OpenRouter / Groq / OpenAI / Gemini)
AI_PROVIDER=openrouter
AI_ENDPOINT=https://openrouter.ai/api/v1/chat/completions
AI_API_KEY=YOUR_AI_API_KEY
AI_MODEL=openai/gpt-4o-mini

# Upload Configuration
MAX_UPLOAD_IMAGE_MB=5
MAX_UPLOAD_VIDEO_MB=15
UPLOAD_DIR=./uploads
```

---

## Database Migrations

Manage database schemas using the migration script:

```bash
# Apply all pending migrations
./scripts/migrate.sh up

# Rollback last migration
./scripts/migrate.sh down 1

# Check migration version status
./scripts/migrate.sh status

# Create a new migration file
./scripts/migrate.sh create <migration_name>
```

---

## End-to-End Automated Testing

An automated verification test script is included to test full end-to-end lifecycle flows:

```bash
go run scripts/test_e2e.go
```

**Verification behavior:**
- Health check verification (`GET /health`).
- Tenant & Ticket initialization (`POST /api/v1/tickets/init`).
- AI Auto-Responder pipeline (`POST /api/v1/chat/send`).
- Multipart screenshot attachment upload (`POST /api/v1/chat/upload`).
- Chat history retrieval (`GET /api/v1/tickets/:id/messages`).
- Telegram Webhook simulation (`POST /webhook/telegram` for ticket claim).
- CSAT rating submission (`POST /api/v1/tickets/:id/rate`).

---

## API Documentation

### 1. Tickets

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/tickets/init` | Initialize or retrieve active support session |
| `POST` | `/api/v1/tickets/:ticket_id/rate` | Submit CSAT satisfaction rating (1-5 stars) |

#### `POST /api/v1/tickets/init`
```json
// Request Body
{
  "key_identifier": "tenant_client_01",
  "app_name": "E-Commerce App",
  "tenant_name": "Store Asia",
  "user_id": "104",
  "user_name": "Rayhan",
  "module_name": "Payment Gateway",
  "diagnostic_info": {
    "url": "https://store.example.com/checkout"
  }
}

// Response (200 OK)
{
  "status": true,
  "ticket": {
    "id": 1,
    "ticket_code": "TCK-20261001-2be43e",
    "user_id": "104",
    "user_name": "Rayhan",
    "module_name": "Payment Gateway",
    "status": "open",
    "created_at": "2026-10-01T21:40:00Z"
  }
}
```

#### `POST /api/v1/tickets/:ticket_id/rate`
```json
// Request Body
{
  "rating": 5,
  "review": "Fast response and very helpful assistance!"
}

// Response (200 OK)
{
  "status": true,
  "message": "Terima kasih atas penilaian Anda!"
}
```

---

### 2. Chat & Attachments

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/chat/send` | Send text message (AI responder + Telegram forward) |
| `POST` | `/api/v1/chat/upload` | Upload image (max 5MB) or video recording (max 15MB) |
| `GET`  | `/api/v1/tickets/:ticket_id/messages` | Retrieve conversation history |

#### `POST /api/v1/chat/send`
```json
// Request Body
{
  "ticket_id": 1,
  "message": "Hello, how do I process a refund request?"
}

// Response (200 OK)
{
  "status": true,
  "user_msg_id": 12,
  "ai_msg_id": 13,
  "sender": "ai",
  "reply": "Hello! To process a refund, please navigate to Orders > Transactions, select the order, and click Request Refund."
}
```

#### `POST /api/v1/chat/upload`
**Request Type:** `multipart/form-data`
* `ticket_id`: `1` (Text field)
* `caption`: `Screenshot saat error tombol bayar` (Text field)
* `attachment`: `[File Binary - Image PNG/JPG or Video MP4/WebM]`

```json
// Response (200 OK)
{
  "status": true,
  "user_msg_id": 14,
  "ai_msg_id": 15,
  "sender": "ai",
  "file_url": "http://localhost:8080/uploads/image_TCK-20261001-2be43e_1790867760.png",
  "reply": "Terima kasih Kak, tangkapan layar sudah kami terima. Tim Programmer akan segera menganalisa tampilan error tersebut."
}
```

#### `GET /api/v1/tickets/:ticket_id/messages`
```json
// Response (200 OK)
{
  "status": true,
  "data": [
    {
      "id": 12,
      "ticket_id": 1,
      "sender_type": "user",
      "sender_name": "Rayhan",
      "message": "Hello, how do I process a refund request?",
      "is_attachment": false,
      "attachment_type": "",
      "attachment_url": "",
      "created_at": "2026-10-01T21:40:05Z"
    },
    {
      "id": 13,
      "ticket_id": 1,
      "sender_type": "ai",
      "sender_name": "AI Support",
      "message": "Hello! To process a refund, please navigate to Orders > Transactions, select the order, and click Request Refund.",
      "is_attachment": false,
      "attachment_type": "",
      "attachment_url": "",
      "created_at": "2026-10-01T21:40:07Z"
    }
  ]
}
```

---

### 3. Real-Time WebSocket (`/ws`)

Connect client frontend widgets directly via WebSocket:
```text
ws://localhost:8080/ws?ticket_id=1&key=tenant_client_01
```

**Incoming server events:**
- `helpdesk_new_message`: Dispatched when an AI reply or Telegram human responder sends a message.
- `helpdesk_status_changed`: Dispatched when a ticket status changes (`escalated`, `waiting_user`, `resolved`).

```json
// Event Example: helpdesk_new_message
{
  "event": "helpdesk_new_message",
  "data": {
    "id": 16,
    "ticket_id": 1,
    "sender_type": "programmer",
    "assigned_programmer": "Rayhan Programmer",
    "message": "Issue has been resolved on the backend, please refresh!",
    "created_at": "2026-10-01T21:45:00Z"
  }
}
```

---

### 4. Telegram Webhook Relay (`POST /webhook/telegram`)

Register single webhook endpoint with Telegram Bot API:
```text
https://api.telegram.org/bot<BOT_TOKEN>/setWebhook?url=https://hub.yourdomain.com/webhook/telegram
```

- **Callback `claim_{ticket_id}`:** Claims ticket, locks assignee, and replays history & media to topic.
- **Callback `resolve_{ticket_id}`:** Changes ticket status to `waiting_user` and prompts client CSAT review.
- **Topic Messages:** Relays programmer replies to client WebSocket room in real-time.

---

## Roadmap Status

- [x] Feature-Driven 3-Tier Layered Architecture
- [x] Multi-Tenant Support (`key_identifier` namespace isolation)
- [x] PostgreSQL Connection Pooling & Migrations (`golang-migrate`)
- [x] AI First-Responder Engine (OpenAI / OpenRouter / Groq / Gemini)
- [x] Telegram Forum Topic Auto-Creation (`🏷️ [Topic] - [Tenant Name]`)
- [x] Image (max 5MB) & Video (max 15MB) Upload Engine
- [x] Human Takeover & Atomic Claim Lock via Telegram Inline Keyboards
- [x] Chat & Media Replay on Claim
- [x] Bi-Directional WebSocket Relay Engine (`/ws`)
- [x] Resolution Flow & CSAT Star Rating (1-5 Stars)
- [x] Automated End-to-End Integration Test Suite (`scripts/test_e2e.go`)

---

## License

This project is open-source software licensed under the [MIT License](LICENSE).
