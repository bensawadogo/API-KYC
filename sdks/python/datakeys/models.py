from __future__ import annotations
from dataclasses import dataclass, field
from typing import List, Optional, Literal

DocType = Literal[
    "NATIONAL_ID", "PASSPORT", "DRIVERS_LICENSE",
    "VOTER_CARD", "RESIDENCE_PERMIT",
]

VerificationStatus = Literal[
    "pending", "processing", "approved",
    "rejected", "manual_review", "expired",
]

Region = Literal[
    "WEST_AFRICA", "NORTH_AFRICA", "EAST_AFRICA",
    "CENTRAL_AFRICA", "SOUTHERN_AFRICA",
]


@dataclass
class KYCVerification:
    id: str
    object: str
    livemode: bool
    created: int
    status: VerificationStatus
    phone_hash: str
    country_code: str
    doc_type: str
    score: float
    provider: str
    flags: List[str]
    aml_score: float
    is_sanctioned: bool
    is_pep: bool
    upload_url: Optional[str] = None
    expires_in: Optional[int] = None
    processed_at: Optional[str] = None

    @classmethod
    def from_dict(cls, d: dict) -> KYCVerification:
        known = {k: v for k, v in d.items() if k in cls.__dataclass_fields__}
        return cls(**known)

    @property
    def is_approved(self) -> bool:
        return self.status == "approved"

    @property
    def is_rejected(self) -> bool:
        return self.status == "rejected"

    @property
    def is_terminal(self) -> bool:
        return self.status in {"approved", "rejected", "manual_review", "expired"}


@dataclass
class CountryDocType:
    code: str
    name: str
    pattern: Optional[str] = None


@dataclass
class Country:
    code: str
    name: str
    phone_prefix: str
    region: Region
    doc_types: List[CountryDocType]
    provider: str

    @classmethod
    def from_dict(cls, d: dict) -> Country:
        return cls(
            code=d["code"],
            name=d["name"],
            phone_prefix=d["phone_prefix"],
            region=d["region"],
            doc_types=[CountryDocType(**dt) for dt in d.get("doc_types", [])],
            provider=d["provider"],
        )
