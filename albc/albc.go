// Package albc provides utilities related to Apple Activation Lock Bypass Codes.
package albc

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

// charset is available character set for a dash-separated "human readable" bypass code string.
// Note these 32 characters fit within a 5 bit encoding and are looked up by index position.
const charset = "0123456789ACDEFGHJKLMNPQRTUVWXYZ"

var dashPositions = []int{5, 10, 14, 18, 22}

// BypassCode is the "raw" form of an Apple Activation Lock Bypass Code.
// See https://developer.apple.com/documentation/devicemanagement/creating-and-using-bypass-codes
type BypassCode [16]byte

// New creates a new random bypass code.
func New() (BypassCode, error) {
	return NewFromReader(rand.Reader)
}

// NewFromReader reads 16 bytes from reader to populate a new bypass code.
func NewFromReader(reader io.Reader) (BypassCode, error) {
	if reader == nil {
		return BypassCode{}, errors.New("nil reader")
	}

	buf := make([]byte, 16)
	_, err := reader.Read(buf)
	bc := BypassCode{}
	copy(bc[:], buf)
	return bc, err
}

// NewFromBytes creates a new bypass code from b.
// The first 16 bytes are copied to populate a new bypass code.
func NewFromBytes(b []byte) (BypassCode, error) {
	if len(b) < 16 {
		return BypassCode{}, fmt.Errorf("invalid length: %d", len(b))
	}

	var out BypassCode
	copy(out[:], b)
	return out, nil
}

// NewFromCode decodes a dash-separated "human readable" bypass code.
func NewFromCode(code string) (BypassCode, error) {
	var bc5 []byte
	for _, r := range code {
		for i, c := range charset {
			if c == r {
				// naively uses any charset rune and
				// skips over anything else
				bc5 = append(bc5, byte(i))
				break
			}
		}
	}

	var out BypassCode

	ret, err := convertBits(bc5, 5, 8)
	if err != nil {
		return out, fmt.Errorf("converting bits: %w", err)
	}

	if len(ret) != 16 {
		return out, fmt.Errorf("invalid length: %d", len(ret))
	}

	copy(out[:], ret)
	return out, nil
}

// Hash generates a hex encoded PBKDF2 derived hash of c.
// This hash is used for e.g. activation locking and unlocking the device using Apple APIs.
// Apple describes the hash as SHA256 with static salt and fixed iterations.
func (c BypassCode) Hash() (string, error) {
	pb, err := pbkdf2.Key(
		sha256.New,
		string(c[:]),
		[]byte{0, 0, 0, 0},
		50000,
		sha256.Size,
	)
	return hex.EncodeToString(pb), err
}

// Code generates the dash-separated "human readable" form of c.
func (c BypassCode) Code() (string, error) {
	bc5, err := convertBits(c[:], 8, 5)
	if err != nil {
		return "", err
	}

	j := 0
	var str strings.Builder
	for i, p := range bc5 {
		if j < len(dashPositions) && i == dashPositions[j] {
			str.WriteString("-")
			j++
		}
		str.WriteByte(charset[p])
	}

	return str.String(), nil
}

// convert binary data from one bits-per-byte arrangement to another.
// Ex: re-arrange 8 bit bytes to groups of 5 when converting to base32.
// This is a modified helper from a Go implementation of the bech32 format.
// https://github.com/FiloSottile/age/blob/c9a35c072716b5ac6cd815366999c9e189b0c317/internal/bech32/bech32.go#L79-L105
func convertBits(data []byte, frombits, tobits byte) ([]byte, error) {
	var ret []byte
	acc := uint32(0)
	bits := byte(0)
	maxv := byte(1<<tobits - 1)
	for idx, value := range data {
		if value>>frombits != 0 {
			return nil, fmt.Errorf("invalid data range: data[%d]=%d (frombits=%d)", idx, value, frombits)
		}
		acc = acc<<frombits | uint32(value)
		bits += frombits
		for bits >= tobits {
			bits -= tobits
			ret = append(ret, byte(acc>>bits)&maxv)
		}
	}

	// the remainder bits are where Apple differs from bech32
	if bits > 0 {
		// zero out most significant bits of the remainder value
		remainderMask := byte(0xff) << bits
		remainderByte := byte(acc) & ^remainderMask & maxv

		if tobits > frombits {
			// re-pack the remainder value into the last value.
			// note there's some bug here if from bits is 6 or 7 and tobits is 8.
			// but we don't much care because it's only 8to5 or 5to8 with bypass codes.
			lastByte := ret[len(ret)-1]
			lastMasked := ((lastByte << bits) & ^(byte(0xff) << (frombits - bits)) & remainderMask)
			ret[len(ret)-1] = (lastByte & remainderMask) | lastMasked | remainderByte
		} else {
			ret = append(ret, remainderByte)
		}
	}

	return ret, nil
}
