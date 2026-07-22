"""Exemple : polling du statut d'une vérification existante"""
import sys
import time
from datakeys import Datakeys

dk = Datakeys("dk_test_datakeys_sandbox_001")

if len(sys.argv) < 2:
    print("Usage: python poll_status.py <verification_id>")
    sys.exit(1)

verification_id = sys.argv[1]


def main():
    start = time.time()
    result = dk.kyc.wait_for_completion(
        verification_id,
        max_wait=120,
        interval=3,
        on_poll=lambda v: print(
            f"[{int(time.time() - start)}s] Status: {v.status}  Score: {v.score}"
        ),
    )

    print("\n=== RÉSULTAT FINAL ===")
    print(f"Status: {result.status}")
    print(f"Score:  {result.score}")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"Erreur: {e}")
        sys.exit(1)
