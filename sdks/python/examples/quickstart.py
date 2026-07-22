"""Quickstart DATAKEYS KYC — 5 minutes pour intégrer"""
from datakeys import Datakeys, KYCError

dk = Datakeys("dk_test_datakeys_sandbox_001")
print(f"Mode: {'PRODUCTION' if dk.livemode else 'SANDBOX'}")

try:
    countries = dk.countries.list()
    print(f"\n{len(countries)} pays africains supportés")
    bf = next((c for c in countries if c.code == "BF"), None)
    if bf:
        print(f"Burkina Faso: {len(bf.doc_types)} types de docs")

    print("\n--- Initiation vérification ---")
    verification = dk.kyc.initiate(
        phone="+22670000001",
        country_code="BF",
        doc_type="NATIONAL_ID",
        doc_number="B1230000",
        full_name="Aminata Ouédraogo",
        consent=True,
    )
    print(f"ID:        {verification.id}")
    print(f"Statut:    {verification.status}")
    print(f"Provider:  {verification.provider}")
    if verification.upload_url:
        print(f"Upload URL: {verification.upload_url}")

    print("\n--- Attente résultat (sandbox = rapide) ---")
    result = dk.kyc.wait_for_completion(verification.id, max_wait=60, interval=2)

    print("\n=== RÉSULTAT FINAL ===")
    print(f"Statut:     {result.status}")
    print(f"Score:      {result.score}")
    print(f"Provider:   {result.provider}")
    print(f"Sanctionné: {result.is_sanctioned}")
    print(f"PEP:        {result.is_pep}")
    if result.flags:
        print(f"Flags:      {', '.join(result.flags)}")

except KYCError as e:
    print(f"KYC Error [{e.code}]: {e}")
except Exception as e:
    print(f"Erreur: {e}")
