import pytest
from datakeys import Datakeys, KYCError
from tests.fixtures import SANDBOX_KEY


def test_empty_api_key_raises_error():
    with pytest.raises(KYCError) as exc:
        Datakeys("")
    assert exc.value.code == "KYC_AUTH_001"


def test_sandbox_key_livemode_false():
    dk = Datakeys(SANDBOX_KEY)
    assert dk.livemode is False


def test_live_key_livemode_true():
    dk = Datakeys("dk_live_fakekey")
    assert dk.livemode is True


def test_base_url_sandbox_default():
    dk = Datakeys(SANDBOX_KEY)
    assert "localhost:8081" in dk.kyc._client.base_url


def test_base_url_custom_override():
    dk = Datakeys(SANDBOX_KEY, base_url="https://custom.api.test")
    assert dk.kyc._client.base_url == "https://custom.api.test"
