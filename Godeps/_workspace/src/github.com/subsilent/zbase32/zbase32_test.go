package zbase32

import (
	"testing"
)

type testcase struct {
	bits    int
	decoded []byte
	encoded string
}

var testcases = []testcase{
  // Test cases from the spec
	{0, []byte{}, ""},
	{1, []byte{0}, "y"},
	{1, []byte{128}, "o"},
	{2, []byte{64}, "e"},
	{2, []byte{192}, "a"},
	{10, []byte{0, 0}, "yy"},
	{10, []byte{128, 128}, "on"},
  {20, []byte{139, 136, 128}, "tqre"},
  {24, []byte{240, 191, 199}, "6n9hq"},
  {24, []byte{212, 122, 4}, "4t7ye"},
  // Note: this test varies from what's in the spec by one character!
  {30, []byte{245, 87, 189, 12}, "6im54d"},

  // Test masking of excess trailing bits
  {0, []byte{255, 255}, ""},
  {1, []byte{255, 255}, "o"},
  {2, []byte{255, 255}, "a"},
  {3, []byte{255, 255}, "h"},
  {4, []byte{255, 255}, "6"},
  {5, []byte{255, 255}, "9"},
  {6, []byte{255, 255}, "9o"},
  {7, []byte{255, 255}, "9a"},
  {8, []byte{255, 255}, "9h"},
  {9, []byte{255, 255}, "96"},
  {10, []byte{255, 255}, "99"},
  {11, []byte{255, 255}, "99o"},
  {12, []byte{255, 255}, "99a"},
  {13, []byte{255, 255}, "99h"},
  {14, []byte{255, 255}, "996"},
  {15, []byte{255, 255}, "999"},
  {16, []byte{255, 255}, "999o"},
}

func TestEncode(t *testing.T) {
	for _, tc := range testcases {
		got, err := Encode(tc.decoded, tc.bits)
    if err != nil {
			t.Errorf("Encode(0x%x, %v): error: %v", tc.decoded, tc.bits, err)
      continue
    }

		if got != tc.encoded {
			t.Errorf("Encode(0x%x, %v) = %q, expected %q", tc.decoded, tc.bits, got, tc.encoded)
		}
	}
}

func TestEncodeLengthChecking(t *testing.T) {
  decoded := []byte{128, 128, 128, 128}
  bits := 33
  got, err := Encode(decoded, bits)
  if err == nil {
		t.Errorf("Encode(0x%x, %v) = %q, expecting error", decoded, bits, got)
  }
}
