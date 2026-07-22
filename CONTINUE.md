# Continue — Instructions pour une IA

Ce document est conçu pour qu'une autre IA (ou un nouveau contributeur) puisse prendre le projet en main immédiatement.

## Résumé du projet

API panafricaine de vérification d'identité (KYC) avec SDKs multi-langues. Un client initie une vérification, upload un document, et reçoit le résultat en temps réel (SSE) ou par webhook. L'API route vers 4 providers d'identité avec fallback automatique, circuit breaker, et screening AML.

## Stack en un coup d'œil

```
API:   Go 1.25 + Fiber v3 + PostgreSQL 16 + Redis 7
SDKs:  JS/TS (tsup), Python (pyproject.toml), Go
Infra: Docker Compose (Postgres, Redis, MinIO, Prometheus, Grafana, Jaeger)
CI/CD: GitHub Actions (Go test, npm build, pip install, Docker build & push)
```

## État actuel (tout ✅)

| Bloc | Statut | Tests |
|------|--------|-------|
| Core API (handlers, service, middleware, auth) | ✅ | 21 packages, tous verts |
| Providers (SmileID, Youverify, SumSub, Local) | ✅ | mockés + tests unitaires |
| Sandbox complet (provider + handler + seed) | ✅ | tests provider, handler, seed |
| Resilience (retry, circuit breaker, fallback, DLQ) | ✅ | tests unitaires |
| Compliance (RuleEvaluator, SAR, consent) | ✅ | tests unitaires |
| SSE Redis Hub | ✅ | tests miniredis |
| SDK JavaScript/TypeScript | ✅ | `npm run build` → CJS + ESM + .d.ts |
| SDK Python | ✅ | `pip install -e .` |
| SDK Go | ✅ | `go build ./...` + `go vet ./...` |
| Docker Compose (full stack) | ✅ | Makefile + smoke tests |
| CI/CD | ✅ | GitHub Actions workflow |

## Pour commencer

```bash
# 1. Lancer l'API en sandbox
cd services/kyc
set POSTGRES_URL=postgres://kyc_user:kyc_secret@localhost:5432/kyc_db?sslmode=disable
set SERVER_ENV=sandbox
go run ./cmd/

# 2. Tester avec le SDK Go (autre terminal)
cd sdks/go
go run ../examples/go/main.go
```

## Comment une IA doit continuer le projet

### 1. Toujours valider avant de modifier
```bash
cd services/kyc && go test ./... -timeout 60s -count=1
cd sdks/go && go build ./... && go vet ./...
cd sdks/javascript && npm run build
cd sdks/python && python -c "from datakeys import Datakeys; print('ok')"
```

### 2. Conventions du code Go
- **Organisation** : `internal/` contient 12 packages métier. Pas de dépendances circulaires entre packages.
- **Constructeurs** : `New<Type>(cfg, logger, ...)` pattern. Toujours passer `*zap.Logger` explicitement.
- **Interfaces** : Définies dans `internal/interfaces.go`. Les providers implémentent `IdentityProvider`.
- **Erreurs** : Renvoyées en valeur de retour (pas de panic). Logguées au niveau appelant.
- **Noms de fichiers** : `snake_case.go` dans `internal/`.
- **Tests** : `_test.go` à côté du fichier testé. Utiliser `testing.T` standard (pas de framework externe).

### 3. Conventions SDK
- **JS/TS** : `src/index.ts` exporte la classe principale. `src/resources/` par endpoint group.
- **Python** : `datakeys/__init__.py` exporte la classe principale. Modèles = dataclasses avec `from_dict()`.
- **Go** : `datakeys.go` = client + config. Fichiers séparés par ressource. `apiResponse[T]` générique.
- **Tous les SDKs** : auto-détection sandbox/prod, timeout 30s, retry 3× backoff, erreurs typées KYC_*.

### 4. Tâches prioritaires restantes
1. **Pagination** sur `GET /v1/kyc/countries` (si > ~50 pays)
2. **Document upload** : endpoint `PUT /v1/kyc/upload/:id` avec validation de taille/type
3. **Rate limiting** : configurable par API key (actuellement global)
4. **Webhook signing** : HMAC signature pour les callbacks sortants
5. **Admin endpoints** : CRUD API keys, stats dashboard
6. **Internationalisation** : messages d'erreur localisés (fr, en, ar, sw, ha)
7. **Tests d'intégration** bout en bout avec docker-compose et seed data

### 5. Anti-patterns à éviter
- Ne pas ajouter de dépendances lourdes (pas de ORM, pas de framework de test)
- Ne pas exposer de secrets/logs sensibles (toujours hasher les clés API, masquer les numéros de document)
- Ne pas supprimer le SandboxProvider (utilisé par les SDKs en mode test)
- Ne pas renommer les codes d'erreur `KYC_*` sans mettre à jour les 3 SDKs
- Ne pas supprimer la compatibilité ascendante des réponses `{ success, data, error, timestamp }`

## Codes d'erreur standardisés

| Code | HTTP | Signification |
|------|------|---------------|
| `KYC_AUTH_001` | 401 | Clé API manquante |
| `KYC_AUTH_002` | 401 | Clé API invalide |
| `KYC_AUTH_003` | 403 | Scope insuffisant |
| `KYC_RATE_001` | 429 | Rate limit dépassé |
| `KYC_VAL_001` | 422 | Données invalides |
| `KYC_VAL_002` | 422 | Type doc non supporté |
| `KYC_VAL_003` | 422 | Pays non supporté |
| `KYC_AML_001` | 403 | Sanctions match |
| `KYC_CONSENT_001` | 403 | Consentement requis |
| `KYC_SERVER_ERR` | 500 | Erreur serveur (retry) |
| `KYC_NETWORK` | 0 | Erreur réseau SDK |

## Dépendances externes

| Dépendance | Usage | Alternative |
|-----------|-------|-------------|
| SmileID API | Provider principal | Youverify, SumSub, Local |
| OpenSanctions | AML screening | LocalAMLProvider |
| PostgreSQL | Stockage persistant | — |
| Redis | Cache, rate limit, SSE | — |
| MinIO | Stockage documents | MemoryStorage |

## Tests

```bash
# 21 suites de tests dans l'API
cd services/kyc && go test ./... -v -count=1

# 3 SDKs
cd sdks/go && go test ./... && go vet ./...
cd sdks/javascript && npm test
cd sdks/python && python -m pytest
```

Tous les tests sont en mémoire ou mockés (miniredis, pas de Redis réel nécessaire). Le sandbox provider peut être testé sans connexion externe.
