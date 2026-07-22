import { randomUUID } from 'crypto';
import { DatakeysClient } from '../client';
import { KYCInitiateParams, KYCVerification, VerificationStatus } from '../types';

const TERMINAL_STATUSES = new Set<VerificationStatus>([
  'approved',
  'rejected',
  'manual_review',
  'expired',
]);

export class KYCResource {
  constructor(private readonly client: DatakeysClient) {}

  async initiate(
    params: KYCInitiateParams,
    options?: { idempotencyKey?: string },
  ): Promise<KYCVerification> {
    const iKey = options?.idempotencyKey ?? params.idempotency_key ?? randomUUID();
    return this.client.request<KYCVerification>('POST', '/v1/kyc/initiate', params, iKey);
  }

  async retrieve(verificationId: string): Promise<KYCVerification> {
    if (!verificationId?.trim()) {
      throw new Error('verificationId est requis');
    }
    return this.client.request<KYCVerification>('GET', `/v1/kyc/status/${verificationId}`);
  }

  async waitForCompletion(
    verificationId: string,
    options?: {
      maxWaitMs?: number;
      intervalMs?: number;
      onPoll?: (v: KYCVerification) => void;
    },
  ): Promise<KYCVerification> {
    const maxWait = options?.maxWaitMs ?? 120_000;
    const interval = options?.intervalMs ?? 3_000;
    const deadline = Date.now() + maxWait;

    while (Date.now() < deadline) {
      const verification = await this.retrieve(verificationId);
      options?.onPoll?.(verification);
      if (TERMINAL_STATUSES.has(verification.status)) {
        return verification;
      }
      await new Promise((r) => setTimeout(r, interval));
    }

    throw new Error(
      `Timeout: vérification ${verificationId} toujours "${(await this.retrieve(verificationId)).status}" après ${maxWait}ms`,
    );
  }
}
