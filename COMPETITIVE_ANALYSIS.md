# Analyse Concurrentielle DATAKEYS KYC

> Analyse des SDKs concurrence (SmileID, Youverify, Prembly/IdentityPass, Dojah, Stripe)
> pour identifier les forces à copier et les lacunes à exploiter.

---

## 1. SmileID — smile-identity-core-js

**GitHub**: github.com/smileidentity/smile-identity-core-js (7★)
**SDKs**: JS/TS, Python, Java, iOS, Android, React Native, Flutter, KMP

### Architecture SDK
```
WebApi            → submit_job(), get_job_status(), get_web_token()
IDApi             → submit_job() pour Enhanced/Basic KYC, Business Verification
Signature         → generate_signature(), confirm_signature() — HMAC partenaire
Utilities         → get_job_status() utilitaire
```

### ✅ À COPIER
1. **Classes spécialisées séparées** : WebApi (documents + images), IDApi (KYC sans image), Signature (HMAC). Chaque classe a un périmètre clair. Implémenter `DKSignature` class en JS pour les webhooks.
2. **`get_web_token()`** : Token éphémère pour Hosted Web Integration. Ajouter `getWebToken()` → utile pour les intégrations front-end où le client ne doit jamais voir la clé API.
3. **HMAC signature obligatoire** : `generate_signature` + `confirm_signature` sur TOUS les appels. DATAKEYS le fait côté serveur mais le SDK devrait le supporter côté client.
4. **Documentation par méthode** : Chaque méthode publique a un lien vers la doc en ligne. Copier le pattern.
5. **Support mobile natif** : iOS, Android, React Native, Flutter. DATAKEYS manque tout le mobile — priorité #1 pour les partenaires fintech.

### ❌ LACUNES DE SMILEID
- SDKs séparés par langage (JS != Java = implémentations différentes) → pas de cohérence
- Pas de retry automatique, pas de timeout configurable
- Signature requise complexe, friction à l'intégration
- `submit_job()` est un fourre-tout monolithique (30+ paramètres)
- zéro auto-détection sandbox/prod

---

## 2. Stripe — stripe-node

**GitHub**: github.com/stripe/stripe-node (4 475★)
**Le gold standard mondial des SDKs**

### Architecture SDK
```
Stripe               → factory client, config, _prepResources()
StripeResource       → base class pour TOUTES les resources
  ├── Customers      → resource.create(), retrieve(), list(), update(), del()
  ├── Charges        → idem
  └── ...
RequestSender        → HTTP, retry, telemetry, auth
PlatformFunctions    → abstraction Node.js / Web Worker / Deno
  ├── NodeHttpClient → http/https natif
  ├── FetchHttpClient→ fetch API
  ├── NodeCrypto     → crypto natif
  └── SubtleCrypto   → Web Crypto API
```

### ✅ À COPIER — CE QUI MANQUE À DATAKEYS
| Feature Stripe | DATAKEYS aujourd'hui | Action |
|---|---|---|
| **`StripeResource` base class** | Resources JS/Python/Go séparées | Ajouter une base class `DKResource` avec `request()`, `_path()`, `_method()` |
| **Auto-pagination** `autoPagingToArray()` | Pas de pagination | Ajouter `.list(autoPaginate: true)` |
| **Platform abstraction** (Node/Worker/Deno) | `fetch` en dur — pas de fallback | Ajouter `PlatformFunctions` pour Node HTTP natif + fetch |
| **Telemetry** (`X-Stripe-Client-Telemetry`) | Rien | Ajouter timing des requêtes dans les headers |
| **`_prepResources()`** auto-attach | `Datakeys` init manuel | Automatiser l'attachement des resources |
| **`static extend()`** pour resources custom | Pas possible | Permettre aux users d'étendre le SDK |
| **`webhookEndpoints.constructEvent()`** | Pas dans le SDK | Ajouter helper webhook intégré |
| **Config object** plutôt que kwargs | API key string + rest params | Créer `DatakeysConfig` complet |
| **Tests avec `nock` / monkey-patch support** | Tests unitaires de base | Pattern d'import qui supporte le mock |

### ✅ À COPIER — DÉJÀ EN PLACE (bon)
- Resource pattern (kyc, countries) → oui, similaire
- Exponential backoff + jitter → oui 500ms base ×2
- Types TypeScript inline → oui
- Error class avec helpers (isAuthError, etc.) → oui

---

## 3. Youverify

**SDKs**: Browser JS (npm youverify-sdk), Android natif
**Approche**: Widgets front-end (vForm, Liveness, Document Capture)

### ✅ À COPIER
1. **vForm (widget embarqué)** : Un composant React prêt à l'emploi pour capturer les documents + selfie. DATAKEYS devrait avoir `@datakeys/kyc-widget` pour React.
2. **Liveness Check module séparé** : `LivenessCheckModule.Builder()` pattern builder. Propre.
3. **Document Capture module** : `DocumentCaptureModule.Builder()` — même pattern. Ajouter un module capture dans le SDK JS.
4. **Customisation UI** : `Builder` pattern avec thème, couleurs, messages personnalisés.

### ❌ LACUNES DE YOUVERIFY
- SDK JS obsolète (1.0.13, jan 2022 — pas mis à jour depuis 2022)
- zéro TypeScript (JS vanilla)
- Pas de SDK Python/Go
- Pas de retry
- Pas d'auto-détection sandbox/prod
- Documentation éparse

---

## 4. Prembly / IdentityPass

**GitHub**: github.com/prembly (1-2★ par repo)
**SDKs**: Python, JS, Flutter (GPL v3)
**Marché**: Nigeria d'abord, puis Ghana, Kenya, Rwanda, SA, Uganda, Sierra Leone

### Approche architecturale
```python
from pyprembly.data.nigeria import DataVerification
# Méthodes par pays:
DataVerification().bank_account_verification()
DataVerification().nin_lookup()
DataVerification().voters_card_lookup()
```

### ✅ À COPIER — ORGANISATION PAR PAYS
1. **Per-country modules** : `Datakeys.BF.verify()`, `Datakeys.NG.verify()`. Chaque pays a SES méthodes spécifiques (NIN pour Nigeria, SSNIT pour Ghana). DATAKEYS a une API générique — c'est bien pour l'unification MAIS trop abstrait pour les devs qui veulent appeler `verifyBVN()` ou `verifyNIN()`. Solution : garder l'API générique ET ajouter des helpers par pays.
2. **`mashupService`** : Appel à plusieurs bases en un seul call. Absolument à copier.
3. **Enum exhaustif par pays** : Pour chaque pays, la liste EXACTE des méthodes disponibles. DATAKEYS devrait avoir une page /docs avec `country.methods[]`.

### ❌ LACUNES DE PREMBLY
- GPL v3 (pas compatible avec logiciel propriétaire)
- Qualité de code médiocre (pas de typage, doc incomplète, pas de CI visible)
- Chaque pays = module séparé = duplication massive
- Pas de retry, pas de timeout, pas de webhook helper

---

## 5. Dojah

**SDKs**: React widget, API REST
**Approche**: `<DojahWidget />` composant React clé en main

### ✅ À COPIER
1. **Widget React** : `<DatakeysWidget />` — 5 lignes pour intégrer la capture KYC complète. Copier l'API `response(type, data)` callback.
2. **`govData` param** : Passage des données gouvernementales (BVN, NIN, DL) directement sans capture document. DATAKEYS devrait supporter `gov_id` en param.
3. **Reference ID + callback** : `referenceId` + `response(type, data)` → pattern propre.

---

## SYNTHÈSE — PRIORITÉ D'ACTIONS

### P0 — CRITICAL (manque compétitif immédiat)
1. **Widget Web/React** `@datakeys/kyc-widget` → concurrence Youverify vForm + Dojah
2. **Helpers par pays** : `dk.verifyBVN(nin, dob)`, `dk.verifyNIN(nin)`, etc.
3. **Support mobile natif** : React Native SDK (priorité #1 pour les fintechs)

### P1 — HAUTE VALEUR
4. **Platform abstraction** : Node natif + Fetch + timeout+retry cohérent (comme Stripe)
5. **Auto-pagination** sur `.list()` → `for await (const c of dk.countries.list({autoPaginate: true}))`
6. **Webhook helper dans le SDK** : `dk.webhooks.constructEvent(payload, sig, secret)` comme Stripe
7. **Telemetry headers** : `X-DataKeys-Latency`, version SDK, runtime info

### P2 — QUALITÉ
8. **Base class `DKResource`** partagée entre resources JS/Python/Go
9. **`getWebToken()`** pour hosted integration (front-end sans clé API)
10. **Console developer** : Mode debug avec log des requêtes (`DK_DEBUG=true`)
11. **Exemples par pays** : Nigeria, Ghana, Kenya, Côte d'Ivoire, Burkina → quickstart pour chaque
12. **Bundle size** : Tree-shaking ready (ESM), vérifier la taille avec `tsup`

### P3 — DIFFÉRENCIATION EXPLOITABLE
- **Conformité BCEAO intégrée** : `consent: true` obligatoire, audit trail. Aucun concurrent ne fait ça.
- **Routing multi-provider automatique** : Si SmileID down → Youverify → SumSub → local. Stripe ne fait pas ça.
- **55 pays africains** : Couverture la plus large. Prembly fait 7 pays. SmileID ~20. Youverify ~30.
- **Sandbox 8 profils africains** : Prêt à l'emploi. Aucun concurrent n'a 8 profils préchargés.
- **Régulations africaines** : 14 règles, 9 ont été ajoutées récemment. C'est un argument de vente unique pour les banques centrales.

---

## CE QUE DATAKEYS FAIT MIEUX QUE TOUS

| Critère | DATAKEYS | SmileID | Youverify | Prembly | Stripe |
|---|---|---|---|---|---|
| **Délai intégration** | 5 min | 30 min+ | 48h annoncé | 15 min | 10 min |
| **Languages SDK** | JS+Py+Go | JS+Java | JS only | JS+Py+Flutter | 7 langages |
| **Auto sandbox/prod** | Oui (préfixe) | Non | Non | Non | Oui (clé) |
| **Retry automatique** | Oui (3x backoff) | Non | Non | Non | Oui |
| **Sandbox profiles** | 8 profils africains | 1-2 profils | 1 profil | 1 profil | N/A |
| **Coverage Afrique** | 55 pays | ~20 | ~30 | 7 | N/A |
| **Conformité BCEAO** | Oui (14 règles) | Non | Non | Non | N/A |
| **Multi-provider** | Oui (4 providers) | Non | Non | Non | N/A |
| **Timeout configurable** | Oui (30s) | Non | Non | Non | Oui (80s) |

---

## PROMPT POUR PROCHAINE SESSION

```
Tu es un ingénieur senior spécialisé developer experience.
Implémente les features suivantes dans nos SDKs DATAKEYS KYC
en t'inspirant de l'architecture stripe-node.

P0 — URGENT:
1. Widget React @datakeys/kyc-widget : composant <DatakeysWidget /> 
   inspiré de Dojah widget + Youverify vForm
2. Helpers par pays : dk.verifyBVN(), dk.verifyNIN(), etc.
   inspiré de Prembly/IdentityPass per-country modules
3. SDK React Native : natif iOS + Android

P1 — HAUTE VALEUR:
4. Platform abstraction (Node http + fetch) inspiré stripe-node
5. Auto-pagination sur dk.countries.list({autoPaginate: true})
6. dk.webhooks.constructEvent(payload, sig, secret) helper
7. Telemetry headers X-DataKeys-Latency

P2 — QUALITÉ:
8. Base class DKResource partagée entre JS/Python/Go
9. getWebToken() pour hosted integration
10. Mode debug (DK_DEBUG=true)

Le code doit être buildable ET testé.
Respecter les conventions existantes dans sdks/javascript, sdks/python, sdks/go.
```
