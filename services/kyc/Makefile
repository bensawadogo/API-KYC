.PHONY: up down build test audit logs ps clean seed-apikey smoke

up:
	docker compose up -d --build
	@echo ""
	@echo "Services disponibles :"
	@echo "  KYC API    \u2192 http://localhost:8081"
	@echo "  Prometheus \u2192 http://localhost:9090"
	@echo "  Grafana    \u2192 http://localhost:3000  (admin/datakeys2025)"
	@echo "  Jaeger     \u2192 http://localhost:16686"
	@echo "  MinIO UI   \u2192 http://localhost:9001  (minioadmin/minioadmin123)"

down:
	docker compose down

build:
	docker compose build kyc-service

test:
	cd services/kyc && set CGO_ENABLED=0 && go test ./... -timeout 60s -count=1

audit:
	cd services/kyc && go build ./...
	cd services/kyc && go vet ./...
	cd services/kyc && set CGO_ENABLED=0 && go test ./... -timeout 60s -count=1 -cover

logs:
	docker compose logs -f kyc-service

ps:
	docker compose ps

clean:
	docker compose down -v --remove-orphans

seed-apikey:
	docker compose exec postgres psql -U kyc_user -d kyc_db -c \
	  "INSERT INTO api_keys (client_name, key_hash, key_prefix, scopes, rate_limit) \
	   VALUES ('DATAKEYS-Test', \
	           encode(digest('dk_test_datakeys_internal_key_001','sha256'),'hex'), \
	           'dk_test_datakey', \
	           ARRAY['kyc:initiate','kyc:status','kyc:admin'], \
	           1000) \
	   ON CONFLICT DO NOTHING;"
	@echo "API key de test cr\u00e9\u00e9e : dk_test_datakeys_internal_key_001"

smoke:
	@echo "Test 1 \u2014 sans cl\u00e9 \u2192 401"
	curl -s -o /dev/null -w "%{http_code}" \
	  -X POST http://localhost:8081/v1/kyc/initiate | grep 401
	@echo " OK"
	@echo "Test 2 \u2014 health/ready"
	curl -s http://localhost:8081/health/ready | python -m json.tool
	@echo "Test 3 \u2014 metrics endpoint"
	curl -s http://localhost:8081/metrics | grep "kyc_initiated_total"
