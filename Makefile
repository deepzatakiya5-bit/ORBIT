.PHONY: up local-up down db wait-db be fe logs mem mem-worker-extract mem-worker-summary mem-worker-decay mem-scheduler

PORT ?= 8081
FE_PORT ?= 3000
POSTGRES_PORT ?= 5434
API_URL ?= http://localhost:$(PORT)
export POSTGRES_PORT

# Full local stack: Postgres + API (embedded UI at /ui/)
up local-up: db wait-db
	@echo "→ API + UI: $(API_URL)  (dev UI: $(API_URL)/ui/)"
	$(MAKE) be

db:
	@echo "→ Postgres on localhost:$(POSTGRES_PORT)"
	docker compose up -d

wait-db:
	@echo "Waiting for Postgres…"
	@for i in $$(seq 1 30); do \
		docker compose exec -T postgres pg_isready -U orbit -d orbit >/dev/null 2>&1 && exit 0; \
		sleep 1; \
	done; \
	echo "Postgres did not become ready in time"; exit 1

# Backend API (loads .env from repo root)
be:
	@echo "→ API: $(API_URL)"
	go run ./apps/api

# Frontend dev server (HTML/CSS/JS live reload; point UI at $(API_URL))
fe:
	@echo "→ FE:  http://localhost:$(FE_PORT)"
	@echo "→ API: $(API_URL)  (set this in the Connection panel if needed)"
	@echo "Start the API in another terminal: make be"
	cd apps/api/web && python3 -m http.server $(FE_PORT)

down:
	docker compose down

logs:
	docker compose logs -f postgres

mem:
	cd apps/memory-service && npm run dev

mem-worker-extract:
	cd apps/memory-service && npm run worker:extract

mem-worker-summary:
	cd apps/memory-service && npm run worker:summary

mem-worker-decay:
	cd apps/memory-service && npm run worker:decay

mem-scheduler:
	cd apps/memory-service && npm run scheduler
