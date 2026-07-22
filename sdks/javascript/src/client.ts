import { DatakeysConfig, APIResponse } from './types';
import { KYCError, KYCErrorCode } from './errors';

const SANDBOX_URL = 'http://localhost:8081';
const PRODUCTION_URL = 'https://api.datakeys.africa';
const SDK_VERSION = '1.0.0';

function detectEnvironment(apiKey: string): 'sandbox' | 'production' {
  return apiKey.startsWith('dk_live_') ? 'production' : 'sandbox';
}

function extractCode(error: string | null): KYCErrorCode {
  const match = (error ?? '').match(/KYC_[A-Z0-9_]+/);
  return (match?.[0] as KYCErrorCode) ?? 'KYC_UNKNOWN';
}

export class DatakeysClient {
  readonly apiKey: string;
  readonly baseURL: string;
  readonly livemode: boolean;
  private timeout: number;
  private maxRetries: number;

  constructor(config: DatakeysConfig) {
    if (!config.apiKey?.trim()) {
      throw new KYCError(
        'API key manquante. Obtenir une clé sur https://dashboard.datakeys.africa',
        'KYC_AUTH_001', 401,
      );
    }

    this.apiKey = config.apiKey;
    this.livemode = config.apiKey.startsWith('dk_live_');
    this.timeout = config.timeout ?? 30_000;
    this.maxRetries = config.maxRetries ?? 3;

    const env = config.environment ?? detectEnvironment(config.apiKey);
    this.baseURL = config.baseURL ?? (env === 'production' ? PRODUCTION_URL : SANDBOX_URL);
  }

  async request<T>(
    method: 'GET' | 'POST',
    path: string,
    body?: unknown,
    idempotencyKey?: string,
  ): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'X-API-Key': this.apiKey,
      'X-SDK-Version': SDK_VERSION,
      'X-SDK-Lang': 'javascript',
      'X-SDK-Livemode': String(this.livemode),
    };

    if (idempotencyKey) {
      headers['Idempotency-Key'] = idempotencyKey;
    }

    let lastError: Error | null = null;

    for (let attempt = 0; attempt < this.maxRetries; attempt++) {
      const controller = new AbortController();
      const timer = setTimeout(() => controller.abort(), this.timeout);

      try {
        const response = await fetch(`${this.baseURL}${path}`, {
          method,
          headers,
          body: body ? JSON.stringify(body) : undefined,
          signal: controller.signal,
        });
        clearTimeout(timer);

        const json = (await response.json()) as APIResponse<T>;

        if (!response.ok) {
          const code = extractCode(json.error);
          if (response.status < 500) {
            throw new KYCError(json.error ?? 'Erreur API', code, response.status, json);
          }
          lastError = new KYCError(json.error ?? 'Erreur serveur', 'KYC_SERVER_ERR', response.status);
          await sleep(exponentialBackoff(attempt));
          continue;
        }

        if (!json.success || json.data === null) {
          throw new KYCError(
            json.error ?? 'Réponse inattendue',
            extractCode(json.error),
            response.status,
            json,
          );
        }

        return json.data as T;
      } catch (err) {
        clearTimeout(timer);
        if (err instanceof KYCError) throw err;

        if ((err as Error).name === 'AbortError') {
          lastError = new KYCError(`Timeout après ${this.timeout}ms`, 'KYC_TIMEOUT', 0);
        } else {
          lastError = new KYCError(`Erreur réseau: ${(err as Error).message}`, 'KYC_NETWORK', 0);
        }

        if (attempt < this.maxRetries - 1) {
          await sleep(exponentialBackoff(attempt));
        }
      }
    }

    throw lastError ?? new KYCError('Échec après tous les retries', 'KYC_NETWORK', 0);
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

function exponentialBackoff(attempt: number): number {
  const base = 500;
  const delay = base * Math.pow(2, attempt);
  const jitter = Math.random() * delay * 0.3;
  return Math.min(delay + jitter, 10_000);
}
