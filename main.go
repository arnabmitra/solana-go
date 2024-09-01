package main

import (
	"context"
	"fmt"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/blocto/solana-go-sdk/program/system"
	"github.com/blocto/solana-go-sdk/types"
	"github.com/mr-tron/base58"
	"log"
	"sync"
	"time"

	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
)

func main() {
	// Initialize the Solana client
	c := client.NewClient(rpc.DevnetRPCEndpoint)

	// Define the sender and receiver public keys
	receiver := common.PublicKeyFromString("DxPv2QMA5cWR5Xfg7tXr5YtJ1EEStg5Kiag9HhkY1mSx")

	// Define the sender's private key
	decode, err := base58.Decode("2WGcYYau2gLu2DUq68SxxXQmCgi77n8hFqqLNbNyg6Xfh2m3tvg8LF5Lgh69CFDux41LUKV1ak1ERHUqiBZnyshz")
	if err != nil {
		log.Fatalf("failed to decode private key, err: %v", err)
	}

	senderPrivateKey, err := types.AccountFromBytes(decode)
	if err != nil {
		log.Fatalf("failed to decode private key, err: %v", err)
	}

	// Get the recent blockhash
	recentBlockhash, err := c.GetLatestBlockhash(context.TODO())
	if err != nil {
		log.Fatalf("failed to get recent blockhash, err: %v", err)
	}

	// Create a new transaction
	tx, err := types.NewTransaction(types.NewTransactionParam{
		Signers: []types.Account{senderPrivateKey},
		Message: types.NewMessage(types.NewMessageParam{
			FeePayer:        senderPrivateKey.PublicKey,
			RecentBlockhash: recentBlockhash.Blockhash,
			Instructions: []types.Instruction{
				system.Transfer(system.TransferParam{
					From:   senderPrivateKey.PublicKey,
					To:     receiver,
					Amount: 1000, // Transfer 1000 lamports
				}),
			},
		}),
	})
	if err != nil {
		log.Fatalf("failed to create transaction, err: %v", err)
	}

	// Send the transaction
	txSig, err := c.SendTransaction(context.TODO(), tx)
	if err != nil {
		log.Fatalf("failed to send transaction, err: %v", err)
	}

	// Print the transaction signature
	fmt.Printf("Transaction signature: %s\n", txSig)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		// Retry mechanism to check the transaction status
		for i := 0; i < 10; i++ {
			txInfo, err := c.GetTransaction(context.TODO(), txSig)
			if err != nil {
				log.Printf("failed to get transaction, err: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			if txInfo != nil {
				// Print the transaction info in a human-readable format
				fmt.Printf("Transaction Info:\n")
				fmt.Printf("  Slot: %d\n", txInfo.Slot)
				fmt.Printf("  Block Time: %d\n", txInfo.BlockTime)
				fmt.Printf("  Transaction:\n")
				fmt.Printf("    Signatures: %v\n", txInfo.Transaction.Signatures)
				fmt.Printf("    Message:\n")
				fmt.Printf("      Account Keys: %v\n", txInfo.Transaction.Message.Accounts)
				fmt.Printf("      Instructions: %v\n", txInfo.Transaction.Message.Instructions)
				fmt.Printf("      Recent Blockhash: %s\n", txInfo.Transaction.Message.RecentBlockHash)

				// Print balance changes by account
				for i, account := range txInfo.Transaction.Message.Accounts {
					fmt.Printf("  Account: %s\n", account)
					fmt.Printf("    Balance Before: %d\n", txInfo.Meta.PreBalances[i])
					fmt.Printf("    Balance After: %d\n", txInfo.Meta.PostBalances[i])
				}
				// Print fee paid
				fmt.Printf("  Fee Paid: %d\n", txInfo.Meta.Fee)
				return
			}
			time.Sleep(2 * time.Second)
		}
		log.Println("transaction not found after multiple attempts")
	}()

	// Wait for the goroutine to finish
	wg.Wait()

	// Fetch confirmed signatures for the address
	signatures, err := c.GetSignaturesForAddress(context.TODO(), senderPrivateKey.PublicKey.String())
	if err != nil {
		log.Fatalf("failed to get confirmed signatures, err: %v", err)
	}

	// Fetch transaction details for each signature
	for _, sig := range signatures {
		tx, err := c.GetTransaction(context.TODO(), sig.Signature)
		txInfo, err := c.GetTransaction(context.TODO(), txSig)
		if err != nil {
			log.Printf("failed to get transaction, err: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if txInfo != nil {
			// Print the transaction info in a human-readable format
			fmt.Printf("Transaction Info:\n")
			fmt.Printf("  Slot: %d\n", txInfo.Slot)
			fmt.Printf("  Block Time: %d\n", txInfo.BlockTime)
			fmt.Printf("  Transaction:\n")
			fmt.Printf("    Signatures: %v\n", txInfo.Transaction.Signatures)
			fmt.Printf("    Message:\n")
			fmt.Printf("      Account Keys: %v\n", txInfo.Transaction.Message.Accounts)
			fmt.Printf("      Instructions: %v\n", txInfo.Transaction.Message.Instructions)
			fmt.Printf("      Recent Blockhash: %s\n", txInfo.Transaction.Message.RecentBlockHash)

			// Print balance changes by account
			for i, account := range txInfo.Transaction.Message.Accounts {
				fmt.Printf("  Account: %s\n", account)
				fmt.Printf("    Balance Before: %d\n", txInfo.Meta.PreBalances[i])
				fmt.Printf("    Balance After: %d\n", txInfo.Meta.PostBalances[i])
			}
			// Print fee paid
			fmt.Printf("  Fee Paid: %d\n", txInfo.Meta.Fee)
			if err != nil {
				log.Printf("failed to get transaction for signature %s, err: %v", sig.Signature, err)
				continue
			}
			fmt.Printf("Transaction for signature %s: %+v\n", sig.Signature, tx)
		}
	}
}
