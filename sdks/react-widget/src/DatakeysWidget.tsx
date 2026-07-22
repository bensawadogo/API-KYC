import React, { useState, useCallback } from 'react';
import type { DatakeysWidgetProps, WidgetStep, WidgetVerification } from './types';
import { ConsentStep } from './steps/ConsentStep';
import { PhoneStep } from './steps/PhoneStep';
import { DocumentStep } from './steps/DocumentStep';
import { ResultStep } from './steps/ResultStep';
import { useKYC } from './hooks/useKYC';
import { usePolling } from './hooks/usePolling';

const STEPS: WidgetStep[] = ['consent', 'phone', 'document', 'processing', 'result'];
const TERMINAL = ['approved', 'rejected', 'manual_review', 'expired'];

const DEFAULT_THEME = {
  primaryColor: '#1a1a2e',
  backgroundColor: '#ffffff',
  borderRadius: 8,
  fontFamily: 'system-ui, sans-serif',
};

export function DatakeysWidget({
  apiKey,
  countryCode,
  onSuccess,
  onError,
  onClose,
  onStep,
  language = 'fr',
  theme: userTheme,
  prefill,
  baseURL,
}: DatakeysWidgetProps) {
  const theme = { ...DEFAULT_THEME, ...userTheme };
  const lang = language === 'en' ? 'en' : 'fr';

  const [step, setStep] = useState<WidgetStep>('consent');
  const [formData, setFormData] = useState({ phone: '', fullName: '', docType: '', docNumber: '' });
  const [verificationId, setVerId] = useState<string | null>(null);
  const [finalResult, setFinal] = useState<WidgetVerification | null>(null);

  const { initiate, retrieve, isLoading, error } = useKYC({ apiKey, baseURL });

  const goToStep = useCallback(
    (s: WidgetStep, data?: unknown) => {
      setStep(s);
      onStep?.(s, data);
    },
    [onStep],
  );

  const pollFn = useCallback(async () => {
    if (!verificationId) return false;
    const v = await retrieve(verificationId);
    if (!v) return false;

    if (!TERMINAL.includes(v.status)) return false;

    const result: WidgetVerification = {
      verificationId: v.id,
      status: v.status as WidgetVerification['status'],
      score: v.score,
      provider: v.provider,
      flags: v.flags,
      isSanctioned: v.is_sanctioned,
      isPEP: v.is_pep,
    };

    setFinal(result);
    goToStep('result', result);

    if (v.status === 'approved' || v.status === 'manual_review') {
      onSuccess(result);
    } else {
      onError({ code: v.flags[0] ?? 'KYC_REJECTED', message: `Statut: ${v.status}`, step: 'result' });
    }
    return true;
  }, [verificationId, retrieve, goToStep, onSuccess, onError]);

  usePolling({ fn: pollFn, interval: 3000, enabled: step === 'processing' && !!verificationId });

  const handleConsent = (accepted: boolean) => {
    if (!accepted) { onClose?.(); return; }
    goToStep('phone');
  };

  const handlePhone = ({ phone, fullName }: { phone: string; fullName: string }) => {
    setFormData((d) => ({ ...d, phone, fullName }));
    goToStep('document');
  };

  const handleDocument = async ({ docType, docNumber }: { docType: string; docNumber: string }) => {
    const merged = { ...formData, docType, docNumber };
    setFormData(merged);
    goToStep('processing');

    const v = await initiate({
      phone: merged.phone,
      country_code: countryCode,
      doc_type: docType as any,
      doc_number: docNumber,
      full_name: merged.fullName,
      consent: true,
      language: lang,
    });

    if (!v) {
      goToStep('document');
      onError({ code: 'KYC_INIT_FAILED', message: error ?? "Erreur lors de l'initiation", step: 'document' });
      return;
    }

    setVerId(v.id);
  };

  const currentIdx = STEPS.indexOf(step);

  const containerStyle: React.CSSProperties = {
    fontFamily: theme.fontFamily,
    backgroundColor: theme.backgroundColor,
    borderRadius: `${theme.borderRadius + 4}px`,
    boxShadow: '0 4px 24px rgba(0,0,0,0.12)',
    width: '100%',
    maxWidth: '480px',
    overflow: 'hidden',
  };

  return (
    <div style={containerStyle}>
      <div
        style={{
          background: theme.primaryColor,
          padding: '16px 24px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <span style={{ color: '#fff', fontWeight: 700, fontSize: '16px', letterSpacing: '0.5px' }}>
          DATAKEYS KYC
        </span>
        {onClose && (
          <button
            onClick={onClose}
            style={{
              background: 'transparent',
              border: 'none',
              color: '#ffffff88',
              cursor: 'pointer',
              fontSize: '18px',
              lineHeight: 1,
            }}
          >
            ×
          </button>
        )}
      </div>

      <div style={{ height: '3px', background: '#f0f0f0' }}>
        <div
          style={{
            height: '100%',
            background: theme.primaryColor,
            width: `${((currentIdx + 1) / STEPS.length) * 100}%`,
            transition: 'width 0.3s ease',
          }}
        />
      </div>

      {step === 'consent' && <ConsentStep onConsent={handleConsent} language={lang} theme={theme} />}
      {step === 'phone' && (
        <PhoneStep countryCode={countryCode} prefill={prefill} onSubmit={handlePhone} theme={theme} language={lang} />
      )}
      {step === 'document' && <DocumentStep countryCode={countryCode} onSubmit={handleDocument} theme={theme} language={lang} />}
      {(step === 'processing' || step === 'result') && (
        <ResultStep verification={finalResult} isProcessing={step === 'processing'} language={lang} theme={theme} onClose={onClose} />
      )}
    </div>
  );
}
