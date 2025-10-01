package albc

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// testConvert checks that both forward-and-back bit (byte) conversion matches.
func testConvert(bc BypassCode) (bool, error) {
	a := bc[:]
	a5, err := convertBits(a, 8, 5)
	if err != nil {
		return false, err
	}
	b, err := convertBits(a5, 5, 8)
	if err != nil {
		return false, err
	}
	if bytes.Equal(a, b) {
		return true, nil
	}
	return false, nil
}

func TestBypassCode(t *testing.T) {
	tests := []struct {
		raw  string
		code string
		hash string
	}{
		{
			raw:  "00000000000000000000000000000000",
			code: "00000-00000-0000-0000-0000-0000",
			hash: "deab860d28deb5b7121d6d8fcf0f78e1471756d1b2c566c03277c23ea8930b4f",
		},
		{
			raw:  "1ea841db5edfafe6075b5ae0d845d254",
			code: "3UM43-PUYVY-QYD1-UVCC-HEHJ-FKA4",
			hash: "6ab40d5eabe7218ec04182f461005600c7e3426bddd82cdb405bde9a1e0014b5",
		},
		{
			raw:  "44ebe63375969fec2da67e87e7317946",
			code: "8LNYD-DVNKU-GYRC-E6GU-3YFD-CT86",
			hash: "c1968cb4c013ea893f1922bb5c39f81e35012c0bd9ce3c01cc2a05873a2499e6",
		},
		{
			raw:  "cb84798c3ca85a674194550a2e96aed8",
			code: "TF27L-31WN1-E6FH-DMAM-52X5-NFV0",
			hash: "23cf8b7873425fd8efe31dc5b6ab9c357eb98a2a59c82ea1084ca8af58cc480a",
		},

		{
			raw:  "89195c9b79178736203bd9d591ea7c0f",
			code: "J4DNT-6VT2Y-3LD8-1VV7-AT3U-LW17",
			hash: "59b9b3fa9ec4b806612b8b1fe6f12fcc3903156a58bcf4cae53a8a78dad563d3",
		},
		{
			raw:  "60110f362c6f7a90dd1ef2845f32482f",
			code: "D08HY-EJDEX-X91Q-8YYA-25YD-K857",
			hash: "4d19162b50dd61536d72c0662dce9d533a1f46137d6db97501ceb171fcbae7dd",
		},
		{
			raw:  "9653b0f9b495d8fab25e728ff041b0f1",
			code: "KT9V1-YEMKQ-DGND-KYFA-7Z0H-EHY1",
			hash: "5893831ab50670e96f5d245a4f597c86eeffd16aac3ee5bb1c2251affb004a33",
		},
		{
			raw:  "0de305a24090fc54b61ed7e9e39569fb",
			code: "1QJHC-8K0K3-Y59E-HYUZ-MY75-C9Z3",
			hash: "1db2f16ad21a135b9c2523725e836c0b0528fec83c195c6ecdf8761fd877889a",
		},
		{
			raw:  "f398ef9199e9f0aefea0e782ab8b61a9",
			code: "YFDFZ-4DTX7-RAXZ-N0WY-1AQ2-V1N1",
			hash: "7c7715d092a5cfcd16a6037555e11e4fa53edda9cda4d8464d58ef39ba9b5b0f",
		},
		{
			raw:  "bd9e6a463a19ac706d379394bf97747a",
			code: "QPG6M-JJU36-P70V-9QKF-ACZ5-VMG2",
			hash: "7cdf895759d090eb9d3ed833d0ed7d5d5b00a11a719293f44aa7741ffbe79f6a",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()

			// test stores it as a hex string for convenience
			b, err := hex.DecodeString(tt.raw)
			if err != nil {
				t.Fatal(err)
			}

			bc, err := NewFromBytes(b)
			if err != nil {
				t.Error(err)
			}

			// converts bits back and forth
			ok, err := testConvert(bc)
			if err != nil {
				t.Error(err)
			} else if !ok {
				t.Error("bit conversion: not equal")
			}

			// check the PBKDF2 hash
			hash, err := bc.Hash()
			if err != nil {
				t.Error(err)
			}
			if have, want := hash, tt.hash; have != want {
				t.Errorf("hash: have %q, want %q", have, want)
			}

			// check the dashed "human readable" form
			code, err := bc.Code()
			if err != nil {
				t.Error(err)
			}
			if have, want := code, tt.code; have != want {
				t.Errorf("code: have %q, want %q", have, want)
			}

			// re-decode the code string
			bc2, err := NewFromCode(code)
			if err != nil {
				t.Error(err)
			}

			// make sure this newly decoded string matches
			if !bytes.Equal(bc[:], bc2[:]) {
				t.Error("decoded bypass codes not equal")
			}
		})
	}

}
