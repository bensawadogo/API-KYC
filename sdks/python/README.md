# datakeys-kyc

SDK officiel DATAKEYS KYC · Vérification d'identité panafricaine
55 pays · Conforme BCEAO · Sandbox inclus

## Installation

```bash
pip install datakeys-kyc
```

## 5 minutes pour intégrer

```python
from datakeys import Datakeys

dk = Datakeys('dk_test_datakeys_sandbox_001')

verification = dk.kyc.initiate(
    phone='+22670000001',
    country_code='BF',
    doc_type='NATIONAL_ID',
    full_name='Aminata Ouédraogo',
    consent=True,
)

result = dk.kyc.wait_for_completion(verification.id)
print(result.status)  # "approved"
print(result.score)   # 0.95
```

## Gestion des erreurs

```python
from datakeys import Datakeys, KYCError

try:
    dk.kyc.initiate(...)
except KYCError as e:
    print(e.code)    # "KYC_AUTH_002"
    print(e.status)  # 401
    if e.is_rate_limit():
        # Attendre avant de réessayer
        pass
```

## Profils sandbox

| phone          | pays | scénario  |
|---------------|------|-----------|
| +22670000001  | BF   | approved  |
| +233200001111 | GH   | rejected  |
| +254700002222 | KE   | sanctions |
| +2250700003333| CI   | pep       |
