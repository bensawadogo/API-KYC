const Datakeys = require('@datakeys/kyc').default;

const dk = new Datakeys('dk_test_datakeys_sandbox_001');

async function main() {
  // 1. Voir les pays supportés
  const countries = await dk.countries.list();
  console.log(`${countries.length} pays supportés`);

  // 2. Initier une vérification
  const v = await dk.kyc.initiate({
    phone:        '+22670000001',
    country_code: 'BF',
    doc_type:     'NATIONAL_ID',
    doc_number:   'B1230000',
    full_name:    'Aminata Ouédraogo',
    consent:      true,
  });
  console.log('ID:', v.id);
  console.log('Upload URL:', v.upload_url);

  // 3. Attendre le résultat
  const result = await dk.kyc.waitForCompletion(v.id);
  console.log('Statut final:', result.status);
  console.log('Score:', result.score);
}

main().catch(console.error);
