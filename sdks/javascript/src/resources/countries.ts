import { DatakeysClient } from '../client';
import { Country, CountryDocType } from '../types';

export class CountriesResource {
  constructor(private readonly client: DatakeysClient) {}

  async list(): Promise<Country[]> {
    return this.client.request<Country[]>('GET', '/v1/kyc/countries');
  }

  async docTypes(countryCode: string): Promise<CountryDocType[]> {
    return this.client.request<CountryDocType[]>(
      'GET',
      `/v1/kyc/countries/${countryCode.toUpperCase()}/doctypes`,
    );
  }
}
