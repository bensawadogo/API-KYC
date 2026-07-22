/// <reference types="@testing-library/jest-dom" />
/// <reference types="jest" />
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { DatakeysWidget } from '../src/DatakeysWidget';

const mockInitiate = jest.fn().mockResolvedValue({
  id: 'ver_test123',
  status: 'pending',
  upload_url: 'http://test/upload',
  score: 0,
  provider: 'sandbox',
  flags: [],
  is_sanctioned: false,
  is_pep: false,
});

const mockRetrieve = jest.fn().mockResolvedValue({
  id: 'ver_test123',
  status: 'approved',
  score: 0.95,
  provider: 'sandbox',
  flags: [],
  is_sanctioned: false,
  is_pep: false,
});

jest.mock('@datakeys/kyc', () => ({
  __esModule: true,
  default: jest.fn().mockImplementation(() => ({
    kyc: { initiate: mockInitiate, retrieve: mockRetrieve },
    livemode: false,
  })),
}));

const defaultProps = {
  apiKey: 'dk_test_sandbox',
  countryCode: 'BF',
  onSuccess: jest.fn(),
  onError: jest.fn(),
  onClose: jest.fn(),
};

describe('DatakeysWidget', () => {
  beforeEach(() => jest.clearAllMocks());

  test('affiche l\'étape consentement au démarrage', () => {
    render(<DatakeysWidget {...defaultProps} />);
    expect(screen.getByText(/Consentement requis/i)).toBeInTheDocument();
    expect(screen.getByText(/J'accepte et continue/i)).toBeInTheDocument();
  });

  test('header affiche DATAKEYS KYC', () => {
    render(<DatakeysWidget {...defaultProps} />);
    expect(screen.getByText('DATAKEYS KYC')).toBeInTheDocument();
  });

  test('clic refuser appelle onClose', () => {
    render(<DatakeysWidget {...defaultProps} />);
    fireEvent.click(screen.getByText(/Refuser/i));
    expect(defaultProps.onClose).toHaveBeenCalled();
  });

  test('avance à l\'étape phone après consentement', () => {
    render(<DatakeysWidget {...defaultProps} />);
    fireEvent.click(screen.getByText(/J'accepte et continue/i));
    expect(screen.getByText(/Vos informations/i)).toBeInTheDocument();
  });

  test('langue EN affiche texte anglais', () => {
    render(<DatakeysWidget {...defaultProps} language="en" />);
    expect(screen.getByText(/Consent required/i)).toBeInTheDocument();
    expect(screen.getByText(/I accept and continue/i)).toBeInTheDocument();
  });

  test('theme custom applique primaryColor', () => {
    render(<DatakeysWidget {...defaultProps} theme={{ primaryColor: '#ff0000' }} />);
    const header = screen.getByText('DATAKEYS KYC').parentElement;
    expect(header?.style.background).toBe('rgb(255, 0, 0)');
  });
});
