from datakeys import Datakeys

dk = Datakeys('dk_test_datakeys_sandbox_001')

# Initier une vérification
v = dk.kyc.initiate(
    phone='+22670000001',
    country_code='BF',
    doc_type='NATIONAL_ID',
    doc_number='B1230000',
    full_name='Aminata Ouédraogo',
    consent=True,
)
print(f"ID: {v.id}")
print(f"Statut: {v.status}")

# Attendre le résultat
result = dk.kyc.wait_for_completion(v.id)
print(f"Résultat final: {result.status} (score: {result.score})")
