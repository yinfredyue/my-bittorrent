package tests

import (
	"reflect"
	"testing"

	"github.com/codecrafters-io/bittorrent-starter-go/decode"
)

func testDecodeHelper(t *testing.T, str string, expected interface{}) {
	result, err := decode.Decode(str)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, result) {
		t.Fatalf("Mismatch! Expected: %v, result: %v", expected, result)
	}
}

func TestDecode(t *testing.T) {
	testDecodeHelper(t, "i52e", 52)
	testDecodeHelper(t, "i-52e", -52)
	testDecodeHelper(t, "5:hello", "hello")
	testDecodeHelper(t, "l5:helloi52el1:s2:ssi32eee", [](interface{}){"hello", 52, [](interface{}){"s", "ss", 32}})
	testDecodeHelper(t, "d3:foo3:bar5:helloi52ee", map[string](interface{}){"foo": "bar", "hello": 52})
}
