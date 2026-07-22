import { DatakeysClient } from '../src/client';
import {
  NigeriaVerification,
  BurkinaVerification,
  GhanaVerification,
  KenyaVerification,
  SenegalVerification,
  IvoireVerification,
  MarocVerification,
  CountryHelpers,
} from '../src/resources/country-helpers';
import { SANDBOX_KEY } from './fixtures';

const mockRequest = jest.fn().mockResolvedValue({
  id: 'ver_test',
  object: 'verification',
  livemode: false,
  created: Date.now(),
  status: 'pending',
  phone_hash: 'xxx',
  country_code: 'NG',
  doc_type: 'NATIONAL_ID',
  score: 0,
  provider: 'sandbox',
  flags: [],
  aml_score: 0,
  is_sanctioned: false,
  is_pep: false,
});
const mockClient = { request: mockRequest } as unknown as DatakeysClient;

beforeEach(() => mockRequest.mockClear());

describe('NigeriaVerification', () => {
  const ng = new NigeriaVerification(mockClient);

  test('verifyNIN appelle /initiate avec country_code NG', async () => {
    await ng.verifyNIN('+2348100000001', '12345678901', 'Test User');
    expect(mockRequest).toHaveBeenCalledWith(
      'POST',
      '/v1/kyc/initiate',
      expect.objectContaining({
        country_code: 'NG',
        doc_type: 'NATIONAL_ID',
        doc_number: '12345678901',
      }),
    );
  });

  test('verifyNIN rejette NIN format invalide', async () => {
    await expect(ng.verifyNIN('+2348100000001', 'BADNIN', 'Test')).rejects.toThrow('NIN invalide');
  });

  test('verifyBVN rejette BVN format invalide', async () => {
    await expect(ng.verifyBVN('+2348100000001', '123', 'Test')).rejects.toThrow('BVN invalide');
  });

  test('verifyVoterCard appelle /initiate avec VOTER_CARD', async () => {
    await ng.verifyVoterCard('+2348100000001', 'VOTER123', 'Test User');
    expect(mockRequest).toHaveBeenCalledWith(
      'POST',
      '/v1/kyc/initiate',
      expect.objectContaining({ doc_type: 'VOTER_CARD' }),
    );
  });
});

describe('BurkinaVerification', () => {
  const bf = new BurkinaVerification(mockClient);

  test('verifyCNIB appelle /initiate avec country_code BF', async () => {
    await bf.verifyCNIB('+22670000001', 'B1234567', 'Aminata O.');
    expect(mockRequest).toHaveBeenCalledWith(
      'POST',
      '/v1/kyc/initiate',
      expect.objectContaining({ country_code: 'BF', doc_number: 'B1234567' }),
    );
  });

  test('verifyCNIB rejette format invalide', async () => {
    await expect(bf.verifyCNIB('+22670000001', '1234567', 'Test')).rejects.toThrow('CNIB invalide');
  });

  test('verifyPassport appelle /initiate avec PASSPORT', async () => {
    await bf.verifyPassport('+22670000001', 'BP123456', 'Aminata O.');
    expect(mockRequest).toHaveBeenCalledWith(
      'POST',
      '/v1/kyc/initiate',
      expect.objectContaining({ doc_type: 'PASSPORT' }),
    );
  });
});

describe('GhanaVerification', () => {
  const gh = new GhanaVerification(mockClient);

  test('verifyGhanaCard rejette format invalide', async () => {
    await expect(gh.verifyGhanaCard('+233200001111', 'BADCARD', 'Test')).rejects.toThrow(
      'Ghana Card invalide',
    );
  });

  test('verifyGhanaCard accepte format GHA-XXXXXXXXX-X', async () => {
    await expect(
      gh.verifyGhanaCard('+233200001111', 'GHA-123456789-0', 'Kwame A.'),
    ).resolves.toBeDefined();
  });
});

describe('KenyaVerification', () => {
  const ke = new KenyaVerification(mockClient);

  test('verifyNationalID rejette format invalide', async () => {
    await expect(ke.verifyNationalID('+254700001111', '123', 'Test')).rejects.toThrow(
      'ID Kenya invalide',
    );
  });
});

describe('SenegalVerification', () => {
  const sn = new SenegalVerification(mockClient);

  test('verifyCNI appelle /initiate avec country_code SN', async () => {
    await sn.verifyCNI('+221700001111', 'SN123456', 'M. Diop');
    expect(mockRequest).toHaveBeenCalledWith(
      'POST',
      '/v1/kyc/initiate',
      expect.objectContaining({ country_code: 'SN' }),
    );
  });
});

describe('IvoireVerification', () => {
  const ci = new IvoireVerification(mockClient);

  test('verifyCNI appelle /initiate avec country_code CI', async () => {
    await ci.verifyCNI('+225000001111', 'CI123456', 'K. Koné');
    expect(mockRequest).toHaveBeenCalledWith(
      'POST',
      '/v1/kyc/initiate',
      expect.objectContaining({ country_code: 'CI' }),
    );
  });
});

describe('MarocVerification', () => {
  const ma = new MarocVerification(mockClient);

  test('verifyCIN rejette format invalide', async () => {
    await expect(ma.verifyCIN('+212600001111', 'BAD', 'Test')).rejects.toThrow('CIN invalide');
  });

  test('verifyCIN accepte format 1-2 lettres + chiffres', async () => {
    await expect(ma.verifyCIN('+212600001111', 'BE123456', 'M. Alaoui')).resolves.toBeDefined();
  });
});

describe('CountryHelpers', () => {
  const helpers = new CountryHelpers(mockClient);

  test('expose tous les pays', () => {
    expect(helpers.NG).toBeInstanceOf(NigeriaVerification);
    expect(helpers.BF).toBeInstanceOf(BurkinaVerification);
    expect(helpers.GH).toBeInstanceOf(GhanaVerification);
    expect(helpers.KE).toBeInstanceOf(KenyaVerification);
    expect(helpers.SN).toBeInstanceOf(SenegalVerification);
    expect(helpers.CI).toBeInstanceOf(IvoireVerification);
    expect(helpers.MA).toBeInstanceOf(MarocVerification);
  });
});

describe('Datakeys — helpers par pays exposés', () => {
  test('dk.BF, dk.NG, etc. sont accessibles', async () => {
    const { Datakeys } = await import('../src/index');
    const dk = new Datakeys(SANDBOX_KEY);
    expect(dk.NG).toBeDefined();
    expect(dk.BF).toBeDefined();
    expect(dk.GH).toBeDefined();
    expect(dk.KE).toBeDefined();
    expect(dk.SN).toBeDefined();
    expect(dk.CI).toBeDefined();
    expect(dk.MA).toBeDefined();
  });
});
