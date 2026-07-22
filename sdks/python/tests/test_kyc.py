import pytest
from datakeys import Datakeys
from tests.fixtures import SANDBOX_KEY


def test_retrieve_empty_id_raises():
    dk = Datakeys(SANDBOX_KEY)
    with pytest.raises(ValueError, match="verification_id est requis"):
        dk.kyc.retrieve("")


def test_wait_for_completion_timeout(monkeypatch):
    dk = Datakeys(SANDBOX_KEY)

    def mock_retrieve(vid):
        return type("obj", (), {"status": "pending", "is_terminal": False})()

    monkeypatch.setattr(dk.kyc, "retrieve", mock_retrieve)

    with pytest.raises(TimeoutError, match="Vérification"):
        dk.kyc.wait_for_completion("ver_test", max_wait=0, interval=1)
