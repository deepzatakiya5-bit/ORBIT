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

---

## Recommended client flow

```
1. POST   /v1/users                          → create user (skeleton)
2. PUT    /v1/users/{userID}/onboarding        → complete profile (required)
3. GET    /v1/users/{userID}                   → verify profile (optional)
4. POST   /v1/conversations                    → start chat + ORBIT greeting
5. GET    /v1/conversations/{id}/messages      → load history (optional)
6. POST   /v1/conversations/{id}/messages      → send messages
```

Steps 4–6 require `onboarding_completed: true`. Otherwise the API returns **403**.

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
  "updated_at": "2026-05-27T10:34:54.44244+05:30"
}
```

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

## 6. Create conversation

Starts a new chat thread. ORBIT sends the **first message** (personalized greeting) automatically.

Requires completed onboarding.

**Request**

```http
POST /v1/conversations
```

**Body**

| Field | Required | Description |
|-------|----------|-------------|
| `user_id` | Yes | UUID from `POST /v1/users` |
| `title` | No | Optional thread title |

**Example body**

```json
{
  "user_id": "5bb10382-37f4-48a2-8f9e-673d772993e9",
  "title": "Daily check-in"
}
```

**Response** `201 Created`

```json
{
  "conversation": {
    "id": "90a70e9b-c133-464e-94e8-3e0bf571d927",
    "user_id": "5bb10382-37f4-48a2-8f9e-673d772993e9",
    "title": "Daily check-in",
    "created_at": "2026-05-27T10:34:55.021699+05:30"
  },
  "greeting": {
    "id": "a94004c5-62ef-49f4-9b12-44a947b72f9f",
    "conversation_id": "90a70e9b-c133-464e-94e8-3e0bf571d927",
    "role": "assistant",
    "content": "Hi Deep! I'm ORBIT, your companion. I know you're here for companionship and someone to talk to — I'll talk with you in en. How are you feeling today?",
    "created_at": "2026-05-27T10:34:55.025938+05:30"
  }
}
```

**Error responses**

| Status | When |
|--------|------|
| `400` | Invalid body or `user_id` |
| `403` | Onboarding not completed |
| `404` | User not found |
| `502` | LLM greeting generation failed |

**cURL**

```bash
curl -X POST http://localhost:8080/v1/conversations \
  -H 'Content-Type: application/json' \
  -d '{"user_id": "USER_ID"}'
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
| `OPENAI_API_KEY` set | Real replies via OpenAI (`OPENAI_MODEL`, default `gpt-4o-mini`) |
| `OPENAI_API_KEY` unset | Dev stub — echoes input with a placeholder prefix |

Profile fields (name, nickname, timezone, communication style, why_here, etc.) are injected into the LLM system prompt for greetings and chat replies.
