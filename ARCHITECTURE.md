# Architecture — DATAKEYS KYC API

## Vue d'ensemble

```
┌─────────────────────────────────────────────────────┐
│                     Clients / SDKs                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐           │
│  │ JS/TS    │  │ Python   │  │ Go       │           │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘           │
│       │             │             │                   │
│       └─────────────┼─────────────┘                   │
└─────────────────────┼─────────────────────────────────┘
                      │ HTTPS / TLS
                      ▼
┌─────────────────────────────────────────────────────┐
│               kyc-service (Go / Fiber v3)            │
│                                                       │
│  ┌────────────┐  ┌────────────┐  ┌────────────────┐ │
│  │ Middleware  │  │ Handlers   │  │ Service Layer  │ │
│  │ • Auth     │  │ • KYC     │  │ • Initiate     │ │
│  │ • RateLimit│  │ • Health  │  │ • Process      │ │
│  │ • Idempot. │  │ • Webhooks│  │ • GetStatus    │ │
│  │ • IPAllow  │  │ • SSE     │  │                │ │
│  │ • Metrics  │  │ • Sandbox │  │ • Compliance   │ │
│  └────────────┘  └────────────┘  └───────┬────────┘ │
│                                          │           │
│  ┌───────────────────────────────────────┼─────────┐ │
│  │         Resilience Layer              │         │ │
│  │  ┌──────────┐ ┌──────────┐ ┌──────┐ │         │ │
│  │  │ Retry    │ │ Circuit  │ │ DLQ  │ │         │ │
│  │  │ Backoff  │ │ Breaker  │ │      │ │         │ │
│  │  └──────────┘ └──────────┘ └──────┘ │         │ │
│  └──────────────────────────────────────┼─────────┘ │
│                                         │           │
│  ┌──────────────────────────────────────┼─────────┐ │
│  │         Provider Router             │         │ │
│  │  ┌────────┐ ┌──────────┐ ┌──────┐   │         │ │
│  │  │SmileID │ │Youverify │ │SumSub│   │         │ │
│  │  └────────┘ └──────────┘ └──────┘   │         │ │
│  │  ┌────────┐ ┌────────────────┐       │         │ │
│  │  │ Local  │ │ Sandbox (dev)  │       │         │ │
│  │  └────────┘ └────────────────┘       │         │ │
│  └──────────────────────────────────────┼─────────┘ │
└──────────────────────────────────────────┼───────────┘
                                           │
            ┌──────────────────────────────┼────────────┐
            │         Data Layer           │            │
            │  ┌──────────┐ ┌──────────┐  │            │
            │  │Postgres  │ │ Redis    │  │            │
            │  │(verif's) │ │(sessions)│  │            │
            │  └──────────┘ └──────────┘  │            │
            │  ┌──────────┐ ┌──────────┐  │            │
            │  │ MinIO    │ │ SSE Hub  │  │            │
            │  │(docs)    │ │(notifs)  │  │            │
            │  └──────────┘ └──────────┘  │            │
            └──────────────────────────────┘            │
```

## Piles technologiques

### API Core
| Technologie | Usage |
|-------------|-------|
| **Go 1.25** | Langage serveur |
| **Fiber v3** | Framework HTTP |
| **PostgreSQL 16** | Base de données principale |
| **Redis 7** | Cache, rate limit, SSE Pub/Sub |
| **MinIO** | Stockage documents (S3-compatible) |
| **Prometheus + Grafana** | Métriques et alerting |
| **Jaeger** | Traçage distribué (OpenTelemetry) |
| **Zap** | Logger structuré |

### SDKs
| Langue | Runtime/Compiler | Build |
|--------|-----------------|-------|
| **JavaScript/TypeScript** | Node 18+ | tsup → CJS + ESM + .d.ts |
| **Python** | Python 3.10+ | pyproject.toml (PEP 621) |
| **Go** | Go 1.21+ | go build |

## Flux de vérification

```
Client                     API                          Provider
  │                         │                             │
  │  POST /v1/kyc/initiate  │                             │
  │────────────────────────▶│                             │
  │                         │ 1. Validation pays/doc      │
  │                         │ 2. Vérification doublon     │
  │                         │ 3. AML screening            │
  │                         │ 4. Compliance rules         │
  │                         │ 5. Création session         │
  │◀────────────────────────│                             │
  │  { id, upload_url }     │                             │
  │                         │                             │
  │─── upload document ────▶│                             │
  │                         │                             │
  │                         │─── webhook callback ──────▶│
  │                         │                             │
  │                         │◀── result (async) ────────│
  │                         │                             │
  │                         │ 1. Mise à jour status       │
  │                         │ 2. Notification SSE         │
  │                         │ 3. Webhook callback URL     │
  │                         │ 4. Nettoyage document       │
  │                         │                             │
  │  GET /v1/kyc/status/:id │                             │
  │────────────────────────▶│                             │
  │◀────────────────────────│                             │
  │  { status, score }      │                             │
```

## Décisions d'architecture clés

### 1. Fallback Router + Circuit Breaker
4 providers (SmileID, Youverify, SumSub, Local) sont enrobés de **retry** (3 tentatives, backoff exponentiel ×2), **circuit breaker** (5 échecs → open, timeout 30s) et **fallback** (si le primary échoue, on tente le suivant). La configuration est dynamique par pays via le `CountryRegistry`.

### 2. Sandbox Provider
En mode `env=sandbox`, tous les providers sont remplacés par `SandboxProvider` qui simule des réponses selon le suffixe du numéro de document. 6 scénarios prédéfinis couvrant approved, rejected, sanctions, PEP, expired. Latence simulée 800ms + jitter 200ms. 8 profils africains préchargés dans `internal/seed/testdata.go`.

### 3. SSE via Redis Pub/Sub
`RedisHub` s'abonne à `kyc:sse:*` sur Redis et fane les notifications vers les connexions SSE locales. Permet le scaling horizontal (plusieurs instances de kyc-service) tout en gardant les notifications temps réel.

### 4. Compliance RuleEvaluator
Moteur de règles règlementaires africaines (BCEAO, UEMOA, ECOWAS, GDPR, AMLD). Chaque pays peut référencer une ou plusieurs régulations. Les règles bloquantes (HasBlocked) rejettent la vérification avant même l'envoi au provider.

### 5. API Key auto-détection sandbox/prod
Les SDKs détectent automatiquement l'environnement via le préfixe de la clé (`dk_test_` → sandbox, `dk_live_` → prod). Zéro config.

### 6. Idempotence + Déduplication
- `Idempotency-Key` sur `POST /initiate` → garantit une seule création de session
- Déduplication des webhooks SmileID/SumSub via Redis `SET NX` (TTL 72h)

## Structure des réponses API

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "timestamp": "2025-07-18T12:00:00Z"
}
```

Les erreurs incluent un code machine (`KYC_AUTH_001`, `KYC_SANCTIONS_001`, etc.) pour que les SDKs puissent faire du matching sans parser le message.

## Chemins de répertoires

| Chemin | Contenu |
|--------|---------|
| `services/kyc/` | API Go (cmd, config, internal, migrations, api) |
| `services/kyc/internal/` | Toute la logique métier (12 packages) |
| `services/kyc/migrations/` | Migrations PostgreSQL |
| `services/kyc/api/` | OpenAPI spec |
| `services/kyc/infra/` | Configs Prometheus, Grafana |
| `sdks/javascript/` | SDK JS/TS |
| `sdks/python/` | SDK Python |
| `sdks/go/` | SDK Go |
| `sdks/examples/` | Quickstarts (JS, Python, Go) |
