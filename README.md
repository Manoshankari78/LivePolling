# Live Polling

A real-time polling platform that lets authenticated creators publish polls, share public voting links, and watch results update instantly as votes arrive.

> **Create a poll → Share the link → Collect votes → Watch results live**

Built with React, Go, MongoDB, Redis Pub/Sub, and WebSockets.

## Features

- Creator registration and JWT-based authentication
- Create, edit, close, and delete polls
- Public poll pages that do not require audience accounts
- Anonymous voting with browser-based voter identification
- Duplicate-vote prevention enforced by MongoDB
- Optional poll expiration
- Live result updates over WebSockets
- Redis Pub/Sub support for multi-instance backend deployments
- Creator dashboard for managing polls
- Backend validation, ownership checks, CORS protection, and rate limiting
- Responsive React frontend with human-readable error messages

## How It Works

1. A creator registers or signs in.
2. The creator creates a poll and shares its public URL.
3. Audience members vote without creating an account.
4. The backend validates and persists each vote in MongoDB.
5. Redis Pub/Sub distributes the updated result across backend instances.
6. WebSocket connections broadcast the latest results to connected viewers.

MongoDB is the source of truth. The frontend never calculates or trusts vote totals locally.

## Architecture

```text
React + Vite
    │
    ├── REST API ───────────────┐
    └── WebSocket connection    │
                                ▼
                         Go + Gin API
                           │       │
                           │       ├── MongoDB
                           │       │    Users, polls, votes
                           │       │
                           │       └── Redis Pub/Sub
                           │            Cross-instance events
                           ▼
                     WebSocket Hub
                       Live results
```

### REST and WebSockets

REST handles request/response operations such as authentication, poll management, reading polls, and submitting votes. WebSockets are used only for server-to-client live result delivery. Redis Pub/Sub allows every backend instance to receive the same poll update and notify its connected clients.

## Technology Stack

| Layer | Technology |
|---|---|
| Frontend | React, Vite |
| Backend | Go, Gin |
| Database | MongoDB |
| Realtime messaging | Redis Pub/Sub |
| Client updates | WebSocket |
| Authentication | JWT |
| Password security | bcrypt |
| Local infrastructure | Docker Compose |

## Project Structure

```text
.
├── backend/
│   ├── cmd/server/main.go
│   ├── config/
│   ├── controllers/
│   ├── database/
│   ├── middleware/
│   ├── models/
│   ├── repositories/
│   ├── routes/
│   ├── services/
│   ├── utils/
│   ├── websocket/
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── context/
│   │   ├── hooks/
│   │   ├── layouts/
│   │   ├── pages/
│   │   ├── services/
│   │   ├── types/
│   │   ├── utils/
│   │   ├── App.jsx
│   │   └── main.jsx
│   └── package.json
├── .env.example
├── docker-compose.yml
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.20+
- Node.js 18+
- Docker and Docker Compose

### 1. Start the infrastructure

From the repository root, start MongoDB and Redis:

```bash
docker compose up -d
```

### 2. Configure and run the backend

```bash
cd backend
cp ../.env.example .env
# Update the values in .env as needed
go mod tidy
go run ./cmd/server
```

### 3. Install and run the frontend

In a new terminal:

```bash
cd frontend
npm install
npm run dev
```

Open the local URL printed by Vite, normally `http://localhost:5173`.

## Environment Variables

### Backend

| Variable | Description |
|---|---|
| `PORT` | HTTP server port |
| `MONGO_URI` | MongoDB connection string |
| `MONGO_DATABASE` | MongoDB database name |
| `REDIS_URL` | Redis connection URL |
| `JWT_SECRET` | Secret used to sign JWTs |
| `JWT_EXPIRES_HOURS` | JWT lifetime in hours |
| `FRONTEND_URL` | Frontend URL used by the application |
| `CORS_ORIGIN` | Allowed frontend origin(s) |

### Frontend

| Variable | Description |
|---|---|
| `VITE_API_URL` | Public backend API base URL |
| `VITE_WS_URL` | Backend WebSocket base URL |

Never commit a real `.env` file or production secrets.

## API Reference

| Method | Endpoint | Auth | Description |
|---|---|:---:|---|
| `POST` | `/api/auth/register` | No | Create an account |
| `POST` | `/api/auth/login` | No | Authenticate a creator |
| `GET` | `/api/auth/me` | Yes | Get the current user |
| `POST` | `/api/polls` | Yes | Create a poll |
| `GET` | `/api/polls` | Yes | List the creator's polls |
| `GET` | `/api/polls/:id` | No | Get a public poll |
| `PUT` | `/api/polls/:id` | Yes | Update an owned poll |
| `DELETE` | `/api/polls/:id` | Yes | Delete an owned poll |
| `POST` | `/api/polls/:id/close` | Yes | Close an owned poll |
| `POST` | `/api/polls/:id/vote` | No | Submit a vote |
| `GET` | `/api/polls/:id/results` | No | Get current results |
| `GET` | `/api/polls/:id/ws` | No | Subscribe to live results |

### Create a poll

```json
{
  "question": "Which language do you prefer?",
  "options": ["JavaScript", "Go", "Python"],
  "expiresAt": null
}
```

### Submit a vote

```json
{
  "optionId": "generated-option-id",
  "voterId": "browser-session-id"
}
```

The client also sends the voter ID in the `X-Voter-ID` header. The server validates all vote data and does not trust client-provided totals.

### Live result event

```json
{
  "pollId": "...",
  "results": [
    { "optionId": "...", "votes": 12 },
    { "optionId": "...", "votes": 8 }
  ],
  "totalVotes": 20
}
```

## Data Model

### Users

Stores creator identity and authentication data. Email addresses are unique, and passwords are stored only as bcrypt hashes.

### Polls

Stores the creator, question, options, vote counts, status, timestamps, and optional expiration time. Polls have indexes for creator, status, and creation time.

### Votes

Stores the poll, selected option, voter ID, and creation time. A unique compound index on `(pollId, voterId)` prevents the same voter ID from voting more than once on a poll, including under concurrent requests.

## Security and Validation

- Passwords are hashed with bcrypt and never returned by the API.
- JWT middleware protects creator-only operations.
- Services verify resource ownership before update, close, or delete operations.
- Poll and vote input is validated on the backend.
- Strict CORS and WebSocket origin validation are supported.
- Request sizes are bounded.
- Vote requests use lightweight in-memory IP rate limiting.
- Internal errors are returned as generic messages.

The included rate limiter is process-local. Multi-instance production deployments should move rate limiting to Redis or an edge provider.

## Testing and Quality Checks

Run backend tests:

```bash
cd backend
go test ./...
```

Build the frontend:

```bash
cd frontend
npm run build
```

Backend tests cover validation, JWT creation and verification, and duplicate-vote service behavior using test repositories where appropriate.

## Production Deployment

For production, use managed MongoDB and Redis services and deploy the frontend to a static hosting provider. Configure:

- A strong, randomly generated `JWT_SECRET`
- Managed MongoDB and Redis connection URLs
- The exact frontend origin in CORS settings
- HTTPS for the frontend and `wss://` for WebSockets
- Public `VITE_API_URL` and `VITE_WS_URL` values
- A backend host that supports long-lived WebSocket connections

Redis Pub/Sub distributes events between backend instances, while each instance manages its own connected WebSocket clients. For larger workloads, consider Redis-backed rate limiting, metrics and tracing, connection-aware scaling, and a durable event stream if event replay is required.

## Known Limitations

- Anonymous duplicate-vote prevention is browser/session based and cannot verify real-world identity.
- The local rate limiter is per process.
- Redis Pub/Sub messages are ephemeral; clients refresh results through REST after reconnecting.
- Production domains and deployment credentials must be configured by the deployer.

## Demo Checklist

1. Register a creator account.
2. Sign in and create a poll with at least two options.
3. Copy the public poll URL.
4. Open the URL in two browser windows.
5. Vote in one window.
6. Confirm the other window updates without a refresh.
7. Return to the dashboard and edit, close, or delete the poll.

## License

No license has been specified for this project yet.
