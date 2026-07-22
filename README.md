# DATAKEYS KYC Platform

API de vérification d'identité panafricaine.
Couvre 55 pays africains avec routing intelligent multi-provider.

## Démarrage rapide

```bash
git clone https://github.com/datakeys/kyc-service
cd kyc-service/services/kyc

# Copie le fichier d'environnement
cp .env.example .env
# Édite .env avec tes vraies valeurs
# JAMAIS committer .env

# Lance l'environnement complet
docker compose up -d

# Seed la clé API de test
docker compose exec postgres psql -U kyc_user -d kyc_db -c \
  "INSERT INTO api_keys (client_name, key_hash, key_prefix, \
   scopes, rate_limit) VALUES ('sandbox-test', \
   encode(digest('dk_test_datakeys_sandbox_001','sha256'),'hex'), \
   'dk_test_datake', \
   ARRAY['kyc:initiate','kyc:status','kyc:admin'], 1000) \
   ON CONFLICT DO NOTHING;"

# Test rapide
curl http://localhost:8081/health/ready
curl http://localhost:8081/sandbox/profiles
```

## Variables d'environnement

Voir `.env.example` — copier vers `.env` et remplir.
Ne jamais committer `.env`.

## Documentation API

http://localhost:8081/docs

## Architecture

- Go 1.25 + Fiber v3
- PostgreSQL 16 + Redis 7 + MinIO
- Providers : SmileID / Youverify / SumSub / Local sandbox
- 55 pays africains, conformité BCEAO
