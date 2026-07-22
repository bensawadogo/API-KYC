from __future__ import annotations

import time
import uuid
from typing import Callable, Optional

from ..client import DatakeysClient
from ..errors import KYCError
from ..models import KYCVerification


class KYCResource:
    def __init__(self, client: DatakeysClient):
        self._client = client

    def initiate(
        self,
        phone: str,
        country_code: str,
        doc_type: str,
        full_name: str,
        consent: bool = True,
        doc_number: Optional[str] = None,
        callback_url: Optional[str] = None,
        language: str = "fr",
        idempotency_key: Optional[str] = None,
    ) -> KYCVerification:
        if not consent:
            raise KYCError("Le consentement est obligatoire (BCEAO)", "KYC_VAL_001", 422)

        params: dict = {
            "phone": phone,
            "country_code": country_code.upper(),
            "doc_type": doc_type,
            "full_name": full_name,
            "consent": True,
            "language": language,
        }
        if doc_number:
            params["doc_number"] = doc_number
        if callback_url:
            params["callback_url"] = callback_url

        ikey = idempotency_key or str(uuid.uuid4())
        data = self._client.request("POST", "/v1/kyc/initiate", params, ikey)
        return KYCVerification.from_dict(data)

    def retrieve(self, verification_id: str) -> KYCVerification:
        if not verification_id or not verification_id.strip():
            raise ValueError("verification_id est requis")
        data = self._client.request("GET", f"/v1/kyc/status/{verification_id}")
        return KYCVerification.from_dict(data)

    def wait_for_completion(
        self,
        verification_id: str,
        max_wait: int = 120,
        interval: int = 3,
        on_poll: Optional[Callable[[KYCVerification], None]] = None,
    ) -> KYCVerification:
        deadline = time.time() + max_wait

        while time.time() < deadline:
            v = self.retrieve(verification_id)
            if on_poll:
                on_poll(v)
            if v.is_terminal:
                return v
            time.sleep(interval)

        raise TimeoutError(
            f"Vérification {verification_id} toujours en attente après {max_wait}s"
        )
