// Quickstart DATAKEYS KYC — 5 minutes pour intégrer
const Datakeys = require('@datakeys/kyc').default;

const dk = new Datakeys('dk_test_datakeys_sandbox_001');
console.log('Mode:', dk.livemode ? 'PRODUCTION' : 'SANDBOX');

async function main() {
  const countries = await dk.countries.list();
  console.log(`\n${countries.length} pays africains supportés`);
  const bf = countries.find((c) => c.code === 'BF');
  console.log(`Burkina Faso: ${bf?.doc_types.length} types de docs`);

  console.log('\n--- Initiation vérification ---');
  const verification = await dk.kyc.initiate({
    phone: '+22670000001',
    country_code: 'BF',
    doc_type: 'NATIONAL_ID',
    doc_number: 'B1230000',
    full_name: 'Aminata Ouédraogo',
    consent: true,
  });

  console.log('ID:        ', verification.id);
  console.log('Statut:    ', verification.status);
  console.log('Provider:  ', verification.provider);
  if (verification.upload_url) {
    console.log('Upload URL:', verification.upload_url);
  }

  console.log('\n--- Attente résultat (sandbox = rapide) ---');
  const result = await dk.kyc.waitForCompletion(verification.id, {
    maxWaitMs: 60_000,
    intervalMs: 2_000,
    onPoll: (v) => console.log('  Status:', v.status),
  });

  console.log('\n=== RÉSULTAT FINAL ===');
  console.log('Statut:    ', result.status);
  console.log('Score:     ', result.score);
  console.log('Provider:  ', result.provider);
  console.log('Sanctionné:', result.is_sanctioned);
  console.log('PEP:       ', result.is_pep);
  if (result.flags.length > 0) {
    console.log('Flags:     ', result.flags.join(', '));
  }
}

main().catch((err) => {
  if (err.code) {
    console.error(`KYC Error [${err.code}]:`, err.message);
  } else {
    console.error('Erreur:', err.message);
  }
  process.exit(1);
});
