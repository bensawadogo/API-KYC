# kyc-go

SDK officiel DATAKEYS KYC · Vérification d'identité panafricaine
55 pays · Conforme BCEAO · Sandbox inclus

## Installation

```bash
go get github.com/datakeys/kyc-go
```

## 5 minutes pour intégrer

```go
package main

import (
    "context"
    "fmt"
    datakeys "github.com/datakeys/kyc-go"
)

func main() {
    dk := datakeys.MustNew("dk_test_datakeys_sandbox_001")
    ctx := context.Background()

    v, _ := dk.KYC.Initiate(ctx, datakeys.InitiateParams{
        Phone:       "+22670000001",
        CountryCode: "BF",
        DocType:     datakeys.DocNationalID,
        FullName:    "Aminata Ouédraogo",
        Consent:     true,
    })

    result, _ := dk.KYC.WaitForCompletion(ctx, v.ID, 0)
    fmt.Println(result.Status) // "approved"
    fmt.Println(result.Score)  // 0.95
}
```

## Gestion des erreurs

```go
dk, _ := datakeys.New(apiKey)

v, err := dk.KYC.Initiate(ctx, params)
if err != nil {
    var kycErr *datakeys.KYCError
    if errors.As(err, &kycErr) {
        fmt.Println(kycErr.Code)   // "KYC_AUTH_002"
        fmt.Println(kycErr.Status) // 401
        if kycErr.IsRateLimit() {
            // Attendre avant de réessayer
        }
    }
}
```

## Profils sandbox

| phone          | pays | scénario  |
|---------------|------|-----------|
| +22670000001  | BF   | approved  |
| +233200001111 | GH   | rejected  |
| +254700002222 | KE   | sanctions |
| +2250700003333| CI   | pep       |

Voir tous les profils : `curl http://localhost:8081/sandbox/profiles`
