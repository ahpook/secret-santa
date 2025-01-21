package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
	"github.com/nspcc-dev/neo-go/pkg/wallet"
)

// Replace with your own private key, RPC endpoint, etc.
const (
	rpcEndpoint  = "http://127.0.0.1:30333"
	wif          = "<YOUR_WIF_HERE>"
	contractHash = "0x1234567890abcdef..." // The deployed contract script hash
)

func main() {
	// Example usage of the backend to interact with the contract
	cli, err := client.New(context.Background(), rpcEndpoint, client.Options{})
	if err != nil {
		log.Fatalf("Failed to create RPC client: %v", err)
	}
	defer cli.Close()

	// Load a wallet account
	w, err := wallet.NewWalletFromWIF(wif)
	if err != nil {
		log.Fatalf("Failed to load wallet from WIF: %v", err)
	}
	acc := w.Accounts[0]

	// You may need to sign transactions with your private key
	err = acc.Decrypt("") // if WIF is not password protected, pass empty string
	if err != nil {
		log.Fatalf("Failed to decrypt account: %v", err)
	}

	// Example: call "AddParticipant" to add a user
	txHash, err := cli.InvokeFunction(contractHash, "addParticipant", []interface{}{"Alice"}, acc, nil)
	if err != nil {
		log.Fatalf("InvokeFunction error: %v", err)
	}
	fmt.Printf("AddParticipant(Alice) TX: %s\n", txHash)

	// You can do other calls in a similar manner,
	// e.g. shuffleAndStorePairs, getAllParticipants, etc.

	// Example: call "shuffleAndStorePairs"
	txHash, err = cli.InvokeFunction(contractHash, "shuffleAndStorePairs", []interface{}{}, acc, nil)
	if err != nil {
		log.Fatalf("shuffleAndStorePairs error: %v", err)
	}
	fmt.Printf("shuffleAndStorePairs TX: %s\n", txHash)
}
