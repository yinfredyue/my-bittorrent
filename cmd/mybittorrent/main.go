package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/bittorrent-starter-go/decode"
)

const BlockMaxSize = 16 * 1024

func HandshakeMsg(infoHash []byte, peerId []byte) string {
	var sb strings.Builder
	sb.WriteByte(19)
	sb.WriteString("BitTorrent protocol") // Don't capitalize "protocol"
	for i := 0; i < 8; i++ {
		sb.WriteByte(0)
	}
	sb.Write(infoHash)
	sb.Write([]byte("deadbeefliveporkhaha"))

	return sb.String()
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

		torrent, err := parseTorrent(string(bytes))
		exit_on_error(err)

		fmt.Printf("Tracker URL: %s\n", torrent.trackerUrl)
		fmt.Printf("Length: %d\n", torrent.info.length)

		info_hash, err := torrent.info.hash()
		exit_on_error(err)
		fmt.Printf("Info Hash: %s\n", hex.EncodeToString(info_hash))

		fmt.Printf("Piece Length: %v\n", torrent.info.pieceLength)
		// fmt.Printf("Piece Hashes:\n")
		// for _, piece_hash := range torrent.info.pieces {
		// 	fmt.Printf("%v\n", hex.EncodeToString([]byte(piece_hash)))
		// }
	} else if command == "peers" {
		if len(os.Args) != 3 {
			fmt.Println("Expect a file name")
			os.Exit(1)
		}

		filename := os.Args[2]
		bytes, err := os.ReadFile(filename)
		exit_on_error(err)

		torrent, err := parseTorrent(string(bytes))
		exit_on_error(err)

		peers, err := torrent.discoverPeers()
		exit_on_error(err)

		for _, peer := range peers {
			fmt.Printf("%v\n", peer)
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

		torrent, err := parseTorrent(string(bytes))
		exit_on_error(err)

		infoHash, err := torrent.info.hash()
		exit_on_error(err)

		conn, err := net.Dial("tcp", peer_address)
		exit_on_error(err)

		defer conn.Close()

		handshakeMsg := HandshakeMsg(infoHash, []byte("deadbeefliveporkhaha"))
		_, err = conn.Write([]byte(handshakeMsg))
		exit_on_error(err)

		response := make([]byte, 68)
		_, err = conn.Read(response)
		if err != nil && !errors.Is(err, io.EOF) {
			exit_on_error(err)
		}

		fmt.Printf("Peer ID: %v\n", hex.EncodeToString(response[48:]))
	} else if command == "download_piece" {
		if len(os.Args) != 6 {
			fmt.Println("Expect: -o output_file torrent_file piece_index")
		}
		outputFilename := os.Args[3]
		filename := os.Args[4]
		piece, err := strconv.Atoi(os.Args[5])
		exit_on_error(err)
		DPrintf("Downloading piece: %v\n", piece)

		bytes, err := os.ReadFile(filename)
		exit_on_error(err)

		torrent, err := parseTorrent(string(bytes))
		exit_on_error(err)

		pieceData, err := torrent.downloadPiece(piece)
		exit_on_error(err)

		err = os.WriteFile(outputFilename, pieceData, 0644)
		exit_on_error(err)

		fmt.Printf("Piece %v downloaded to %v\n", piece, outputFilename)
	} else if command == "download" {
		if len(os.Args) != 5 {
			fmt.Println("Expect: -o output_file torrent_file")
		}
		outputFilename := os.Args[3]
		torrentFilename := os.Args[4]

		bytes, err := os.ReadFile(torrentFilename)
		exit_on_error(err)

		torrent, err := parseTorrent(string(bytes))
		exit_on_error(err)

		// err = torrent.downloadFile(outputFilename)
		err = torrent.downloadFilePipelining(outputFilename)
		exit_on_error(err)

		fmt.Printf("Downloaded %v to %v\n", torrent.info.name, outputFilename)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
