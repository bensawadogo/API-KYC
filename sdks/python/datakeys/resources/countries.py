from __future__ import annotations

from typing import List

from ..client import DatakeysClient
from ..models import Country, CountryDocType


class CountriesResource:
    def __init__(self, client: DatakeysClient):
        self._client = client

    def list(self) -> List[Country]:
        data = self._client.request("GET", "/v1/kyc/countries")
        return [Country.from_dict(c) for c in data]

    def doc_types(self, country_code: str) -> List[CountryDocType]:
        data = self._client.request(
            "GET", f"/v1/kyc/countries/{country_code.upper()}/doctypes"
        )
        return [CountryDocType(**dt) for dt in data]
