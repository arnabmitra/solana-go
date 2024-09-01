package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	"github.com/blocto/solana-go-sdk/types"
)

func main() {
	// Generate a new keypair
	newAccount := types.NewAccount()

	// Retrieve the public key
	publicKey := newAccount.PublicKey.ToBase58()
	fmt.Printf("Generated Public Key: %s\n", publicKey)

	// Initialize the Solana client
	c := client.NewClient(rpc.DevnetRPCEndpoint)

	// Request airdrop
	txSig, err := c.RequestAirdrop(context.TODO(), newAccount.PublicKey.ToBase58(), 1000000000)
	if err != nil {
		log.Fatalf("failed to request airdrop, err: %v", err)
	}

	// Print the transaction signature
	fmt.Printf("Airdrop requested with transaction signature: %s\n", txSig)

	// Check the transaction status
	resp, err := http.Get(fmt.Sprintf("https://api.devnet.solana.com?method=getTransaction&params=[\"%s\"]", txSig))
	if err != nil {
		log.Fatalf("failed to get transaction status, err: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response body, err: %v", err)
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("failed to parse response body, err: %v", err)
	}

	// Print the funding status
	fmt.Printf("Funding Status: %v\n", result)
}
