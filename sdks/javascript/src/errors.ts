export type KYCErrorCode =
  | 'KYC_AUTH_001' | 'KYC_AUTH_002' | 'KYC_AUTH_003'
  | 'KYC_RATE_001'
  | 'KYC_VAL_001' | 'KYC_VAL_002'
  | 'KYC_AML_SANCTION'
  | 'KYC_IDMP_001'
  | 'KYC_SERVER_ERR'
  | 'KYC_NETWORK'
  | 'KYC_TIMEOUT'
  | 'KYC_UNKNOWN';

export class KYCError extends Error {
  readonly code: KYCErrorCode;
  readonly status: number;
  readonly raw: unknown;

  constructor(
    message: string,
    code: KYCErrorCode,
    status: number,
    raw?: unknown,
  ) {
    super(message);
    this.name = 'KYCError';
    this.code = code;
    this.status = status;
    this.raw = raw;
    Object.setPrototypeOf(this, KYCError.prototype);
  }

  isAuthError(): boolean { return this.code.startsWith('KYC_AUTH'); }
  isRateLimit(): boolean { return this.code.startsWith('KYC_RATE'); }
  isValidation(): boolean { return this.code.startsWith('KYC_VAL'); }
  isSanctions(): boolean { return this.code === 'KYC_AML_SANCTION'; }
  isIdempotent(): boolean { return this.code === 'KYC_IDMP_001'; }
  isServerError(): boolean { return this.status >= 500; }
}
