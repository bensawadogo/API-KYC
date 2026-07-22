# @datakeys/kyc

SDK officiel DATAKEYS KYC · Vérification d'identité panafricaine
55 pays · Conforme BCEAO · Sandbox inclus

## Installation

```bash
npm install @datakeys/kyc
```

## 5 minutes pour intégrer

```js
import Datakeys from '@datakeys/kyc';

const dk = new Datakeys('dk_test_datakeys_sandbox_001');

const verification = await dk.kyc.initiate({
  phone:        '+22670000001',
  country_code: 'BF',
  doc_type:     'NATIONAL_ID',
  full_name:    'Aminata Ouédraogo',
  consent:      true,
});

const result = await dk.kyc.waitForCompletion(verification.id);
console.log(result.status); // "approved"
console.log(result.score);  // 0.95
```

## Gestion des erreurs

```js
import Datakeys, { KYCError } from '@datakeys/kyc';

try {
  await dk.kyc.initiate({ ... });
} catch (err) {
  if (err instanceof KYCError) {
    console.log(err.code);   // "KYC_AUTH_002"
    console.log(err.status); // 401
    if (err.isRateLimit()) {
      // Attendre avant de réessayer
    }
  }
}
```

## Profils sandbox

| phone          | pays | scénario  |
|---------------|------|-----------|
| +22670000001  | BF   | approved  |
| +233200001111 | GH   | rejected  |
| +254700002222 | KE   | sanctions |
| +2250700003333| CI   | pep       |

Voir tous les profils : `curl http://localhost:8081/sandbox/profiles`
