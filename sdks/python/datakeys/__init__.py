"""
DATAKEYS KYC SDK
Vérification d'identité panafricaine — 55 pays.

Usage rapide:
    from datakeys import Datakeys

    dk = Datakeys('dk_test_datakeys_sandbox_001')
    v  = dk.kyc.initiate(
        phone='+22670000001',
        country_code='BF',
        doc_type='NATIONAL_ID',
        full_name='Aminata Ouédraogo',
        consent=True,
    )
    result = dk.kyc.wait_for_completion(v.id)
    print(result.status)  # "approved"
"""

from .client import DatakeysClient
from .errors import KYCError
from .models import KYCVerification, Country, CountryDocType
from .resources.kyc import KYCResource
from .resources.countries import CountriesResource

__version__ = "1.0.0"
__all__ = [
    "Datakeys",
    "KYCError",
    "KYCVerification",
    "Country",
    "CountryDocType",
]


class Datakeys:
    def __init__(self, api_key: str, **kwargs):
        client = DatakeysClient(api_key, **kwargs)
        self.kyc = KYCResource(client)
        self.countries = CountriesResource(client)
        self.livemode = client.livemode
