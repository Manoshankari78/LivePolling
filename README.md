# Live Polling Tool

A production-oriented real-time polling application built for the GUVI Developer Internship task.

## Overview

The application implements:

**Create Poll → Share Link → Audience Votes → Results Update Live**

Authenticated creators register/login with JWT, create and manage polls, share a public poll URL, and watch results update in real time. Audience members vote without an account. MongoDB persists users, polls, and votes. Redis Pub/Sub distributes vote events between backend instances, and WebSockets push those events to connected React clients.

## Tech Stack

- Frontend: React + Vite
- Backend: Go + Gin
- Database: MongoDB
- Realtime: Redis Pub/Sub + WebSocket
- Authentication: JWT
- Password hashing: bcrypt
- Local infrastructure: Docker Compose
  
### REST vs WebSocket

REST is used for request/response operations: registration, login, poll management, reading a poll, and submitting a vote. WebSocket is used only for server-to-client live result delivery. Redis Pub/Sub decouples persistence from the realtime broadcast layer and also allows multiple backend instances to receive the same poll event.

## Project Structure

```text
live-polling-app/
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

## Database Design

### users

- `_id`
- `name`
- `email` — unique index
- `passwordHash`
- `createdAt`
- `updatedAt`

### polls

- `_id`
- `creatorId`
- `question`
- `options[]` with `id`, `text`, and `votes`
- `status` — active/closed
- `createdAt`
- `updatedAt`
- `expiresAt` — optional

Indexes: `creatorId`, `status`, and `createdAt`.

### votes

- `_id`
- `pollId`
- `optionId`
- `voterId`
- `createdAt`

A unique compound index on `(pollId, voterId)` prevents the same voter ID from voting twice for the same poll.

## Poll Editing

Creators can edit active polls. Before the first vote, option structure can change freely. After voting starts, the number of options is kept fixed so historical vote counts cannot be accidentally discarded; existing option IDs/counts are preserved while their text and the question may be edited. Closed polls cannot be edited.

## Duplicate Vote Prevention

The public client creates a random voter ID and stores it in `localStorage`. Every vote sends that ID in `X-Voter-ID`. MongoDB enforces uniqueness on `(pollId, voterId)`, so duplicate requests are rejected even if two requests race. This is intentionally described as **practical anonymous-session prevention**, not strong identity verification: a user who clears storage or changes devices can obtain another voter ID. A production system needing stronger identity guarantees can use authenticated voters, signed server-side cookies, email verification, or device/IP risk controls.

## Realtime Flow

1. React opens `/api/polls/:id/ws`.
2. Go registers the socket under that poll ID.
3. A vote is POSTed to `/api/polls/:id/vote`.
4. Gin validates the poll, option, expiration/status, and voter ID.
5. MongoDB inserts the vote under the unique compound index.
6. MongoDB atomically increments the selected option's stored count.
7. Go reads the authoritative poll result and publishes an event to Redis.
8. Every backend instance subscribed to `poll:*:updates` receives the event.
9. The local WebSocket hub broadcasts to every connected viewer of that poll.
10. React replaces its result state immediately — no polling or page refresh.

## Authentication Flow

1. Registration validates name/email/password and hashes the password with bcrypt.
2. Login compares the submitted password with the bcrypt hash.
3. A signed JWT containing the user ID is returned.
4. Protected requests send `Authorization: Bearer <token>`.
5. Gin middleware validates the signature and expiry and stores the user ID in request context.
6. Controllers/services enforce creator ownership for update/delete/close operations.

Passwords and password hashes are never returned by the API.

## API

| Method | Endpoint | Auth | Purpose |
|---|---|---|---|
| POST | `/api/auth/register` | No | Create account |
| POST | `/api/auth/login` | No | Authenticate |
| GET | `/api/auth/me` | Yes | Current user |
| POST | `/api/polls` | Yes | Create poll |
| GET | `/api/polls` | Yes | Creator dashboard polls |
| GET | `/api/polls/:id` | No | Read public poll |
| PUT | `/api/polls/:id` | Yes | Update poll |
| DELETE | `/api/polls/:id` | Yes | Delete poll |
| POST | `/api/polls/:id/close` | Yes | Close poll |
| POST | `/api/polls/:id/vote` | No | Cast anonymous vote |
| GET | `/api/polls/:id/results` | No | Read current results |
| GET | `/api/polls/:id/ws` | No | Live result WebSocket |

### Create poll request

```json
{
  "question": "Which language do you prefer?",
  "options": ["JavaScript", "Go", "Python"],
  "expiresAt": null
}
```

### Vote request

```json
{
  "optionId": "generated-option-id",
  "voterId": "browser-session-id"
}
```

### Live event

```json
{
  "pollId": "...",
  "results": [
    {"optionId": "...", "votes": 12},
    {"optionId": "...", "votes": 8}
  ],
  "totalVotes": 20
}
```

## Validation and Errors

- `400` invalid request data
- `401` missing/invalid authentication
- `403` authenticated user does not own the resource
- `404` resource not found
- `409` duplicate vote or conflicting operation
- `500` internal server error

Frontend error messages are translated into human-readable notifications.

## Local Setup

### 1. Start MongoDB and Redis

```bash
docker compose up -d
```

### 2. Backend

```bash
cd backend
cp ../.env.example .env
# Edit secrets/URLs if necessary
go mod tidy
go run ./cmd/server
```

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

Open the Vite URL shown in the terminal, normally `http://localhost:5173`.

## Environment Variables

The backend reads:

- `PORT`
- `MONGO_URI`
- `MONGO_DATABASE`
- `REDIS_URL`
- `JWT_SECRET`
- `JWT_EXPIRES_HOURS`
- `FRONTEND_URL`
- `CORS_ORIGIN`

The frontend reads:

- `VITE_API_URL`
- `VITE_WS_URL`

Never commit a real `.env` file.

## Testing

Backend unit tests cover validation, JWT creation/verification, and duplicate-vote service behavior with test-only repository doubles. The application itself always uses real MongoDB and Redis repositories in production/runtime code.

```bash
cd backend
go test ./...
```

Frontend:

```bash
cd frontend
npm run build
```

## Docker / Production

The root Compose file intentionally runs MongoDB and Redis locally. For deployment, use managed services such as MongoDB Atlas and a managed Redis provider. Deploy the React frontend to a static hosting platform and the Go backend to a platform supporting long-running HTTP/WebSocket processes.

Production settings must provide:

- a strong random `JWT_SECRET`
- managed MongoDB URI
- managed Redis URL
- exact frontend origin in CORS
- HTTPS frontend URL
- `wss://` WebSocket URL
- `VITE_API_URL` pointing to the public backend API
- `VITE_WS_URL` pointing to the public WebSocket API

Ensure the chosen backend host supports WebSockets and does not terminate idle connections too aggressively.

## Security Considerations

- bcrypt password hashing
- JWT authentication
- creator authorization checks
- backend validation of all poll/vote data
- MongoDB uniqueness for duplicate votes
- strict CORS origin configuration
- environment-based secrets
- WebSocket origin validation
- bounded request sizes
- basic in-memory IP rate limiting on vote requests
- generic internal error responses

The included rate limiter is intentionally lightweight and per-process. A multi-instance production deployment should move rate limiting to Redis or the edge layer.

## Scaling Considerations

MongoDB remains the source of truth for persistent state. Redis Pub/Sub lets every backend instance receive the same event, while each instance broadcasts to its own connected WebSocket clients. Because Redis Pub/Sub is not durable, clients fetch the current results over REST on initial load/reconnect; the WebSocket stream is an acceleration mechanism, not the source of truth.

For larger workloads, use Redis-backed rate limiting, connection-aware horizontal scaling, metrics/tracing, a durable event stream if event replay is required, and MongoDB transaction/consistency strategies appropriate to the deployment topology.

## Design Decisions

- **React:** component-based UI and simple reactive result rendering.
- **Go/Gin:** fast, typed HTTP backend with clear service/repository boundaries.
- **MongoDB:** natural document representation for polls/options and flexible metadata.
- **Redis Pub/Sub:** distributes live events across backend instances without coupling sockets directly to vote requests.
- **WebSocket:** server push is appropriate because all viewers need immediate updates.
- **REST:** request/response semantics fit CRUD and vote commands.

## Known Limitations

- Anonymous duplicate-vote prevention is session/device based and cannot prove real-world identity.
- The lightweight rate limiter is process-local.
- Redis Pub/Sub events are ephemeral; REST remains the recovery path after reconnects.
- Public deployment credentials and domains must be supplied by the deployer.

## Demo Checklist

1. Register a creator.
2. Login.
3. Create a poll with at least two options.
4. Copy the public link.
5. Open it in two browser windows.
6. Vote in one window.
7. Observe the other window update without refreshing.
8. Return to the dashboard and edit/close/delete the poll.

## Internship Interview Talking Points

Be prepared to explain:

- why REST and WebSocket have different responsibilities
- why Redis is required instead of direct in-process broadcasts
- why MongoDB is the source of truth
- how JWT authentication works
- why bcrypt is used instead of encryption for passwords
- how the compound vote index prevents races
- why frontend vote counts are never trusted
- why WebSocket reconnect performs a REST refresh
- how multiple Go instances can share Redis Pub/Sub events
- what changes would be needed for stronger anonymous identity and large-scale rate limiting
