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
	"net/netip"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/bittorrent-starter-go/decode"
	"github.com/codecrafters-io/bittorrent-starter-go/encode"
)

const BlockMaxSize = 16 * 1024

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

func parseTorrent(s string) (*Torrent, error) {
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

func (torrent *Torrent) discoverPeers() ([]netip.AddrPort, error) {
	req, err := http.NewRequest("GET", torrent.trackerUrl, nil)
	if err != nil {
		return []netip.AddrPort{}, err
	}

	info_hash, err := torrent.info.hash()
	if err != nil {
		return []netip.AddrPort{}, err
	}

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
	if err != nil {
		return []netip.AddrPort{}, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []netip.AddrPort{}, err
	}

	decoded_resp, err := decode.Decode(string(body))
	if err != nil {
		return []netip.AddrPort{}, err
	}

	decoded_dict := decoded_resp.(map[string](interface{}))
	peers := decoded_dict["peers"].(string)
	peer_addrports := make([]netip.AddrPort, 0)
	for i := 0; i < len(peers); i += 6 {
		// Each peer is represented with 6 bytes.
		// First 4 bytes is IP, where each byte is a number in the IP.
		// Last 2 bytes is port, in big-endian order.
		port := binary.BigEndian.Uint16([]byte(peers[i+4 : i+6]))
		addrBytes := [4]byte{peers[i], peers[i+1], peers[i+2], peers[i+3]}
		addrport := netip.AddrPortFrom(netip.AddrFrom4(addrBytes), port)
		peer_addrports = append(peer_addrports, addrport)
	}

	return peer_addrports, nil
}

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

// Convert uint32 to big-endian bytes. Return a []byte with `numBytes`. Zeros
// are added to the front.
func uint32_to_bytes(val uint32, numBytes int) []byte {
	if numBytes >= 4 {
		res := make([]byte, numBytes, 4)
		binary.BigEndian.PutUint32(res[numBytes-4:], val)
		return res
	}

	res := make([]byte, 4)
	binary.BigEndian.PutUint32(res, val)
	return res[4-numBytes:]
}

func exit_on_error(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func assert(condition bool, errMsg string) {
	if !condition {
		panic(errMsg)
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

		torrent, err := parseTorrent(string(bytes))
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
		exit_on_error(err)

		fmt.Printf("Peer ID: %v\n", hex.EncodeToString(response[48:]))
	} else if command == "download_piece" {
		if len(os.Args) < 6 {
			fmt.Println("Expect: -o output_file torrent_file piece_index")
		}
		outputFilename := os.Args[3]
		filename := os.Args[4]
		piece, err := strconv.Atoi(os.Args[5])
		exit_on_error(err)
		DPrintf("Piece: %v\n", piece)

		bytes, err := os.ReadFile(filename)
		exit_on_error(err)

		torrent, err := parseTorrent(string(bytes))
		exit_on_error(err)

		peers, err := torrent.discoverPeers()
		exit_on_error(err)

		infoHash, err := torrent.info.hash()
		exit_on_error(err)

		peer := peers[0]
		conn, err := net.Dial("tcp", peer.String())
		exit_on_error(err)

		defer conn.Close()

		handshakeMsg := HandshakeMsg(infoHash, []byte("deadbeefliveporkhaha"))
		bytesWritten, err := conn.Write([]byte(handshakeMsg))
		assert(bytesWritten == 68, "Expect to write 68 bytes for handshake")
		exit_on_error(err)

		// handshake response
		response := make([]byte, 68)
		bytesRead, err := conn.Read(response)
		exit_on_error(err)
		assert(bytesRead == 68, "Expect handshake response to be 68 bytes")
		DPrintf("handshake response received\n")

		// bitfield message
		bitfieldMsg := make([]byte, 128)
		bytesRead, err = conn.Read(bitfieldMsg)
		exit_on_error(err)
		assert(bytesRead >= 5, "Expect to read at least 5 bytes")
		assert(uint8(bitfieldMsg[4]) == 5, "bitfield message should have message id = 5")
		DPrintf("bitfield received\n")

		// send interested message
		interestedMsg := [5]byte{0, 0, 0, 5, 2}
		bytesWritten, err = conn.Write(interestedMsg[:])
		assert(bytesWritten == 5, "Expect to write 5 bytes for interested message")
		exit_on_error(err)
		DPrintf("interested message sent\n")

		// unchoke message
		unchokeMsg := make([]byte, 128)
		bytesRead, err = conn.Read(unchokeMsg)
		exit_on_error(err)
		assert(bytesRead == 5, "Expect to read 5 bytes for unchoke message")
		assert(uint8(unchokeMsg[4]) == 1, "unchoke message should have message id = 1")
		DPrintf("unchoke message received\n")

		pieceData := make([]byte, torrent.info.pieceLength)

		// for each block in the piece:
		// send a request message
		// read a piece message
		for blockIdx := 0; blockIdx*BlockMaxSize < torrent.info.pieceLength; blockIdx++ {
			// request message
			var blockSize int
			if (blockIdx+1)*BlockMaxSize < torrent.info.pieceLength {
				blockSize = BlockMaxSize
			} else {
				blockSize = torrent.info.pieceLength - blockIdx*BlockMaxSize
			}
			DPrintf("BlockIdx: %v, BlockSize: %v\n", blockIdx, blockSize)
			blockOffset := blockIdx * blockSize

			// 4-byte message length, 1-byte message id, and a payload of:
			// - 4-byte block index
			// - 4-byte block offset within the piece (in bytes)
			// - 4-byte block length
			// Note: message length starts counting from the message id.
			msgLengthBytes := uint32_to_bytes(13, 4)
			msgIdBytes := uint32_to_bytes(6, 1)
			blockIdxBytes := uint32_to_bytes(uint32(blockIdx), 4)
			blockOffsetBytes := uint32_to_bytes(uint32(blockOffset), 4)
			blockLengthBytes := uint32_to_bytes(uint32(blockSize), 4)
			bytesToWrite := []([]byte){msgLengthBytes, msgIdBytes, blockIdxBytes, blockOffsetBytes, blockLengthBytes}

			var requestMsg [17]byte
			writeIdx := 0
			for _, bytes := range bytesToWrite {
				for _, b := range bytes {
					requestMsg[writeIdx] = b
					writeIdx++
				}
			}
			assert(writeIdx == len(requestMsg), "Expect all bytes written")
			bytesWritten, err = conn.Write(requestMsg[:])
			exit_on_error(err)
			assert(bytesWritten == len(requestMsg), "Expect to send the whole message")

			// piece message
			// 4-byter message length, 1-byte message id, and a payload of
			// - 4-byte block index
			// - 4-byte block offset within the piece (in bytes)
			// - data
			pieceMsgHdr := make([]byte, 5)
			bytesRead, err = conn.Read(pieceMsgHdr)
			exit_on_error(err)
			totalBytesToRead := int(binary.BigEndian.Uint32(pieceMsgHdr[0:4]))
			assert(bytesRead == 5, "Expect to read the full header")
			assert(pieceMsgHdr[4] == 7, "Expect message id = 7")
			totalBytesToRead -= 1

			pieceMsgMetadata := make([]byte, 8)
			bytesRead, err = conn.Read(pieceMsgMetadata)
			exit_on_error(err)
			assert(bytesRead == 8, "Expect to read 8 bytes of metadata")
			receivedBlockIdx := int(binary.BigEndian.Uint32(pieceMsgMetadata[:4]))
			receivedBlockStartOffset := int(binary.BigEndian.Uint32(pieceMsgMetadata[4:8]))
			assert(blockIdx == receivedBlockIdx, "blockIdx doesn't match received")
			assert(blockOffset == receivedBlockStartOffset, "blockOffset doesn't match received")
			totalBytesToRead -= 8

			for writeOffset := 0; totalBytesToRead > 0; {
				bytesRead, err = conn.Read(pieceData[blockOffset+writeOffset:])
				exit_on_error(err)

				totalBytesToRead -= bytesRead
				writeOffset += bytesRead
			}
		}

		err = os.WriteFile(outputFilename, pieceData, 0644)
		exit_on_error(err)

		fmt.Printf("Piece %v downloaded to %v\n", piece, outputFilename)
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}
