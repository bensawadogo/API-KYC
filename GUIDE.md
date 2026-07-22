# Guide développeur — DATAKEYS KYC

## Structure du projet

```
├── services/kyc/               ← API Go
│   ├── cmd/main.go             ← Point d'entrée (wiring)
│   ├── config/config.go        ← Configuration (variables d'env)
│   ├── internal/
│   │   ├── interfaces.go       ← Interfaces partagées
│   │   ├── model/              ← Structures de données (KYC, APIKey)
│   │   ├── handler/            ← Handlers HTTP (Fiber)
│   │   ├── service/            ← Logique métier (Initiate, Process)
│   │   ├── middleware/         ← Auth, rate limit, idempotence
│   │   ├── provider/           ← Providers (SmileID, Youverify, SumSub, Local, Sandbox)
│   │   ├── repository/         ← Accès PostgreSQL
│   │   ├── compliance/         ← RuleEvaluator, SAR
│   │   ├── notification/       ← SSE, SMS, RedisHub
│   │   ├── resilience/         ← Retry, circuit breaker, fallback, DLQ
│   │   ├── registry/           ← Country registry
│   │   ├── countries/          ← Données des pays africains
│   │   ├── storage/            ← MinIO, Memory
│   │   ├── webhook/            ← Webhook sender
│   │   ├── audit/              ← Audit logger
│   │   ├── observability/      ← Prometheus, OpenTelemetry
│   │   ├── job/                ← Background jobs (cleanup)
│   │   └── seed/               ← Données de test sandbox
│   ├── migrations/             ← SQL migrations
│   ├── api/                    ← OpenAPI spec
│   ├── infra/                  ← Prometheus, Grafana
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── Makefile
│
├── sdks/
│   ├── javascript/             ← SDK JS/TS (tsup build)
│   │   ├── src/index.ts        ← Point d'entrée
│   │   ├── src/client.ts       ← HTTP client
│   │   ├── src/errors.ts       ← KYCError typé
│   │   ├── src/types.ts        ← Types TS
│   │   └── src/resources/     ← kyc.ts, countries.ts
│   ├── python/                 ← SDK Python
│   │   ├── datakeys/           ← Package
│   │   │   ├── client.py       ← HTTP client
│   │   │   ├── models.py       ← Dataclasses
│   │   │   ├── errors.py       ← KYCError
│   │   │   └── resources/     ← kyc.py, countries.py
│   │   └── pyproject.toml
│   ├── go/                     ← SDK Go
│   │   ├── datakeys.go         ← Client principal
│   │   ├── models.go           ← Types
│   │   ├── errors.go           ← KYCError
│   │   ├── kyc.go              ← KYC service
│   │   └── countries.go        ← Countries service
│   └── examples/               ← Quickstarts (JS, Python, Go)
│
├── ARCHITECTURE.md             ← Architecture système
├── CONTINUE.md                 ← Instructions pour IA
└── GUIDE.md                    ← Ce fichier
```

## Configuration (variables d'environnement)

### Obligatoires
| Variable | Description |
|----------|-------------|
| `POSTGRES_URL` | URL de connexion PostgreSQL |

### Serveur
| Variable | Défaut | Description |
|----------|--------|-------------|
| `SERVER_PORT` | `8081` | Port d'écoute |
| `SERVER_ENV` | `development` | `development`, `production`, `sandbox` |
| `TLS_CERT_PATH` | — | Chemin vers le certificat TLS |
| `TLS_KEY_PATH` | — | Chemin vers la clé TLS |

### Providers
| Variable | Défaut | Description |
|----------|--------|-------------|
| `SMILEID_API_KEY` | — | API key SmileID |
| `SMILEID_PARTNER_ID` | — | Partner ID SmileID |
| `SMILEID_SANDBOX` | `true` | Mode sandbox SmileID |
| `YOUVERIFY_API_KEY` | — | API key Youverify |
| `YOUVERIFY_SANDBOX` | `true` | Mode sandbox Youverify |
| `SUMSUB_APP_TOKEN` | — | App token SumSub |
| `SUMSUB_SECRET_KEY` | — | Secret key SumSub |
| `SUMSUB_SANDBOX` | `true` | Mode sandbox SumSub |

### Stockage
| Variable | Défaut | Description |
|----------|--------|-------------|
| `STORAGE_PROVIDER` | `memory` | `memory` ou `minio` |
| `MINIO_ENDPOINT` | `localhost:9000` | Endpoint MinIO |
| `MINIO_ACCESS_KEY` | — | Access key MinIO |
| `MINIO_SECRET_KEY` | — | Secret key MinIO |
| `MINIO_BUCKET` | `kyc-documents` | Bucket MinIO |

### Resilience
| Variable | Défaut | Description |
|----------|--------|-------------|
| `PROVIDER_RETRY_MAX` | `3` | Nombre max de tentatives |
| `PROVIDER_RETRY_BASE_DELAY` | `500ms` | Délai initial backoff |
| `PROVIDER_RETRY_MAX_DELAY` | `10s` | Délai max backoff |
| `CB_MAX_FAILURES` | `5` | Seuil d'ouverture circuit breaker |
| `CB_TIMEOUT` | `30s` | Timeout avant half-open |

### AML
| Variable | Défaut | Description |
|----------|--------|-------------|
| `AML_PROVIDER` | `local` | `local` ou `opensanctions` |
| `AML_THRESHOLD` | `0.70` | Seuil de score AML |

## Commandes courantes

### API
```bash
# Lancer en mode sandbox
cd services/kyc && set POSTGRES_URL=postgres://user:pass@localhost:5432/kyc_db?sslmode=disable && set SERVER_ENV=sandbox && go run ./cmd/

# Lancer avec docker-compose
cd services/kyc && make up

# Tous les tests
cd services/kyc && go test ./... -timeout 60s -count=1 -v

# Couverture
cd services/kyc && go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
```

### SDKs
```bash
# JS/TS
cd sdks/javascript && npm install && npm run build

# Python
cd sdks/python && pip install -e . && python -c "from datakeys import Datakeys; print('ok')"

# Go
cd sdks/go && go build ./... && go vet ./...
```

## Ajouter un provider

1. Créer `internal/provider/<name>.go` implémentant `IdentityProvider`
2. Créer `internal/provider/<name>_test.go` avec tests unitaires
3. Ajouter la config dans `config/config.go` (struct + parsing)
4. Ajouter la config dans `.env.example` et `docker-compose.yml`
5. Instancier dans `cmd/main.go` et l'ajouter au router
6. Ajouter des scénarios sandbox dans `internal/provider/sandbox.go` si pertinent
7. Mettre à jour les SDKs si le provider expose des fonctionnalités uniques

## Ajouter un pays

1. Ajouter dans `internal/countries/registry.go` (code, nom, prefix, région, provider, doc types, régulations)
2. Ajouter les règles de validation de numéro de document
3. Ajouter un profil sandbox dans `internal/seed/testdata.go`
4. Mettre à jour les SDKs (types, exemples)

## Ajouter une règle de compliance

1. Ajouter la constante de régulation dans `internal/compliance/rules.go`
2. Implémenter l'interface `Rule` (`Name()`, `Evaluate()`, `AppliesTo()`)
3. Enregistrer dans `RuleEvaluator`
4. Ajouter des cas de test dans `internal/compliance/rules_test.go`

## Endpoints API

| Méthode | Path | Auth | Description |
|---------|------|------|-------------|
| `GET` | `/health/live` | — | Liveness probe |
| `GET` | `/health/ready` | — | Readiness probe |
| `GET` | `/metrics` | — | Prometheus metrics |
| `GET` | `/docs` | — | Swagger UI |
| `GET` | `/v1/kyc/countries` | — | Liste des pays |
| `GET` | `/v1/kyc/countries/:code/doctypes` | — | Types de doc par pays |
| `POST` | `/v1/kyc/initiate` | API key | Initier une vérification |
| `GET` | `/v1/kyc/status/:id` | API key | Statut d'une vérification |
| `GET` | `/v1/notifications/stream` | — | SSE notifications |
| `POST` | `/v1/kyc/webhook/smileid` | IP allowlist | Webhook SmileID |
| `POST` | `/v1/kyc/webhook/sumsub` | IP allowlist | Webhook SumSub |
| `GET` | `/sandbox/profiles` | — | Profils sandbox |
| `GET` | `/sandbox/reset` | — | Reset sandbox |
| `POST` | `/sandbox/simulate` | — | Simulation sandbox |

## Migration du schéma BDD

Les migrations sont dans `services/kyc/migrations/` et exécutées au démarrage par `repository.RunMigrations()`.

Format: `NNN_description.sql` (ex: `006_phone_hash.sql`).

Pour ajouter une migration :
1. Créer le fichier SQL
2. Ajouter `repository.RunMigrations()` — il les exécute dans l'ordre alphabétique
3. Mettre à jour les requêtes dans `internal/repository/postgres.go`
4. Ajouter les champs dans `internal/model/kyc.go`
