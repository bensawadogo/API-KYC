// Exemple : polling du statut d'une vérification existante
const Datakeys = require('@datakeys/kyc').default;

const dk = new Datakeys('dk_test_datakeys_sandbox_001');
const verificationId = process.argv[2];

if (!verificationId) {
  console.error('Usage: node poll-status.js <verification_id>');
  process.exit(1);
}

async function main() {
  const result = await dk.kyc.waitForCompletion(verificationId, {
    maxWaitMs: 120_000,
    intervalMs: 3_000,
    onPoll: (v) => {
      const elapsed = Math.round((Date.now() - v.created * 1000) / 1000);
      console.log(`[${elapsed}s] Status: ${v.status}  Score: ${v.score}`);
    },
  });

  console.log('\n=== RÉSULTAT FINAL ===');
  console.log(JSON.stringify(result, null, 2));
}

main().catch((err) => {
  console.error('Erreur:', err.message);
  process.exit(1);
});
