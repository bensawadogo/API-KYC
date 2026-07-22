from __future__ import annotations

import json
import re
import time
import urllib.error
import urllib.request
from typing import Any, Optional, TypeVar

from .errors import KYCError

SANDBOX_URL = "http://localhost:8081"
PRODUCTION_URL = "https://api.datakeys.africa"
SDK_VERSION = "1.0.0"

T = TypeVar("T")


def _extract_code(error: Optional[str]) -> str:
    if not error:
        return "KYC_UNKNOWN"
    m = re.search(r"KYC_[A-Z0-9_]+", error)
    return m.group(0) if m else "KYC_UNKNOWN"


def _exponential_backoff(attempt: int) -> float:
    import random

    delay = 0.5 * (2**attempt)
    jitter = random.uniform(0, delay * 0.3)
    return min(delay + jitter, 10.0)


class DatakeysClient:
    def __init__(
        self,
        api_key: str,
        base_url: Optional[str] = None,
        timeout: int = 30,
        max_retries: int = 3,
    ):
        if not api_key or not api_key.strip():
            raise KYCError(
                "API key manquante. Obtenir une clé sur https://dashboard.datakeys.africa",
                "KYC_AUTH_001",
                401,
            )

        self.api_key = api_key
        self.livemode = api_key.startswith("dk_live_")
        self.timeout = timeout
        self.max_retries = max_retries
        self.base_url = base_url or (PRODUCTION_URL if self.livemode else SANDBOX_URL)

    def request(
        self,
        method: str,
        path: str,
        body: Optional[dict] = None,
        idempotency_key: Optional[str] = None,
    ) -> Any:
        headers = {
            "Content-Type": "application/json",
            "X-API-Key": self.api_key,
            "X-SDK-Version": SDK_VERSION,
            "X-SDK-Lang": "python",
            "X-SDK-Livemode": str(self.livemode).lower(),
        }
        if idempotency_key:
            headers["Idempotency-Key"] = idempotency_key

        data = json.dumps(body).encode() if body else None
        url = f"{self.base_url}{path}"
        last_error: Optional[Exception] = None

        for attempt in range(self.max_retries):
            try:
                req = urllib.request.Request(url, data=data, headers=headers, method=method)
                with urllib.request.urlopen(req, timeout=self.timeout) as resp:
                    result = json.loads(resp.read())

                if not result.get("success"):
                    raise KYCError(
                        result.get("error", "Erreur API"),
                        _extract_code(result.get("error")),
                        200,
                        result,
                    )

                return result.get("data")

            except urllib.error.HTTPError as e:
                raw_body = e.read().decode(errors="replace")
                try:
                    err_json = json.loads(raw_body)
                except Exception:
                    err_json = {}

                msg = err_json.get("error", "Erreur API")
                code = _extract_code(msg)

                if e.code < 500:
                    raise KYCError(msg, code, e.code, err_json)

                last_error = KYCError("Erreur serveur", "KYC_SERVER_ERR", e.code)
                time.sleep(_exponential_backoff(attempt))

            except urllib.error.URLError as e:
                last_error = KYCError(f"Erreur réseau: {e.reason}", "KYC_NETWORK", 0)
                if attempt < self.max_retries - 1:
                    time.sleep(_exponential_backoff(attempt))

        raise last_error or KYCError("Échec après tous les retries", "KYC_NETWORK", 0)
