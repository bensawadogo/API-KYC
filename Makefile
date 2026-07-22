.PHONY: all test test-api test-sdks test-widget lint build docker-up docker-down \
        docker-logs smoke clean help

all: test lint

# ─── Tests ────────────────────────────────────────

test: test-api test-sdks test-widget

test-api:
	cd services/kyc && set CGO_ENABLED=0 && go test ./... -timeout 120s -count=1

test-sdks:
	cd sdks/go && go build ./... 2>&1 | tail -3
	cd sdks/go && go vet ./...
	cd sdks/javascript && npm run typecheck 2>&1 | tail -3
	cd sdks/javascript && npm run build 2>&1 | tail -3
	cd sdks/javascript && npm test 2>&1 | tail -5

test-widget:
	cd sdks/react-widget && npm run typecheck 2>&1 | tail -3
	cd sdks/react-widget && npm run build 2>&1 | tail -3
	cd sdks/react-widget && npm test 2>&1 | tail -5

# ─── Lint + Build ──────────────────────────────────

lint:
	cd services/kyc && go vet ./...
	cd sdks/go && go vet ./...

build:
	cd services/kyc && go build ./cmd/...
	cd sdks/javascript && npm run build 2>&1 | tail -3
	cd sdks/react-widget && npm run build 2>&1 | tail -3

# ─── Docker ───────────────────────────────────────

docker-up:
	cd services/kyc && docker compose up -d --build
	@echo ""
	@echo "KYC API    => http://localhost:8081"
	@echo "Grafana    => http://localhost:3000 (admin/\$${GF_ADMIN_PASSWORD})"
	@echo "Jaeger     => http://localhost:16686"
	@echo "MinIO UI   => http://localhost:9001"
	@echo ""
	@echo "Clé API sandbox : dk_test_datakeys_sandbox_001"
	@echo "Seeded automatiquement dans PostgreSQL"

docker-down:
	cd services/kyc && docker compose down -v --remove-orphans

docker-logs:
	cd services/kyc && docker compose logs -f kyc-service

# ─── Smoke ────────────────────────────────────────

smoke:
	cd services/kyc && make smoke

# ─── Clean ────────────────────────────────────────

clean:
	cd services/kyc && docker compose down -v --remove-orphans 2>/dev/null || true
	cd sdks/javascript && rm -rf dist node_modules 2>/dev/null || true
	cd sdks/react-widget && rm -rf dist node_modules 2>/dev/null || true
	cd sdks/python && find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	rm -f coverage.out

# ─── Help ─────────────────────────────────────────

help:
	@echo "DATAKEYS KYC — Commandes disponibles"
	@echo ""
	@echo "  make test       Run all tests (API + SDK JS + Widget)"
	@echo "  make test-api   Run Go API tests only"
	@echo "  make test-sdks  Run SDK tests (Go + JS)"
	@echo "  make test-widget Run React widget tests"
	@echo "  make lint       Run Go vet"
	@echo "  make build      Compile Go + JS + Widget"
	@echo "  make docker-up  Start full stack (Postgres + Redis + MinIO + API + Grafana)"
	@echo "  make docker-down Stop and clean volumes"
	@echo "  make smoke      Smoke test the running API"
	@echo "  make clean      Remove build artifacts + node_modules"
