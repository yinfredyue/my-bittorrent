package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/bittorrent-starter-go/decode"
	"github.com/codecrafters-io/bittorrent-starter-go/encode"
)

type Info struct {
	length      int
	name        string
	pieceLength int
	pieces      [](string) // binary format, not hex format
}

func (info *Info) hash() ([]byte, error) {
	dict := map[string](interface{}){
		"length":       info.length,
		"name":         info.name,
		"piece length": info.pieceLength,
		"pieces":       strings.Join(info.pieces, ""),
	}

	encoded_info, err := encode.Encode(dict)
	if err != nil {
		return []byte{}, err
	}

	h := sha1.New()
	h.Write([]byte(encoded_info))
	info_hash := h.Sum(nil)

	return info_hash, nil
}

type Torrent struct {
	trackerUrl string
	info       Info
}

func decodeTorrent(s string) (*Torrent, error) {
	decoded_raw, err := decode.Decode(s)
	if err != nil {
		return nil, err
	}

	decoded := decoded_raw.(map[string](interface{}))
	trackerUrl := decoded["announce"].(string)
	info_dict := decoded["info"].(map[string](interface{}))
	length := info_dict["length"].(int)
	name := info_dict["name"].(string)
	pieceLength := info_dict["piece length"].(int)
	if err != nil {
		return nil, err
	}

	pieces_raw := info_dict["pieces"].(string)
	pieces := make([](string), 0)
	for i := 0; i < len(pieces_raw); i += 20 {
		pieceHash := (pieces_raw[i : i+20])
		pieces = append(pieces, pieceHash)
	}

	info := Info{
		length:      length,
		name:        name,
		pieceLength: pieceLength,
		pieces:      pieces,
	}

	torrent := Torrent{
		trackerUrl: trackerUrl,
		info:       info,
	}

	return &torrent, nil
}

func exit_on_error(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
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

		torrent, err := decodeTorrent(string(bytes))
		exit_on_error(err)

		fmt.Printf("Tracker URL: %s\n", torrent.trackerUrl)
		fmt.Printf("Length: %d\n", torrent.info.length)

		info_hash, err := torrent.info.hash()
		exit_on_error(err)
		fmt.Printf("Info Hash: %s\n", hex.EncodeToString(info_hash))

		fmt.Printf("Piece Length: %v\n", torrent.info.pieceLength)
		fmt.Printf("Piece Hashes:\n")
		for _, piece_hash := range torrent.info.pieces {
			fmt.Printf("%v\n", hex.EncodeToString([]byte(piece_hash)))
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
		torrent, err := decodeTorrent(string(bytes))
		exit_on_error(err)

		// discover peers
		req, err := http.NewRequest("GET", torrent.trackerUrl, nil)
		exit_on_error(err)

		info_hash, err := torrent.info.hash()
		exit_on_error(err)

		query := req.URL.Query()
		query.Add("info_hash", string(info_hash))
		query.Add("peer_id", "deadbeefliveporkhaha")
		query.Add("port", "6881")
		query.Add("uploaded", "0")
		query.Add("downloaded", "0")
		query.Add("left", strconv.Itoa(torrent.info.length))
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
			ip := fmt.Sprintf("%v.%v.%v.%v", peers[i], peers[i+1], peers[i+2], peers[i+3])
			port := binary.BigEndian.Uint16([]byte(peers[i+4 : i+6]))
			fmt.Printf("%v:%v\n", ip, port)
		}
	} else if command == "handshake" {
		if len(os.Args) != 4 {
			fmt.Println("Expect a file name and a peer address")
			os.Exit(1)
		}

		filename := os.Args[2]
		peer_address := os.Args[3]

		bytes, err := os.ReadFile(filename)
		exit_on_error(err)

		torrent, err := decodeTorrent(string(bytes))
		exit_on_error(err)

		info_hash, err := torrent.info.hash()
		exit_on_error(err)

		conn, err := net.Dial("tcp", peer_address)
		exit_on_error(err)

		defer conn.Close()

		var sb strings.Builder
		sb.WriteByte(19)
		sb.WriteString("BitTorrent protocol") // Don't capitalize "protocol"
		for i := 0; i < 8; i++ {
			sb.WriteByte(0)
		}
		sb.Write(info_hash)
		sb.Write([]byte("deadbeefliveporkhaha"))

		_, err = conn.Write([]byte(sb.String()))
		exit_on_error(err)

		response := make([]byte, 68)
		_, err = conn.Read(response)
		exit_on_error(err)

		fmt.Printf("Peer ID: %v\n", hex.EncodeToString(response[48:]))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
