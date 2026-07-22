// Exemple : recevoir les webhooks DATAKEYS KYC
const crypto = require('crypto');

const WEBHOOK_SECRET = process.env.WEBHOOK_HMAC_SECRET;

function verifyWebhookSignature(payload, signature) {
  if (!WEBHOOK_SECRET) {
    console.warn('WEBHOOK_HMAC_SECRET non configuré');
    return false;
  }
  const expected = crypto.createHmac('sha256', WEBHOOK_SECRET).update(payload).digest('hex');
  return crypto.timingSafeEqual(Buffer.from(signature, 'hex'), Buffer.from(expected, 'hex'));
}

function webhookHandler(req, res) {
  const signature = req.headers['x-datakeys-signature'];
  const rawBody = JSON.stringify(req.body);

  if (!verifyWebhookSignature(rawBody, signature)) {
    return res.status(401).json({ error: 'Signature invalide' });
  }

  const { event, data } = req.body;

  switch (event) {
    case 'kyc.approved':
      console.log(`✅ KYC approuvé: ${data.id}  Score: ${data.score}  Provider: ${data.provider}`);
      break;
    case 'kyc.rejected':
      console.log(`❌ KYC rejeté: ${data.id}  Flags: ${data.flags.join(', ')}`);
      break;
    case 'kyc.manual_review':
      console.log(`⏳ Review manuelle: ${data.id}`);
      break;
    default:
      console.log(`Event inconnu: ${event}`);
  }

  res.json({ received: true });
}

module.exports = { webhookHandler, verifyWebhookSignature };
