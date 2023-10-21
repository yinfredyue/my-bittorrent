package tests

import (
	"testing"

	"github.com/codecrafters-io/bittorrent-starter-go/encode"
)

func testEncodeHelper(t *testing.T, v interface{}, expected string) {
	encoded, err := encode.Encode(v)
	if err != nil {
		t.Fatal(err)
	}

	if encoded != expected {
		t.Fatalf("Expected: %v, result: %v", expected, encoded)
	}
}

func TestEncode(t *testing.T) {
	testEncodeHelper(t, 52, "i52e")
	testEncodeHelper(t, -52, "i-52e")
	testEncodeHelper(t, "hello", "5:hello")
	testEncodeHelper(t, [](interface{}){"hello", 52, [](interface{}){"s", "ss", 32}}, "l5:helloi52el1:s2:ssi32eee")
	testEncodeHelper(t, map[string](interface{}){"foo": "bar", "hello": 52}, "d3:foo3:bar5:helloi52ee")
}
