import { DatakeysClient } from './client';
import { KYCResource } from './resources/kyc';
import { CountriesResource } from './resources/countries';
import { CountryHelpers } from './resources/country-helpers';
import type {
  BurkinaVerification,
  NigeriaVerification,
  GhanaVerification,
  KenyaVerification,
  SenegalVerification,
  IvoireVerification,
  MarocVerification,
} from './resources/country-helpers';
import { DatakeysConfig } from './types';

export * from './types';
export * from './errors';
export * from './resources/country-helpers';

export class Datakeys {
  readonly kyc: KYCResource;
  readonly countries: CountriesResource;
  readonly livemode: boolean;

  readonly NG: NigeriaVerification;
  readonly BF: BurkinaVerification;
  readonly GH: GhanaVerification;
  readonly KE: KenyaVerification;
  readonly SN: SenegalVerification;
  readonly CI: IvoireVerification;
  readonly MA: MarocVerification;

  constructor(apiKey: string, config?: Omit<DatakeysConfig, 'apiKey'>) {
    const client = new DatakeysClient({ apiKey, ...config });
    this.kyc = new KYCResource(client);
    this.countries = new CountriesResource(client);
    this.livemode = client.livemode;

    const helpers = new CountryHelpers(client);
    this.NG = helpers.NG;
    this.BF = helpers.BF;
    this.GH = helpers.GH;
    this.KE = helpers.KE;
    this.SN = helpers.SN;
    this.CI = helpers.CI;
    this.MA = helpers.MA;
  }
}

export default Datakeys;
