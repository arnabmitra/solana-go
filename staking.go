package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/gagliardetto/solana-go/rpc"
)

type Validator struct {
	NodePubkey     string
	Commission     int
	ActivatedStake uint64
}

func main() {
	// Connect to Solana mainnet
	client := rpc.New(rpc.MainNetBeta_RPC)

	// Fetch the list of validators
	voteAccounts, err := client.GetVoteAccounts(context.Background(), &rpc.GetVoteAccountsOpts{})
	if err != nil {
		log.Fatalf("failed to get vote accounts: %v", err)
	}

	// Combine current and delinquent validators
	var validators []Validator
	for _, v := range voteAccounts.Current {
		validators = append(validators, Validator{v.NodePubkey.String(), int(v.Commission), v.ActivatedStake})
	}
	for _, v := range voteAccounts.Delinquent {
		validators = append(validators, Validator{v.NodePubkey.String(), int(v.Commission), v.ActivatedStake})
	}

	// Sort validators by staking amount
	sort.Slice(validators, func(i, j int) bool {
		return validators[i].ActivatedStake > validators[j].ActivatedStake
	})

	// Create a CSV file
	file, err := os.Create("validators.csv")
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write to CSV
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"NodePubkey", "Commission", "ActivatedStake"})

	// Write validator data
	for _, v := range validators {
		writer.Write([]string{v.NodePubkey, fmt.Sprintf("%d", v.Commission), fmt.Sprintf("%d", v.ActivatedStake)})
	}

	fmt.Println("Validators sorted by staking amount and saved to validators.csv")
}
