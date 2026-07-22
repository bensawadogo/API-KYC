from __future__ import annotations
from typing import Optional, Any


class KYCError(Exception):
    def __init__(
        self,
        message: str,
        code: str,
        status: int,
        raw: Optional[Any] = None,
    ):
        super().__init__(message)
        self.code = code
        self.status = status
        self.raw = raw

    def is_auth_error(self) -> bool:
        return self.code.startswith("KYC_AUTH")

    def is_rate_limit(self) -> bool:
        return self.code.startswith("KYC_RATE")

    def is_validation(self) -> bool:
        return self.code.startswith("KYC_VAL")

    def is_sanctions(self) -> bool:
        return self.code == "KYC_AML_SANCTION"

    def is_server_error(self) -> bool:
        return self.status >= 500

    def __repr__(self) -> str:
        return f"KYCError(code={self.code!r}, status={self.status}, message={str(self)!r})"
