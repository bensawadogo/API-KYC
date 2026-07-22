import { useState, useCallback } from 'react';
import Datakeys from '@datakeys/kyc';
import type { KYCVerification, KYCInitiateParams } from '@datakeys/kyc';

interface UseKYCOptions {
  apiKey: string;
  baseURL?: string;
}

interface UseKYCState {
  verification: KYCVerification | null;
  isLoading: boolean;
  error: string | null;
}

export function useKYC({ apiKey, baseURL }: UseKYCOptions) {
  const [state, setState] = useState<UseKYCState>({
    verification: null,
    isLoading: false,
    error: null,
  });

  const client = new Datakeys(apiKey, { baseURL });

  const initiate = useCallback(
    async (params: KYCInitiateParams): Promise<KYCVerification | null> => {
      setState((s) => ({ ...s, isLoading: true, error: null }));
      try {
        const v = await client.kyc.initiate(params);
        setState((s) => ({ ...s, verification: v, isLoading: false }));
        return v;
      } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : 'Erreur';
        setState((s) => ({ ...s, isLoading: false, error: msg }));
        return null;
      }
    },
    [apiKey],
  );

  const retrieve = useCallback(async (id: string): Promise<KYCVerification | null> => {
    try {
      const v = await client.kyc.retrieve(id);
      setState((s) => ({ ...s, verification: v }));
      return v;
    } catch {
      return null;
    }
  }, [apiKey]);

  return { ...state, initiate, retrieve };
}
