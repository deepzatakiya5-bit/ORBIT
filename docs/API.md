# ORBIT API Guide

Base URL (local): `http://localhost:8080`

**Dev UI:** open [http://localhost:8080/ui/](http://localhost:8080/ui/) in your browser to test the full flow without curl.

All JSON endpoints use:

```
Content-Type: application/json
```

Errors return:

```json
{
  "error": "human-readable message"
}
```

## Authentication

Protected routes require a JWT:

```http
Authorization: Bearer <access_token>
```

**Get a token from:** `POST /v1/users`, `PUT /v1/users/{id}/onboarding`, `POST /v1/auth/login`, or `POST /v1/auth/refresh`.

Tokens are valid for **30 days**. Use **refresh** before expiry to extend another 30 days. After expiry, use **login** with your `email`, `phone`, or `device_id`.

| Route | Auth required |
|-------|----------------|
| `POST /v1/users` | No |
| `POST /v1/auth/login` | No |
| `POST /v1/auth/refresh` | Bearer token (must still be valid) |
| `GET /health`, `GET /ready` | No |
| All other `/v1/*` routes | Yes |

---

## Recommended client flow

**New user**

```
1. POST   /v1/users
2. PUT    /v1/users/{userID}/onboarding   (include email, phone, or device_id)
3. POST   /v1/conversations              (creates thread + greeting once)
4. POST   /v1/conversations/{id}/messages
```

**Returning user**

```
1. POST   /v1/auth/login                  (email, phone, or device_id)
2. POST   /v1/conversations              (returns existing thread; no new greeting)
3. POST   /v1/conversations/{id}/messages
```

**Before token expires**

```
POST /v1/auth/refresh   (Authorization: Bearer <current token>)
```

---

## Auth: Login

Resume an existing account. Provide **exactly one** identifier (saved during onboarding).

**Request**

```http
POST /v1/auth/login
```

```json
{
  "device_id": "my-phone-001"
}
```

Or `"email": "you@example.com"` or `"phone": "+919876543210"`.

**Response** `200 OK`

```json
{
  "id": "5bb10382-37f4-48a2-8f9e-673d772993e9",
  "name": "Deep",
  "device_id": "my-phone-001",
  "onboarding_completed": true,
  "access_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

| Status | When |
|--------|------|
| `400` | Zero or multiple identifiers |
| `404` | No user with that identifier |
| `403` | Onboarding not completed |

---

## Auth: Refresh

Issue a **new** token with a fresh 30-day expiry. The current token must **not** be expired.

**Request**

```http
POST /v1/auth/refresh
Authorization: Bearer <access_token>
```

**Response** `200 OK`

```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "user_id": "5bb10382-37f4-48a2-8f9e-673d772993e9"
}
```

| Status | When |
|--------|------|
| `401` | Missing, invalid, or expired token |

---

## Test flow (curl)

Use `http://localhost:8081` if that is your `PORT`. Replace `TOKEN` and IDs from responses.

### A — New user + login identifier

```bash
BASE=http://localhost:8081

# 1. Create user
RESP=$(curl -s -X POST $BASE/v1/users)
TOKEN=$(echo $RESP | jq -r '.access_token')
USER_ID=$(echo $RESP | jq -r '.id')

# 2. Onboarding (device_id for easy re-login)
curl -s -X PUT $BASE/v1/users/$USER_ID/onboarding \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Deep",
    "nickname": "Deep",
    "gender": "male",
    "birthdate": "1998-05-15",
    "occupation": "Engineer",
    "preferred_language": "en",
    "timezone": "Asia/Kolkata",
    "communication_style": "warm_short",
    "why_here": ["companionship"],
    "device_id": "test-device-001"
  }' | jq .
```

### B — Login (same user, new token)

Simulate a new browser: discard old token, log in by device_id only.

```bash
LOGIN=$(curl -s -X POST $BASE/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"device_id": "test-device-001"}')
echo $LOGIN | jq .
TOKEN=$(echo $LOGIN | jq -r '.access_token')
USER_ID=$(echo $LOGIN | jq -r '.id')
```

### C — Refresh (extend session)

```bash
curl -s -X POST $BASE/v1/auth/refresh \
  -H "Authorization: Bearer $TOKEN" | jq .
# Save the new access_token from the response
```

### D — Refresh fails when expired

Use an old/invalid token — expect 401:

```bash
EXPIRED_TOKEN=replace-with-expired-or-garbage-jwt
curl -s -X POST $BASE/v1/auth/refresh \
  -H "Authorization: Bearer $EXPIRED_TOKEN"
```

### E — Chat after login

```bash
CONV=$(curl -s -X POST $BASE/v1/conversations \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{}')
CONV_ID=$(echo $CONV | jq -r '.conversation.id')

curl -s -X POST $BASE/v1/conversations/$CONV_ID/messages \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"content": "Hello again ORBIT"}' | jq .
```

### UI test

1. `make be` + `make fe` (or `/ui/`)
2. Create user → onboarding with **Device ID** `test-device-001`
3. Clear site data (or incognito)
4. **Log in** with `test-device-001` → should restore same user and chat access
5. Each browser **tab** has its own session (`sessionStorage`) — two users can test in two tabs

---

## 1. Health check

Verify the API process is running.

**Request**

```http
GET /health
```

**Response** `200 OK`

```json
{
  "status": "ok"
}
```

---

## 2. Readiness check

Verify the API can reach PostgreSQL.

**Request**

```http
GET /ready
```

**Response** `200 OK`

```json
{
  "status": "ready"
}
```

**Response** `503 Service Unavailable` (database down)

```json
{
  "status": "unavailable",
  "error": "connection refused"
}
```

---

## 3. Create user

Creates a new user with no profile yet. The mobile app should store `id` locally (e.g. secure storage) and show the onboarding screen next.

**Request**

```http
POST /v1/users
```

No body.

**Response** `201 Created`

```json
{
  "id": "5bb10382-37f4-48a2-8f9e-673d772993e9",
  "onboarding_completed": false,
  "created_at": "2026-05-27T10:34:54.44244+05:30",
  "updated_at": "2026-05-27T10:34:54.44244+05:30",
  "access_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

Store `access_token` and send it on all later requests.

**cURL**

```bash
curl -X POST http://localhost:8080/v1/users
```

---

## 4. Complete onboarding

Saves the user profile. Required before chat. Can only be called once per user.

**Request**

```http
PUT /v1/users/{userID}/onboarding
```

**Body**

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Full name |
| `nickname` | No | How ORBIT should address them (defaults to `name`) |
| `gender` | Yes | See [Gender values](#gender-values) |
| `birthdate` | Yes | `YYYY-MM-DD`, user must be 13+ |
| `occupation` | Yes | Job or role |
| `preferred_language` | Yes | e.g. `en`, `hi` |
| `timezone` | Yes | IANA timezone, e.g. `Asia/Kolkata` |
| `communication_style` | Yes | See [Communication style](#communication-style) |
| `why_here` | Yes | 1–5 items; see [Why here](#why-here) |
| `email` | No* | Unique; lowercased |
| `phone` | No* | Unique |
| `device_id` | No* | Unique; good for mobile dev |

\* At least **one** of `email`, `phone`, or `device_id` is required (for `POST /v1/auth/login` later).

**Example body**

```json
{
  "name": "Deep Zatakiya",
  "nickname": "Deep",
  "gender": "male",
  "birthdate": "1998-05-15",
  "occupation": "Software Engineer",
  "preferred_language": "en",
  "timezone": "Asia/Kolkata",
  "communication_style": "warm_short",
  "why_here": ["companionship", "goals"]
}
```

**Response** `200 OK`

```json
{
  "id": "5bb10382-37f4-48a2-8f9e-673d772993e9",
  "name": "Deep Zatakiya",
  "nickname": "Deep",
  "gender": "male",
  "birthdate": "1998-05-15T00:00:00Z",
  "occupation": "Software Engineer",
  "preferred_language": "en",
  "timezone": "Asia/Kolkata",
  "communication_style": "warm_short",
  "why_here": ["companionship", "goals"],
  "onboarding_completed": true,
  "created_at": "2026-05-27T10:34:54.44244+05:30",
  "updated_at": "2026-05-27T10:34:54.770054+05:30"
}
```

**Error responses**

| Status | When |
|--------|------|
| `400` | Invalid or missing fields |
| `404` | User not found |
| `409` | Onboarding already completed |

**cURL**

```bash
curl -X PUT http://localhost:8080/v1/users/USER_ID/onboarding \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Deep Zatakiya",
    "nickname": "Deep",
    "gender": "male",
    "birthdate": "1998-05-15",
    "occupation": "Software Engineer",
    "preferred_language": "en",
    "timezone": "Asia/Kolkata",
    "communication_style": "warm_short",
    "why_here": ["companionship", "goals"]
  }'
```

---

## 5. Get user

Returns the current profile. Use on app launch to decide whether to show onboarding or the chat home screen.

**Request**

```http
GET /v1/users/{userID}
```

**Response** `200 OK` (onboarding complete)

```json
{
  "id": "5bb10382-37f4-48a2-8f9e-673d772993e9",
  "name": "Deep Zatakiya",
  "nickname": "Deep",
  "gender": "male",
  "birthdate": "1998-05-15T00:00:00Z",
  "occupation": "Software Engineer",
  "preferred_language": "en",
  "timezone": "Asia/Kolkata",
  "communication_style": "warm_short",
  "why_here": ["companionship", "goals"],
  "onboarding_completed": true,
  "created_at": "2026-05-27T10:34:54.44244+05:30",
  "updated_at": "2026-05-27T10:34:54.770054+05:30"
}
```

**Response** `200 OK` (onboarding not complete)

```json
{
  "id": "5bb10382-37f4-48a2-8f9e-673d772993e9",
  "onboarding_completed": false,
  "created_at": "2026-05-27T10:34:54.44244+05:30",
  "updated_at": "2026-05-27T10:34:54.44244+05:30"
}
```

**Error responses**

| Status | When |
|--------|------|
| `400` | Invalid UUID |
| `404` | User not found |

**cURL**

```bash
curl http://localhost:8080/v1/users/USER_ID
```

---

## 6. Conversation (one per user)

Each user has **at most one** conversation thread. Messages accumulate in that thread across sessions.

Requires completed onboarding.

### Get conversation

```http
GET /v1/conversations
```

**Headers:** `Authorization: Bearer <access_token>`

**Response** `200 OK` — existing thread

```json
{
  "conversation": {
    "id": "90a70e9b-c133-464e-94e8-3e0bf571d927",
    "user_id": "5bb10382-37f4-48a2-8f9e-673d772993e9",
    "created_at": "2026-05-27T10:34:55.021699+05:30",
    "updated_at": "2026-05-27T11:00:00.021699+05:30"
  }
}
```

No thread yet:

```json
{
  "conversation": null
}
```

### Create or resume conversation

```http
POST /v1/conversations
```

**Body** (optional)

| Field | Required | Description |
|-------|----------|-------------|
| `title` | No | Optional thread title (only applied on first create) |

The user is taken from the JWT (do not pass `user_id`).

**Response** `201 Created` — first time only; includes ORBIT’s greeting

```json
{
  "conversation": {
    "id": "90a70e9b-c133-464e-94e8-3e0bf571d927",
    "user_id": "5bb10382-37f4-48a2-8f9e-673d772993e9",
    "created_at": "2026-05-27T10:34:55.021699+05:30"
  },
  "greeting": {
    "id": "a94004c5-62ef-49f4-9b12-44a947b72f9f",
    "conversation_id": "90a70e9b-c133-464e-94e8-3e0bf571d927",
    "role": "assistant",
    "content": "Hi Deep! I'm ORBIT, your companion...",
    "created_at": "2026-05-27T10:34:55.025938+05:30"
  }
}
```

**Response** `200 OK` — returning user; same thread, **no** `greeting` field

```json
{
  "conversation": {
    "id": "90a70e9b-c133-464e-94e8-3e0bf571d927",
    "user_id": "5bb10382-37f4-48a2-8f9e-673d772993e9",
    "created_at": "2026-05-27T10:34:55.021699+05:30",
    "updated_at": "2026-05-27T11:00:00.021699+05:30"
  }
}
```

**Error responses**

| Status | When |
|--------|------|
| `400` | Invalid JSON body |
| `403` | Onboarding not completed |
| `404` | User not found |
| `502` | LLM greeting generation failed (create only) |

**cURL**

```bash
curl -X POST http://localhost:8080/v1/conversations \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{}'
```

---

## 7. List messages

Returns all messages in a conversation, oldest first.

**Request**

```http
GET /v1/conversations/{conversationID}/messages
```

**Response** `200 OK`

```json
{
  "messages": [
    {
      "id": "a94004c5-62ef-49f4-9b12-44a947b72f9f",
      "conversation_id": "90a70e9b-c133-464e-94e8-3e0bf571d927",
      "role": "assistant",
      "content": "Hi Deep! I'm ORBIT...",
      "created_at": "2026-05-27T10:34:55.025938+05:30"
    },
    {
      "id": "1bfc54ed-4dd5-4d72-a1ec-a81dfd4a105e",
      "conversation_id": "90a70e9b-c133-464e-94e8-3e0bf571d927",
      "role": "user",
      "content": "Hey ORBIT, I have an interview tomorrow.",
      "created_at": "2026-05-27T10:35:10.502414+05:30"
    }
  ]
}
```

Empty conversation:

```json
{
  "messages": []
}
```

**Error responses**

| Status | When |
|--------|------|
| `400` | Invalid conversation UUID |
| `404` | Conversation not found |

**cURL**

```bash
curl http://localhost:8080/v1/conversations/CONVERSATION_ID/messages
```

---

## 8. Send message

Sends a user message and returns both the saved user message and ORBIT's reply.

Requires completed onboarding.

**Request**

```http
POST /v1/conversations/{conversationID}/messages
```

**Body**

| Field | Required | Description |
|-------|----------|-------------|
| `content` | Yes | User's message text |

**Example body**

```json
{
  "content": "Hey ORBIT, I have an interview tomorrow."
}
```

**Response** `200 OK`

```json
{
  "user_message": {
    "id": "1bfc54ed-4dd5-4d72-a1ec-a81dfd4a105e",
    "conversation_id": "90a70e9b-c133-464e-94e8-3e0bf571d927",
    "role": "user",
    "content": "Hey ORBIT, I have an interview tomorrow.",
    "created_at": "2026-05-27T10:35:10.502414+05:30"
  },
  "assistant_message": {
    "id": "35b0bb91-127c-48ea-8452-19a59ebbaefd",
    "conversation_id": "90a70e9b-c133-464e-94e8-3e0bf571d927",
    "role": "assistant",
    "content": "That's a big day tomorrow — how are you feeling about it?",
    "created_at": "2026-05-27T10:35:10.510983+05:30"
  }
}
```

**Error responses**

| Status | When |
|--------|------|
| `400` | Missing `content` or invalid conversation UUID |
| `403` | Onboarding not completed |
| `404` | Conversation not found |
| `502` | LLM reply generation failed |

**cURL**

```bash
curl -X POST http://localhost:8080/v1/conversations/CONVERSATION_ID/messages \
  -H 'Content-Type: application/json' \
  -d '{"content": "Hey ORBIT, I have an interview tomorrow."}'
```

---

## Reference values

### Gender values

| Value | |
|-------|---|
| `male` | |
| `female` | |
| `non_binary` | |
| `other` | |
| `prefer_not_to_say` | |

### Communication style

| Value | Meaning |
|-------|---------|
| `warm_short` | Warm and concise |
| `warm_long` | Warm and detailed |
| `direct_short` | Direct and concise |
| `playful_short` | Playful and concise |

### Why here

Pick **1 to 5** (no duplicates):

| Value | |
|-------|---|
| `companionship` | |
| `stress_support` | |
| `goals` | |
| `loneliness` | |
| `curiosity` | |
| `self_growth` | |

### Message roles

| Role | Description |
|------|-------------|
| `user` | Message from the human |
| `assistant` | Message from ORBIT |
| `system` | Internal (not returned via API today) |

---

## Full local walkthrough

Prerequisites: Postgres running (`docker compose up -d`), `.env` configured, API started (`go run ./apps/api/...`).

```bash
# 1. Create user
USER=$(curl -s -X POST http://localhost:8080/v1/users)
USER_ID=$(echo $USER | jq -r '.id')

# 2. Onboarding
curl -s -X PUT http://localhost:8080/v1/users/$USER_ID/onboarding \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Deep Zatakiya",
    "nickname": "Deep",
    "gender": "male",
    "birthdate": "1998-05-15",
    "occupation": "Software Engineer",
    "preferred_language": "en",
    "timezone": "Asia/Kolkata",
    "communication_style": "warm_short",
    "why_here": ["companionship", "goals"]
  }'

# 3. Start conversation (includes greeting)
CONV=$(curl -s -X POST http://localhost:8080/v1/conversations \
  -H 'Content-Type: application/json' \
  -d "{\"user_id\": \"$USER_ID\"}")
CONV_ID=$(echo $CONV | jq -r '.conversation.id')

# 4. Send a message
curl -s -X POST http://localhost:8080/v1/conversations/$CONV_ID/messages \
  -H 'Content-Type: application/json' \
  -d '{"content": "Hey ORBIT, I have an interview tomorrow."}'

# 5. List history
curl -s http://localhost:8080/v1/conversations/$CONV_ID/messages
```

---

## LLM behavior

| Environment | Behavior |
|-------------|----------|
| `LLM_PROVIDER=ollama` | Local replies via Ollama (`OLLAMA_MODEL`, default `qwen2.5:7b-instruct`) |
| `LLM_PROVIDER=gemini` + `GEMINI_API_KEY` set | Replies via Gemini |
| Otherwise | Dev stub — echoes input with a placeholder prefix |

Profile fields (name, nickname, timezone, communication style, why_here, etc.) are injected into the LLM system prompt for greetings and chat replies.

---

## Long-term memory service

If `MEMORY_SERVICE_URL` is configured in the API:

- each chat turn asynchronously triggers memory extraction
- memory context is fetched before each assistant reply and injected into the prompt

Memory service endpoints:

- `GET /v1/users/{userId}/memory`
- `GET /v1/users/{userId}/memory/search?query=...`
- `GET /v1/users/{userId}/memory/context?query=...`
- `GET /v1/users/{userId}/timeline`
- `GET /v1/users/{userId}/timeline/insights`
- `GET /v1/users/{userId}/summaries`
- `POST /v1/users/{userId}/memory/{memoryId}/reinforce`
- `POST /v1/users/{userId}/memory/{memoryId}/feedback`
- `DELETE /v1/users/{userId}/memory/{memoryId}`
- `GET /v1/users/{userId}/export`

See `apps/memory-service/README.md` for architecture, workers, and deployment model.
