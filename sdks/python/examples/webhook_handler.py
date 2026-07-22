"""Exemple : recevoir les webhooks DATAKEYS KYC"""
import hmac
import os

WEBHOOK_SECRET = os.getenv("WEBHOOK_HMAC_SECRET")


def verify_webhook_signature(payload: bytes, signature: str) -> bool:
    if not WEBHOOK_SECRET:
        print("WARN: WEBHOOK_HMAC_SECRET non configuré")
        return False
    expected = hmac.new(
        WEBHOOK_SECRET.encode(), payload, "sha256"
    ).hexdigest()
    return hmac.compare_digest(expected, signature)


def webhook_handler(request_body: dict, headers: dict) -> dict:
    signature = headers.get("x-datakeys-signature", "")
    raw_body = str(request_body)

    if not verify_webhook_signature(raw_body.encode(), signature):
        return {"error": "Signature invalide"}, 401

    event = request_body.get("event")
    data = request_body.get("data", {})

    if event == "kyc.approved":
        print(f"✅ KYC approuvé: {data.get('id')}")
        print(f"   Score: {data.get('score')}")
    elif event == "kyc.rejected":
        print(f"❌ KYC rejeté: {data.get('id')}")
        print(f"   Flags: {', '.join(data.get('flags', []))}")
    elif event == "kyc.manual_review":
        print(f"⏳ Review manuelle: {data.get('id')}")
    else:
        print(f"Event inconnu: {event}")

    return {"received": True}, 200
