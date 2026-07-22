# @datakeys/kyc-widget

Widget React officiel DATAKEYS KYC.
Intégration en 5 lignes. Sandbox inclus.
Conforme BCEAO. 55 pays africains.

## Installation

```bash
npm install @datakeys/kyc-widget @datakeys/kyc
```

## Usage basique (5 lignes)

```jsx
import { DatakeysModal } from '@datakeys/kyc-widget';

function App() {
  return (
    <DatakeysModal
      isOpen={true}
      apiKey="dk_test_datakeys_sandbox_001"
      countryCode="BF"
      onSuccess={(v) => console.log('Approuvé:', v.verificationId)}
      onError={(e)   => console.error('Erreur:', e.code)}
    />
  );
}
```

## Personnalisation

```jsx
<DatakeysWidget
  apiKey="dk_test_datakeys_sandbox_001"
  countryCode="NG"
  language="en"
  theme={{
    primaryColor: '#006400',
    borderRadius: 12,
  }}
  prefill={{
    phone:    '+2348100000001',
    fullName: 'Chukwuemeka Okafor',
  }}
  onSuccess={(v) => {
    activateUserAccount(v.verificationId);
  }}
  onError={(e) => {
    if (e.code === 'KYC_AML_SANCTION') {
      showSanctionsMessage();
    }
  }}
  onClose={() => setModalOpen(false)}
  onStep={(step) => trackAnalytics('kyc_step', step)}
/>
```

## Props

| Prop | Type | Default | Description |
|---|---|---|---|
| apiKey | string | — | Clé API (dk_test_ ou dk_live_) |
| countryCode | string | — | Code ISO-2 du pays |
| onSuccess | fn | — | Callback vérification approuvée |
| onError | fn | — | Callback erreur |
| onClose | fn | — | Callback fermeture |
| onStep | fn | — | Callback changement d'étape |
| language | string | 'fr' | Langue (fr/en/ar/sw/ha) |
| theme | object | — | Couleurs, bordures |
| prefill | object | — | Préremplir champs |
| baseURL | string | — | Override URL API |
