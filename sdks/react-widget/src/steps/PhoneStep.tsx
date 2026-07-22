import React, { useState } from 'react';

interface PhoneStepProps {
  countryCode: string;
  prefill?: { phone?: string; fullName?: string };
  onSubmit: (data: { phone: string; fullName: string }) => void;
  theme: { primaryColor: string; borderRadius: number };
  language: 'fr' | 'en';
}

const PHONE_PREFIXES: Record<string, string> = {
  BF: '+226', NG: '+234', KE: '+254', GH: '+233',
  SN: '+221', CI: '+225', MA: '+212', TZ: '+255',
  ZA: '+27',  CM: '+237', UG: '+256', ET: '+251',
};

export function PhoneStep({ countryCode, prefill, onSubmit, theme, language }: PhoneStepProps) {
  const prefix = PHONE_PREFIXES[countryCode] ?? '+';
  const [phone, setPhone] = useState(prefill?.phone ?? '');
  const [fullName, setFullName] = useState(prefill?.fullName ?? '');
  const [errors, setErrors] = useState<Record<string, string>>({});

  const validate = (): boolean => {
    const e: Record<string, string> = {};
    const fullPhone = phone.startsWith('+') ? phone : prefix + phone;
    if (!/^\+[1-9]\d{7,14}$/.test(fullPhone)) {
      e.phone = language === 'fr' ? 'Numéro invalide (format: +22670000000)' : 'Invalid number (format: +22670000000)';
    }
    if (fullName.trim().length < 2) {
      e.fullName = language === 'fr' ? 'Nom complet requis (min 2 caractères)' : 'Full name required (min 2 chars)';
    }
    setErrors(e);
    return Object.keys(e).length === 0;
  };

  const handleSubmit = () => {
    if (!validate()) return;
    const fullPhone = phone.startsWith('+') ? phone : prefix + phone;
    onSubmit({ phone: fullPhone, fullName: fullName.trim() });
  };

  const inputStyle = (hasError: boolean) => ({
    width: '100%',
    padding: '10px 12px',
    border: `1px solid ${hasError ? '#e53e3e' : '#ddd'}`,
    borderRadius: `${theme.borderRadius}px`,
    fontSize: '14px',
    outline: 'none',
    boxSizing: 'border-box' as const,
    marginTop: '6px',
  });

  return (
    <div style={{ padding: '24px' }}>
      <h2 style={{ color: theme.primaryColor, marginBottom: '20px' }}>
        {language === 'fr' ? 'Vos informations' : 'Your information'}
      </h2>

      <div style={{ marginBottom: '16px' }}>
        <label style={{ fontSize: '13px', color: '#555' }}>
          {language === 'fr' ? 'Numéro de téléphone' : 'Phone number'}
        </label>
        <input
          type="tel"
          value={phone}
          onChange={(e) => setPhone(e.target.value)}
          placeholder={`${prefix}70000000`}
          style={inputStyle(!!errors.phone)}
        />
        {errors.phone && (
          <p style={{ color: '#e53e3e', fontSize: '12px', marginTop: '4px' }}>{errors.phone}</p>
        )}
      </div>

      <div style={{ marginBottom: '24px' }}>
        <label style={{ fontSize: '13px', color: '#555' }}>
          {language === 'fr' ? 'Nom complet' : 'Full name'}
        </label>
        <input
          type="text"
          value={fullName}
          onChange={(e) => setFullName(e.target.value)}
          placeholder={language === 'fr' ? 'Aminata Ouédraogo' : 'John Doe'}
          style={inputStyle(!!errors.fullName)}
        />
        {errors.fullName && (
          <p style={{ color: '#e53e3e', fontSize: '12px', marginTop: '4px' }}>{errors.fullName}</p>
        )}
      </div>

      <button
        onClick={handleSubmit}
        style={{
          width: '100%',
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
        {language === 'fr' ? 'Continuer' : 'Continue'}
      </button>
    </div>
  );
}
