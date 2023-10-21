package decode

import (
	"reflect"
	"testing"

	"github.com/codecrafters-io/bittorrent-starter-go/decode"
)

func testHelper(t *testing.T, str string, expected [](interface{})) {
	result, err := decode.Decode(str)
	if err != nil {
		t.Fatal(err)
	}

	if len(expected) != len(result) {
		t.Fatalf("Different length. Expected: %v, result: %v", expected, result)
	}

	for i := 0; i < len(expected); i++ {
		expected_val := expected[i]
		result_val := result[i]
		if !reflect.DeepEqual(expected_val, result_val) {
			t.Fatalf("Mismatch! Expected: %v, result: %v", expected_val, result_val)
		}
	}
}

func TestDecode(t *testing.T) {
	testHelper(t, "i52e", [](interface{}){52})
	testHelper(t, "i-52e", [](interface{}){-52})
	testHelper(t, "5:hello", [](interface{}){"hello"})
	testHelper(t, "l5:helloi52el1:s2:ssi32eee", [](interface{}){[](interface{}){"hello", 52, [](interface{}){"s", "ss", 32}}})
}
