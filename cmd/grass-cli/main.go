package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"grassdb/pkg/client"
)

func main() {
	peersStr := flag.String("peers", "localhost:50051,localhost:50052,localhost:50053", "Comma-separated list of peer addresses")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: grass-cli [-peers=...] <command> <args>")
		fmt.Println("Commands:")
		fmt.Println("  set <key> <value>")
		fmt.Println("  get <key>")
		os.Exit(1)
	}

	peers := strings.Split(*peersStr, ",")
	c := client.NewClient(peers)

	command := args[0]
	switch command {
	case "set":
		if len(args) != 3 {
			fmt.Println("Usage: grass-cli set <key> <value>")
			os.Exit(1)
		}
		key, value := args[1], args[2]
		err := c.Set(key, value)
		if err != nil {
			fmt.Printf("Error setting key: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("OK")

	case "get":
		if len(args) != 2 {
			fmt.Println("Usage: grass-cli get <key>")
			os.Exit(1)
		}
		key := args[1]
		val, found, err := c.Get(key)
		if err != nil {
			fmt.Printf("Error getting key: %v\n", err)
			os.Exit(1)
		}
		if !found {
			fmt.Println("(nil)")
		} else {
			fmt.Println(val)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
