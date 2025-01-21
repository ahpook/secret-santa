package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nspcc-dev/neo-go/pkg/rpc/client"
)

// Adjust to your environment
const (
	rpcEndpoint  = "http://127.0.0.1:30333"
	contractHash = "0x1234567890abcdef..."
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <operation> [args...]")
		return
	}

	operation := os.Args[1]
	args := os.Args[2:] // remaining arguments

	cli, err := client.New(context.Background(), rpcEndpoint, client.Options{})
	if err != nil {
		log.Fatalf("Failed to create RPC client: %v", err)
	}
	defer cli.Close()

	switch operation {
	case "getAllParticipants":
		result, err := cli.InvokeFunction(contractHash, "getAllParticipants", nil, nil, nil)
		if err != nil {
			log.Fatalf("InvokeFunction error: %v", err)
		}
		fmt.Printf("Participants: %+v\n", result)
	case "getPairedWith":
		if len(args) < 1 {
			log.Fatal("Missing argument: name")
		}
		name := args[0]
		result, err := cli.InvokeFunction(contractHash, "getPairedWith", []interface{}{name}, nil, nil)
		if err != nil {
			log.Fatalf("InvokeFunction error: %v", err)
		}
		fmt.Printf("Paired with: %+v\n", result)
	default:
		fmt.Println("Unknown operation")
	}
}
