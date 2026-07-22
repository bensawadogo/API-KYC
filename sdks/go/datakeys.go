// Package datakeys est le SDK officiel DATAKEYS KYC.
//
// Vérification d'identité panafricaine — 55 pays.
// Conforme BCEAO. Sandbox inclus.
//
// Usage:
//
//	dk := datakeys.New("dk_test_datakeys_sandbox_001")
//	v, err := dk.KYC.Initiate(ctx, datakeys.InitiateParams{
//	    Phone:       "+22670000001",
//	    CountryCode: "BF",
//	    DocType:     datakeys.DocNationalID,
//	    FullName:    "Aminata Ouédraogo",
//	    Consent:     true,
//	})
package datakeys

// Datakeys est le client principal du SDK.
type Datakeys struct {
	KYC       *KYCService
	Countries *CountriesService
	Livemode  bool
}

// New crée un client DATAKEYS KYC.
// La clé préfixée dk_test_ utilise automatiquement le sandbox.
// La clé préfixée dk_live_ utilise la production.
func New(apiKey string, cfg ...Config) (*Datakeys, error) {
	var config Config
	if len(cfg) > 0 {
		config = cfg[0]
	}

	c, err := newClient(apiKey, config)
	if err != nil {
		return nil, err
	}

	return &Datakeys{
		KYC:       &KYCService{client: c},
		Countries: &CountriesService{client: c},
		Livemode:  c.Livemode,
	}, nil
}

// MustNew est comme New mais panique si la clé est invalide.
func MustNew(apiKey string, cfg ...Config) *Datakeys {
	dk, err := New(apiKey, cfg...)
	if err != nil {
		panic(err)
	}
	return dk
}
