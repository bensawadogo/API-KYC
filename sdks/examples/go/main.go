package main

import (
	"context"
	"fmt"
	"log"

	"github.com/datakeys/kyc-go"
)

func main() {
	dk := datakeys.New("dk_test_datakeys_sandbox_001")
	ctx := context.Background()

	v, err := dk.KYC.Initiate(ctx, datakeys.InitiateParams{
		Phone:       "+22670000001",
		CountryCode: "BF",
		DocType:     "NATIONAL_ID",
		DocNumber:   "B1230000",
		FullName:    "Aminata Ouédraogo",
		Consent:     true,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("ID:", v.ID)
	fmt.Println("Statut:", v.Status)

	result, err := dk.KYC.WaitForCompletion(ctx, v.ID, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Résultat: %s (score: %.2f)\n", result.Status, result.Score)
}
