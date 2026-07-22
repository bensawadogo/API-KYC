import React from 'react';

interface ConsentStepProps {
  onConsent: (accepted: boolean) => void;
  language: 'fr' | 'en';
  theme: { primaryColor: string; borderRadius: number };
}

const TEXTS: Record<string, { title: string; body: string; accept: string; decline: string }> = {
  fr: {
    title: 'Consentement requis',
    body: "En continuant, vous acceptez que vos données d'identité soient vérifiées conformément à la réglementation BCEAO. Vos données sont chiffrées et ne sont pas stockées après vérification.",
    accept: "J'accepte et continue",
    decline: 'Refuser',
  },
  en: {
    title: 'Consent required',
    body: 'By continuing, you agree to have your identity data verified in accordance with BCEAO regulations. Your data is encrypted and not stored after verification.',
    accept: 'I accept and continue',
    decline: 'Decline',
  },
};

export function ConsentStep({ onConsent, language, theme }: ConsentStepProps) {
  const t = TEXTS[language] ?? TEXTS.fr;

  return (
    <div style={{ padding: '24px' }}>
      <h2 style={{ color: theme.primaryColor, fontSize: '20px', marginBottom: '16px' }}>
        {t.title}
      </h2>
      <p
        style={{
          color: '#555',
          lineHeight: '1.6',
          marginBottom: '24px',
          fontSize: '14px',
        }}
      >
        {t.body}
      </p>
      <div style={{ display: 'flex', gap: '12px' }}>
        <button
          onClick={() => onConsent(true)}
          style={{
            flex: 1,
            padding: '12px',
            background: theme.primaryColor,
            color: '#fff',
            border: 'none',
            borderRadius: `${theme.borderRadius}px`,
            cursor: 'pointer',
            fontSize: '14px',
            fontWeight: 600,
          }}
        >
          {t.accept}
        </button>
        <button
          onClick={() => onConsent(false)}
          style={{
            padding: '12px 20px',
            background: 'transparent',
            color: '#888',
            border: '1px solid #ddd',
            borderRadius: `${theme.borderRadius}px`,
            cursor: 'pointer',
            fontSize: '14px',
          }}
        >
          {t.decline}
        </button>
      </div>
    </div>
  );
}
