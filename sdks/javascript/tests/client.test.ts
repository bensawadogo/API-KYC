import { Datakeys, KYCError } from '../src';
import { SANDBOX_KEY } from './fixtures';

describe('Datakeys client', () => {
  test('lève KYCError si api key vide', () => {
    expect(() => new Datakeys('')).toThrow(KYCError);
    expect(() => {
      try { new Datakeys('') } catch (e) {
        expect((e as KYCError).code).toBe('KYC_AUTH_001');
        throw e;
      }
    }).toThrow();
  });

  test('livemode false avec clé dk_test_', () => {
    const dk = new Datakeys(SANDBOX_KEY);
    expect(dk.livemode).toBe(false);
  });

  test('livemode true avec clé dk_live_', () => {
    const dk = new Datakeys('dk_live_fakekey');
    expect(dk.livemode).toBe(true);
  });

  test('baseURL sandbox par défaut', () => {
    const dk = new Datakeys(SANDBOX_KEY);
    expect((dk as any).kyc['client'].baseURL).toContain('localhost:8081');
  });

  test('baseURL custom override', () => {
    const dk = new Datakeys(SANDBOX_KEY, { baseURL: 'https://custom.api.test' });
    expect((dk as any).kyc['client'].baseURL).toBe('https://custom.api.test');
  });
});
