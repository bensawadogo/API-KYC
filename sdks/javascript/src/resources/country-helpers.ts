import { DatakeysClient } from '../client';
import { KYCVerification } from '../types';

export class NigeriaVerification {
  constructor(private client: DatakeysClient) {}

  async verifyNIN(
    phone: string,
    nin: string,
    fullName: string,
  ): Promise<KYCVerification> {
    this._validateNIN(nin);
    return this.client.request<KYCVerification>('POST', '/v1/kyc/initiate', {
      phone,
      country_code: 'NG',
      doc_type: 'NATIONAL_ID',
      doc_number: nin,
      full_name: fullName,
      consent: true,
      language: 'en',
    });
  }

  async verifyBVN(
    phone: string,
    bvn: string,
    fullName: string,
  ): Promise<KYCVerification> {
    if (!/^[0-9]{11}$/.test(bvn)) {
      throw new Error('BVN invalide: doit contenir 11 chiffres');
    }
    return this.client.request<KYCVerification>('POST', '/v1/kyc/initiate', {
      phone,
      country_code: 'NG',
      doc_type: 'NATIONAL_ID',
      doc_number: bvn,
      full_name: fullName,
      consent: true,
      language: 'en',
    });
  }

  async verifyVoterCard(
    phone: string,
    voterNumber: string,
    fullName: string,
  ): Promise<KYCVerification> {
    return this.client.request<KYCVerification>('POST', '/v1/kyc/initiate', {
      phone,
      country_code: 'NG',
      doc_type: 'VOTER_CARD',
      doc_number: voterNumber,
      full_name: fullName,
      consent: true,
      language: 'en',
    });
  }

  private _validateNIN(nin: string) {
    if (!/^[0-9]{11}$/.test(nin)) {
      throw new Error('NIN invalide: doit contenir 11 chiffres');
    }
  }
}

export class BurkinaVerification {
  constructor(private client: DatakeysClient) {}

  async verifyCNIB(
    phone: string,
    cnib: string,
    fullName: string,
  ): Promise<KYCVerification> {
    if (!/^[A-Z][0-9]{7}$/.test(cnib)) {
      throw new Error('CNIB invalide: format B1234567 (1 lettre + 7 chiffres)');
    }
    return this.client.request<KYCVerification>('POST', '/v1/kyc/initiate', {
      phone,
      country_code: 'BF',
      doc_type: 'NATIONAL_ID',
      doc_number: cnib,
      full_name: fullName,
      consent: true,
      language: 'fr',
    });
  }

  async verifyPassport(
    phone: string,
    passport: string,
    fullName: string,
  ): Promise<KYCVerification> {
    return this.client.request<KYCVerification>('POST', '/v1/kyc/initiate', {
      phone,
      country_code: 'BF',
      doc_type: 'PASSPORT',
      doc_number: passport,
      full_name: fullName,
      consent: true,
      language: 'fr',
    });
  }
}

export class GhanaVerification {
  constructor(private client: DatakeysClient) {}

  async verifyGhanaCard(
    phone: string,
    ghanaCard: string,
    fullName: string,
  ): Promise<KYCVerification> {
    if (!/^GHA-[0-9]{9}-[0-9]$/.test(ghanaCard)) {
      throw new Error('Ghana Card invalide: format GHA-000000000-0');
    }
    return this.client.request<KYCVerification>('POST', '/v1/kyc/initiate', {
      phone,
      country_code: 'GH',
      doc_type: 'NATIONAL_ID',
      doc_number: ghanaCard,
      full_name: fullName,
      consent: true,
      language: 'en',
    });
  }
}

export class KenyaVerification {
  constructor(private client: DatakeysClient) {}

  async verifyNationalID(
    phone: string,
    idNumber: string,
    fullName: string,
  ): Promise<KYCVerification> {
    if (!/^[0-9]{8}$/.test(idNumber)) {
      throw new Error('ID Kenya invalide: 8 chiffres requis');
    }
    return this.client.request<KYCVerification>('POST', '/v1/kyc/initiate', {
      phone,
      country_code: 'KE',
      doc_type: 'NATIONAL_ID',
      doc_number: idNumber,
      full_name: fullName,
      consent: true,
      language: 'en',
    });
  }
}

export class SenegalVerification {
  constructor(private client: DatakeysClient) {}

  async verifyCNI(
    phone: string,
    cni: string,
    fullName: string,
  ): Promise<KYCVerification> {
    return this.client.request<KYCVerification>('POST', '/v1/kyc/initiate', {
      phone,
      country_code: 'SN',
      doc_type: 'NATIONAL_ID',
      doc_number: cni,
      full_name: fullName,
      consent: true,
      language: 'fr',
    });
  }
}

export class IvoireVerification {
  constructor(private client: DatakeysClient) {}

  async verifyCNI(
    phone: string,
    cni: string,
    fullName: string,
  ): Promise<KYCVerification> {
    return this.client.request<KYCVerification>('POST', '/v1/kyc/initiate', {
      phone,
      country_code: 'CI',
      doc_type: 'NATIONAL_ID',
      doc_number: cni,
      full_name: fullName,
      consent: true,
      language: 'fr',
    });
  }
}

export class MarocVerification {
  constructor(private client: DatakeysClient) {}

  async verifyCIN(
    phone: string,
    cin: string,
    fullName: string,
  ): Promise<KYCVerification> {
    if (!/^[A-Z]{1,2}[0-9]{5,6}$/.test(cin)) {
      throw new Error('CIN invalide: format BE123456');
    }
    return this.client.request<KYCVerification>('POST', '/v1/kyc/initiate', {
      phone,
      country_code: 'MA',
      doc_type: 'NATIONAL_ID',
      doc_number: cin,
      full_name: fullName,
      consent: true,
      language: 'ar',
    });
  }
}

export class CountryHelpers {
  readonly BF: BurkinaVerification;
  readonly NG: NigeriaVerification;
  readonly GH: GhanaVerification;
  readonly KE: KenyaVerification;
  readonly SN: SenegalVerification;
  readonly CI: IvoireVerification;
  readonly MA: MarocVerification;

  constructor(client: DatakeysClient) {
    this.BF = new BurkinaVerification(client);
    this.NG = new NigeriaVerification(client);
    this.GH = new GhanaVerification(client);
    this.KE = new KenyaVerification(client);
    this.SN = new SenegalVerification(client);
    this.CI = new IvoireVerification(client);
    this.MA = new MarocVerification(client);
  }
}
