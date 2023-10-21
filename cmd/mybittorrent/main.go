package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/codecrafters-io/bittorrent-starter-go/decode"
	"github.com/codecrafters-io/bittorrent-starter-go/encode"
)

func exit_on_error(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printDecoded(v interface{}) {
	jsonOutput, _ := json.Marshal(v)
	fmt.Println(string(jsonOutput))
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

		printDecoded(decoded)
	} else if command == "info" {
		if len(os.Args) != 3 {
			fmt.Println("Expect a file name")
			os.Exit(1)
		}

		filename := os.Args[2]
		bytes, err := os.ReadFile(filename)
		exit_on_error(err)

		// decode
		torrent := string(bytes)
		decoded_raw, err := decode.Decode(torrent)
		exit_on_error(err)

		decoded := decoded_raw.(map[string](interface{}))
		trackerUrl := decoded["announce"].(string)
		info := decoded["info"].(map[string](interface{}))
		length := info["length"].(int)
		exit_on_error(err)

		fmt.Printf("Tracker URL: %s\n", trackerUrl)
		fmt.Printf("Length: %d\n", length)

		// extract info hash
		encoded_info, err := encode.Encode(info)
		exit_on_error(err)

		h := sha1.New()
		h.Write([]byte(encoded_info))
		info_hash := h.Sum(nil)

		fmt.Printf("Info Hash: %s\n", hex.EncodeToString(info_hash))

		// extract pieces
		fmt.Printf("Piece Length: %v\n", info["piece length"])
		fmt.Printf("Piece Hashes:\n")
		pieces := info["pieces"].(string)
		for i := 0; i < len(pieces); i += 20 {
			piece_hash := hex.EncodeToString([]byte(pieces[i : i+20]))
			fmt.Printf("%v\n", piece_hash)
		}
	} else if command == "peers" {
		if len(os.Args) != 3 {
			fmt.Println("Expect a file name")
			os.Exit(1)
		}

		filename := os.Args[2]
		bytes, err := os.ReadFile(filename)
		exit_on_error(err)

		// decode
		torrent := string(bytes)
		decoded_raw, err := decode.Decode(torrent)
		exit_on_error(err)

		decoded := decoded_raw.(map[string](interface{}))
		trackerUrl := decoded["announce"].(string)
		info := decoded["info"].(map[string](interface{}))
		length := info["length"].(int)
		exit_on_error(err)

		// extract info hash
		encoded_info, err := encode.Encode(info)
		exit_on_error(err)

		h := sha1.New()
		h.Write([]byte(encoded_info))
		info_hash := h.Sum(nil)

		// discover peers
		req, err := http.NewRequest("GET", trackerUrl, nil)
		exit_on_error(err)

		query := req.URL.Query()
		query.Add("info_hash", string(info_hash))
		query.Add("peer_id", "deadbeefliveporkhaha")
		query.Add("port", "6881")
		query.Add("uploaded", "0")
		query.Add("downloaded", "0")
		query.Add("left", strconv.Itoa(length))
		query.Add("compact", "1")
		req.URL.RawQuery = query.Encode()

		client := http.Client{}
		resp, err := client.Do(req)
		exit_on_error(err)

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		exit_on_error(err)

		decoded_resp, err := decode.Decode(string(body))
		exit_on_error(err)

		decoded_dict := decoded_resp.(map[string](interface{}))
		peers := decoded_dict["peers"].(string)
		for i := 0; i < len(peers); i += 6 {
			// Each peer is represented with 6 bytes.
			// First 4 bytes is IP, where each byte is a number in the IP.
			// Last 2 bytes is port, in big-endian order.
			ip := fmt.Sprintf("%v:%v:%v:%v", peers[i], peers[i+1], peers[i+2], peers[i+3])
			port := binary.BigEndian.Uint16([]byte(peers[i+4 : i+6]))
			fmt.Printf("%v:%v\n", ip, port)
		}
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
