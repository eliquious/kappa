package zbase32

import (
  "errors"
)

type encoding struct {
	encodeMap string
	decodeMap [256]byte
}

func newEncoding(encodeMap string) *encoding {
	e := new(encoding)
	e.encodeMap = encodeMap
	for i := 0; i < len(e.decodeMap); i++ {
		e.decodeMap[i] = 0xFF
	}
	for i := 0; i < len(encodeMap); i++ {
		e.decodeMap[encodeMap[i]] = byte(i)
	}
	return e
}

var zBase32Encoding = newEncoding("ybndrfg8ejkmcpqxot1uwisza345h769")

var InsufficientInputError = errors.New("too few bytes supplied for z-base-32 encoding")

func Encode(src []byte, bits int) (string, error) {
	if bits > len(src) * 8 {
    return "", InsufficientInputError
  }

  buf := make([]byte, (bits+4)/5)
	dst := buf

	for i := 0; i < bits; i += 5 {
		b0 := src[0]
		b1 := byte(0)

    if len(src) > 1 {
			b1 = src[1]
		}

		char := byte(0)
    offset := uint(i % 8)

		if offset < 4 {
			char = b0 & (31 << (3 - offset)) >> (3 - offset)
		} else {
			char = b0 & (31 >> (offset - 3)) << (offset - 3)
			char |= b1 & (255 << (11 - offset)) >> (11 - offset)
		}

    // If src is longer than necessary, mask trailing bits to zero
    if i + 5 > bits {
      char &= 255 << uint((i + 5) - bits)
    }

		dst[0] = zBase32Encoding.encodeMap[char]
		dst = dst[1:]

		if offset > 2 {
			src = src[1:]
		}
	}

	return string(buf), nil
}

func EncodeAll(src []byte) (string, error) {
	return Encode(src, len(src)*8)
}
