package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/codecrafters-io/bittorrent-starter-go/decode"
	"github.com/codecrafters-io/bittorrent-starter-go/encode"
)

func exit_on_error(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

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
		exit_on_error(err)

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else if command == "info" {
		if len(os.Args) != 3 {
			fmt.Println("Expect a file name")
			os.Exit(1)
		}

		filename := os.Args[2]
		bytes, err := os.ReadFile(filename)
		exit_on_error(err)

		torrent := string(bytes)
		decoded_raw, err := decode.Decode(torrent)
		exit_on_error(err)

		decoded := decoded_raw.(map[string](interface{}))
		info := decoded["info"].(map[string](interface{}))

		fmt.Printf("Tracker URL: %s\n", decoded["announce"])
		fmt.Printf("Length: %d\n", info["length"])

		encoded_info, err := encode.Encode(info)
		exit_on_error(err)

		h := sha1.New()
		h.Write([]byte(encoded_info))
		hash := hex.EncodeToString(h.Sum(nil))

		fmt.Printf("Info hash: %s\n", hash)
		fmt.Printf("Piece Length: %v\n", info["piece length"])

		fmt.Printf("Piece Hashes:\n")
		pieces := info["pieces"].(string)
		for i := 0; i < len(pieces); i += 20 {
			piece_hash := hex.EncodeToString([]byte(pieces[i : i+20]))
			fmt.Printf("%v\n", piece_hash)
		}
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
