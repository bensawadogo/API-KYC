import React, { useState } from 'react';

interface DocumentStepProps {
  countryCode: string;
  onSubmit: (data: { docType: string; docNumber: string }) => void;
  theme: { primaryColor: string; borderRadius: number };
  language: 'fr' | 'en';
}

const COUNTRY_DOCS: Record<string, Array<{ code: string; label_fr: string; label_en: string; placeholder: string; pattern?: RegExp }>> = {
  BF: [
    { code: 'NATIONAL_ID', label_fr: 'CNIB', label_en: 'National ID', placeholder: 'B1234567', pattern: /^[A-Z][0-9]{7}$/ },
    { code: 'PASSPORT', label_fr: 'Passeport', label_en: 'Passport', placeholder: 'AB123456' },
  ],
  NG: [
    { code: 'NATIONAL_ID', label_fr: 'NIN', label_en: 'NIN', placeholder: '12345678901', pattern: /^[0-9]{11}$/ },
    { code: 'PASSPORT', label_fr: 'Passeport', label_en: 'Passport', placeholder: 'A12345678' },
    { code: 'VOTER_CARD', label_fr: "Carte d'électeur", label_en: 'Voter Card', placeholder: '9JA00000001234' },
  ],
  GH: [
    { code: 'NATIONAL_ID', label_fr: 'Ghana Card', label_en: 'Ghana Card', placeholder: 'GHA-000000000-0', pattern: /^GHA-[0-9]{9}-[0-9]$/ },
    { code: 'PASSPORT', label_fr: 'Passeport', label_en: 'Passport', placeholder: 'G1234567' },
  ],
  KE: [
    { code: 'NATIONAL_ID', label_fr: 'ID national', label_en: 'National ID', placeholder: '12345678', pattern: /^[0-9]{8}$/ },
    { code: 'PASSPORT', label_fr: 'Passeport', label_en: 'Passport', placeholder: 'AK123456' },
  ],
  DEFAULT: [
    { code: 'NATIONAL_ID', label_fr: 'Carte nationale', label_en: 'National ID', placeholder: '...' },
    { code: 'PASSPORT', label_fr: 'Passeport', label_en: 'Passport', placeholder: '...' },
  ],
};

export function DocumentStep({ countryCode, onSubmit, theme, language }: DocumentStepProps) {
  const docs = COUNTRY_DOCS[countryCode] ?? COUNTRY_DOCS.DEFAULT;
  const [selectedDoc, setSelectedDoc] = useState(docs[0].code);
  const [docNumber, setDocNumber] = useState('');
  const [error, setError] = useState('');

  const currentDoc = docs.find((d) => d.code === selectedDoc)!;

  const validate = () => {
    if (currentDoc.pattern && !currentDoc.pattern.test(docNumber)) {
      setError(language === 'fr' ? `Format invalide. Exemple: ${currentDoc.placeholder}` : `Invalid format. Example: ${currentDoc.placeholder}`);
      return false;
    }
    if (docNumber.trim().length < 4) {
      setError(language === 'fr' ? 'Numéro de document requis' : 'Document number required');
      return false;
    }
    return true;
  };

  const handleSubmit = () => {
    setError('');
    if (!validate()) return;
    onSubmit({ docType: selectedDoc, docNumber: docNumber.trim() });
  };

  return (
    <div style={{ padding: '24px' }}>
      <h2 style={{ color: theme.primaryColor, marginBottom: '20px' }}>
        {language === 'fr' ? 'Type de document' : 'Document type'}
      </h2>

      <div style={{ marginBottom: '16px' }}>
        {docs.map((doc) => (
          <label
            key={doc.code}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '10px',
              padding: '12px',
              border: `2px solid ${selectedDoc === doc.code ? theme.primaryColor : '#eee'}`,
              borderRadius: `${theme.borderRadius}px`,
              marginBottom: '8px',
              cursor: 'pointer',
              transition: 'border-color 0.2s',
            }}
          >
            <input
              type="radio"
              name="docType"
              value={doc.code}
              checked={selectedDoc === doc.code}
              onChange={() => {
                setSelectedDoc(doc.code);
                setDocNumber('');
                setError('');
              }}
              style={{ accentColor: theme.primaryColor }}
            />
            <span style={{ fontSize: '14px', fontWeight: 500 }}>
              {language === 'fr' ? doc.label_fr : doc.label_en}
            </span>
          </label>
        ))}
      </div>

      <div style={{ marginBottom: '24px' }}>
        <label style={{ fontSize: '13px', color: '#555' }}>
          {language === 'fr' ? 'Numéro du document' : 'Document number'}
        </label>
        <input
          type="text"
          value={docNumber}
          onChange={(e) => {
            setDocNumber(e.target.value.toUpperCase());
            setError('');
          }}
          placeholder={currentDoc.placeholder}
          style={{
            width: '100%',
            padding: '10px 12px',
            border: `1px solid ${error ? '#e53e3e' : '#ddd'}`,
            borderRadius: `${theme.borderRadius}px`,
            fontSize: '14px',
            marginTop: '6px',
            boxSizing: 'border-box',
          }}
        />
        {error && (
          <p style={{ color: '#e53e3e', fontSize: '12px', marginTop: '4px' }}>{error}</p>
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
