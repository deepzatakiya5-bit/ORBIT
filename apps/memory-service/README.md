# ORBIT Long-Term Memory Service

Production-oriented long-term memory subsystem for a persistent AI companion.
This service is intentionally separated from the current Go chat API so it can scale independently and evolve memory logic without coupling to request latency.

## 1) Architecture

### Layered design

1. **Raw conversation storage**  
   Existing `conversations` + `messages` tables in Postgres remain source-of-truth.
2. **Memory extraction pipeline**  
   BullMQ job (`memory-extraction`) runs asynchronously after conversation updates.
3. **Episodic event store**  
   `timeline_event` stores life events with dates/entities/location/inference flag.
4. **Semantic memory store**  
   `memory` + `memory_embedding` (`pgvector`) stores structured memory + vectors.
5. **Hierarchical summaries**  
   `daily_summary -> weekly_summary -> monthly_summary -> yearly_summary`.
6. **Retrieval + ranking engine**  
   Hybrid scoring = semantic similarity + salience + confidence + decay + recency.
7. **Prompt memory injection**  
   `buildMemoryContext()` builds compact, budgeted context before response generation.
8. **Memory decay system**  
   Nightly decay recomputes `decayScore` and derived `salience`.
9. **Contradiction resolution**  
   Contradiction groups mark older memories as `SUPERSEDED` while preserving history.
10. **Background schedulers**  
   BullMQ repeat jobs trigger decay and periodic summarization.

## 2) Folder structure

```text
apps/memory-service
├── prisma/
│   └── schema.prisma
├── src/
│   ├── app.ts
│   ├── index.ts
│   ├── config/
│   ├── db/
│   ├── llm/
│   ├── prompts/
│   ├── types/
│   ├── services/
│   ├── jobs/
│   │   ├── queues.ts
│   │   ├── scheduler.ts
│   │   └── workers/
│   └── interfaces/http/
│       └── routes.ts
└── README.md
```

## 3) Memory extraction pipeline

`POST /internal/events/conversation-updated` enqueues extraction:

- loads user + recent messages
- prompts Ollama with strict JSON schema (`format: "json"`)
- validates JSON with zod
- persists only meaningful memories (`confidence >= 0.45`)
- deduplicates against active same-content/same-type memory
- computes salience/decay/recency
- stores vector embedding for semantic retrieval
- writes timeline events
- runs contradiction resolution

## 4) Hierarchical summary flow

`messages -> daily -> weekly -> monthly -> yearly`

Focus dimensions:
- life events/transitions
- emotional shifts
- goals and progress drift
- relationship changes
- routines + identity evolution

## 5) Retrieval and ranking

`retrieveMemories()`:

1. embeds user query
2. vector search on `memory_embedding`
3. filters to `validity = ACTIVE`
4. computes final score:
   - 40% semantic similarity
   - 25% salience
   - 15% confidence
   - 10% decay
   - 10% temporal recency boost
5. returns top-N memories for prompt context

## 6) Memory salience + decay

- `calculateSalience()` combines recency, emotional weight, repetition, explicit importance, user emphasis, goal relevance, and decay.
- `updateDecay()` applies exponential half-life per memory type:
  - identity (slow)
  - goals/relationships (slow-medium)
  - preferences/routines (medium)
  - facts/events (faster)
- memories are **not deleted**, only ranked lower over time.

## 7) Contradiction handling

Contradictory memories share a `contradictionGroup` key (example: `location.residence`).
When a stronger/newer memory arrives:
- prior active memories in the same group become `SUPERSEDED`
- `validUntil` is set
- provenance metadata links superseder id

## 8) Prompt memory injection

`buildMemoryContext()` injects:
- identity anchors
- active goals
- recent emotional patterns
- fresh daily + weekly summaries

Context is token-budgeted (`MEMORY_CONTEXT_BUDGET_TOKENS`) and logged in `memory_access_log`.

## 9) API endpoints

- `GET /v1/users/:userId/memory`
- `GET /v1/users/:userId/memory/search?query=...`
- `GET /v1/users/:userId/memory/context?query=...`
- `POST /v1/users/:userId/memory/:memoryId/reinforce`
- `POST /v1/users/:userId/memory/:memoryId/feedback`
- `DELETE /v1/users/:userId/memory/:memoryId`
- `GET /v1/users/:userId/timeline`
- `GET /v1/users/:userId/timeline/insights`
- `GET /v1/users/:userId/summaries`
- `GET /v1/users/:userId/export`
- `GET /internal/debug/users/:userId/retrieval-viewer?query=...`
- `GET /internal/debug/users/:userId/contradictions`

## 10) End-to-end memory flow

1. User sends chat message in Go API.
2. Go API stores message normally.
3. Go API calls memory-service internal endpoint (async trigger).
4. Worker extracts structured memories/events.
5. Embeddings + salience + contradiction resolution applied.
6. Daily/weekly/... summaries roll up by scheduled jobs.
7. Before next assistant reply, chat backend requests `memory/context`.
8. Prompt receives compact continuity context.

## 11) Cron strategy

- 02:00 daily: decay recomputation (`memory-decay`)
- 02:15 daily: user summary rollups (`memory-summarization`)

## 12) Production recommendations

- isolate memory service from synchronous chat response path
- use idempotency keys for event ingestion
- shard workers by user hash when scale increases
- add dead-letter queues for repeated LLM parse failures
- add circuit-breakers for upstream LLM quota/rate-limit failures
- maintain per-user extraction quotas to prevent runaway costs

## 13) Scaling strategy

- **DB**: pgvector IVF/HNSW indexes and partitioned access logs
- **Workers**: independent autoscaling pools (`extraction`, `summary`, `decay`)
- **Redis**: dedicated cluster for queue durability and throughput
- **API**: stateless horizontal scale behind gateway

## 14) Token optimization strategy

- extract from bounded windows (`<=120` messages, recency-first)
- compress with hierarchical summaries rather than replaying raw transcripts
- memory-context budget truncation and salience-first selection
- skip low-value extraction (`qualityScore < 0.45`)

## 15) Security and privacy

- memory-level provenance (`sourceConversationId`, timestamps, confidence)
- soft-delete/inactivation instead of physical deletion by default
- user export endpoint for transparency and portability
- encrypt DB at rest and TLS in transit
- redact sensitive fields before LLM extraction prompt when required
- add tenant-aware authorization on all userId routes (gateway/middleware)

## 16) Setup

```bash
cd apps/memory-service
npm install
npm run prisma:generate
npm run dev
```

Ensure Ollama is running locally and pull models:

```bash
ollama pull qwen2.5:7b-instruct
ollama pull nomic-embed-text
```

Start workers separately:

```bash
npm run worker:extract
npm run worker:summary
npm run worker:decay
npm run scheduler
```

