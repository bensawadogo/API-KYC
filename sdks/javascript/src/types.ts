export type Environment = 'sandbox' | 'production';

export interface DatakeysConfig {
  apiKey: string;
  environment?: Environment;
  timeout?: number;
  maxRetries?: number;
  baseURL?: string;
}

export interface KYCInitiateParams {
  phone: string;
  country_code: string;
  doc_type: DocType;
  doc_number?: string;
  full_name: string;
  consent: true;
  callback_url?: string;
  language?: 'fr' | 'en' | 'ar' | 'sw' | 'ha';
  idempotency_key?: string;
}

export type DocType =
  | 'NATIONAL_ID'
  | 'PASSPORT'
  | 'DRIVERS_LICENSE'
  | 'VOTER_CARD'
  | 'RESIDENCE_PERMIT';

export type VerificationStatus =
  | 'pending'
  | 'processing'
  | 'approved'
  | 'rejected'
  | 'manual_review'
  | 'expired';

export interface KYCVerification {
  id: string;
  object: 'verification';
  livemode: boolean;
  created: number;
  status: VerificationStatus;
  phone_hash: string;
  country_code: string;
  doc_type: DocType;
  score: number;
  provider: string;
  flags: VerificationFlag[];
  aml_score: number;
  is_sanctioned: boolean;
  is_pep: boolean;
  upload_url?: string;
  expires_in?: number;
  processed_at?: string;
}

export type VerificationFlag =
  | 'EXPIRED_DOC'
  | 'SANCTIONS_MATCH'
  | 'PEP_DETECTED'
  | 'INVALID_FORMAT'
  | 'LOW_CONFIDENCE'
  | 'DUPLICATE_PHONE'
  | 'MANUAL_REVIEW_REQUIRED'
  | 'COUNTRY_PHONE_MISMATCH'
  | 'PROVIDER_UNAVAILABLE'
  | 'AML_CHECK_FAILED';

export interface Country {
  code: string;
  name: string;
  phone_prefix: string;
  region: Region;
  doc_types: CountryDocType[];
  provider: string;
}

export type Region =
  | 'WEST_AFRICA'
  | 'NORTH_AFRICA'
  | 'EAST_AFRICA'
  | 'CENTRAL_AFRICA'
  | 'SOUTHERN_AFRICA';

export interface CountryDocType {
  code: DocType;
  name: string;
  pattern?: string;
}

export interface APIResponse<T> {
  success: boolean;
  data: T | null;
  error: string | null;
  timestamp: string;
}
