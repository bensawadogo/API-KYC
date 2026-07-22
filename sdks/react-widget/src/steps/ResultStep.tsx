import React from 'react';
import type { WidgetVerification } from '../types';

interface ResultStepProps {
  verification: WidgetVerification | null;
  isProcessing: boolean;
  language: 'fr' | 'en';
  theme: { primaryColor: string; borderRadius: number };
  onClose?: () => void;
}

export function ResultStep({ verification, isProcessing, language, theme, onClose }: ResultStepProps) {
  if (isProcessing) {
    return (
      <div style={{ padding: '48px 24px', textAlign: 'center' }}>
        <div
          style={{
            width: '48px',
            height: '48px',
            border: `4px solid ${theme.primaryColor}22`,
            borderTop: `4px solid ${theme.primaryColor}`,
            borderRadius: '50%',
            animation: 'dk-spin 1s linear infinite',
            margin: '0 auto 24px',
          }}
        />
        <style>{`@keyframes dk-spin{to{transform:rotate(360deg)}}`}</style>
        <p style={{ color: '#555', fontSize: '14px' }}>
          {language === 'fr' ? 'Vérification en cours...' : 'Verification in progress...'}
        </p>
      </div>
    );
  }

  if (!verification) return null;

  const isApproved = verification.status === 'approved';
  const isManual = verification.status === 'manual_review';

  const icon = isApproved ? '✅' : isManual ? '⏳' : '❌';
  const color = isApproved ? '#38a169' : isManual ? '#d69e2e' : '#e53e3e';
  const titleFr = isApproved ? 'Identité vérifiée' : isManual ? 'En cours de révision' : 'Vérification échouée';
  const titleEn = isApproved ? 'Identity verified' : isManual ? 'Under review' : 'Verification failed';
  const bodyFr = isApproved
    ? `Votre identité a été vérifiée avec un score de ${Math.round(verification.score * 100)}%.`
    : isManual
    ? "Votre dossier est en cours d'examen par notre équipe."
    : `La vérification n'a pas abouti. Flags: ${verification.flags.join(', ') || 'aucun'}`;

  return (
    <div style={{ padding: '24px', textAlign: 'center' }}>
      <div style={{ fontSize: '48px', marginBottom: '16px' }}>{icon}</div>
      <h2 style={{ color, marginBottom: '12px' }}>
        {language === 'fr' ? titleFr : titleEn}
      </h2>
      <p style={{ color: '#666', fontSize: '14px', marginBottom: '8px' }}>
        {language === 'fr' ? bodyFr : bodyFr}
      </p>
      <p style={{ color: '#999', fontSize: '12px', marginBottom: '24px' }}>
        ID: {verification.verificationId}
      </p>
      {onClose && (
        <button
          onClick={onClose}
          style={{
            padding: '10px 32px',
            background: theme.primaryColor,
            color: '#fff',
            border: 'none',
            borderRadius: `${theme.borderRadius}px`,
            cursor: 'pointer',
            fontSize: '14px',
          }}
        >
          {language === 'fr' ? 'Fermer' : 'Close'}
        </button>
      )}
    </div>
  );
}
