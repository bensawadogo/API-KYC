package main

import (
	"context"
	"fmt"
	"log"

	datakeys "github.com/datakeys/kyc-go"
)

func main() {
	dk := datakeys.MustNew("dk_test_datakeys_sandbox_001")
	fmt.Printf("Mode: livemode=%v\n", dk.Livemode)

	ctx := context.Background()

	v, err := dk.KYC.Initiate(ctx, datakeys.InitiateParams{
		Phone:       "+22670000001",
		CountryCode: "BF",
		DocType:     datakeys.DocNationalID,
		DocNumber:   "B1230000",
		FullName:    "Aminata Ouédraogo",
		Consent:     true,
	})
	if err != nil {
		log.Fatalf("Initiate error: %v", err)
	}
	fmt.Printf("ID:       %s\n", v.ID)
	fmt.Printf("Statut:   %s\n", v.Status)
	fmt.Printf("Provider: %s\n", v.Provider)

	result, err := dk.KYC.WaitForCompletion(ctx, v.ID, 0,
		datakeys.WithPollCallback(func(v *datakeys.KYCVerification) {
			fmt.Printf("  polling... status=%s\n", v.Status)
		}),
	)
	if err != nil {
		log.Fatalf("WaitForCompletion error: %v", err)
	}

	fmt.Printf("\n=== RÉSULTAT ===\n")
	fmt.Printf("Statut:    %s\n", result.Status)
	fmt.Printf("Score:     %.2f\n", result.Score)
	fmt.Printf("Approuvé:  %v\n", result.IsApproved())
}
