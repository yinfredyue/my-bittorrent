package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

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
