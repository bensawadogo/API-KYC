import { Datakeys, KYCVerification } from '../src';
import { SANDBOX_KEY } from './fixtures';

const MOCK_RESPONSE: KYCVerification = {
  id: 'ver_test', object: 'verification', livemode: false,
  created: Date.now(), status: 'pending', phone_hash: 'abc',
  country_code: 'BF', doc_type: 'NATIONAL_ID', score: 0,
  provider: 'sandbox', flags: [], aml_score: 0,
  is_sanctioned: false, is_pep: false,
};

function mockFetchOk(data: unknown) {
  const body = JSON.stringify({ success: true, data, error: null, timestamp: new Date().toISOString() });
  return jest.spyOn(global, 'fetch').mockResolvedValue({
    ok: true,
    status: 200,
    json: async () => JSON.parse(body),
  } as Response);
}

describe('KYC resource', () => {
  test('initiate appelle le serveur avec les bons paramètres', async () => {
    const fetchMock = mockFetchOk(MOCK_RESPONSE);

    const dk = new Datakeys(SANDBOX_KEY);
    const result = await dk.kyc.initiate({
      phone: '+22670000001',
      country_code: 'BF',
      doc_type: 'NATIONAL_ID',
      full_name: 'Test',
      consent: true,
    });

    expect(result.id).toBe('ver_test');
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/v1/kyc/initiate'),
      expect.objectContaining({
        method: 'POST',
        headers: expect.objectContaining({ 'X-API-Key': SANDBOX_KEY }),
      }),
    );

    fetchMock.mockRestore();
  });

  test('retrieve lève erreur si id vide', async () => {
    const dk = new Datakeys(SANDBOX_KEY);
    await expect(dk.kyc.retrieve('')).rejects.toThrow('verificationId est requis');
  });

  test('waitForCompletion timeout sur statut non terminal', async () => {
    const dk = new Datakeys(SANDBOX_KEY);
    const mockRetrieve = jest.spyOn(dk.kyc, 'retrieve');
    mockRetrieve.mockResolvedValue({
      id: 'ver_test',
      object: 'verification',
      livemode: false,
      created: Date.now(),
      status: 'pending',
      phone_hash: 'abc',
      country_code: 'BF',
      doc_type: 'NATIONAL_ID',
      score: 0,
      provider: 'sandbox',
      flags: [],
      aml_score: 0,
      is_sanctioned: false,
      is_pep: false,
    });

    await expect(dk.kyc.waitForCompletion('ver_test', { maxWaitMs: 100, intervalMs: 50 })).rejects.toThrow(
      'Timeout',
    );

    mockRetrieve.mockRestore();
  });
});
