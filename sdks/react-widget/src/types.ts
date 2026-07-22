export interface DatakeysWidgetProps {
  apiKey: string;
  countryCode: string;
  onSuccess: (verification: WidgetVerification) => void;
  onError: (error: WidgetError) => void;
  onClose?: () => void;
  onStep?: (step: WidgetStep, data?: unknown) => void;
  language?: 'fr' | 'en' | 'ar' | 'sw' | 'ha';
  theme?: WidgetTheme;
  prefill?: WidgetPrefill;
  baseURL?: string;
}

export type WidgetStep =
  | 'consent'
  | 'phone'
  | 'document'
  | 'processing'
  | 'result';

export interface WidgetTheme {
  primaryColor?: string;
  backgroundColor?: string;
  borderRadius?: number;
  fontFamily?: string;
}

export interface WidgetPrefill {
  phone?: string;
  fullName?: string;
  docType?: string;
  docNumber?: string;
}

export interface WidgetVerification {
  verificationId: string;
  status: 'approved' | 'rejected' | 'manual_review';
  score: number;
  provider: string;
  flags: string[];
  isSanctioned: boolean;
  isPEP: boolean;
}

export interface WidgetError {
  code: string;
  message: string;
  step: WidgetStep;
}
