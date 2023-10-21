package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/decode"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No command")
		os.Exit(1)
	}
	command := os.Args[1]

	if command == "decode" {
		if len(os.Args) != 3 {
			fmt.Println("Expect a single argument")
			os.Exit(1)
		}
		bencodedValue := os.Args[2]

		decoded, err := decode.Decode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
