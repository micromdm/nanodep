package api

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMAIDCheckinJWT(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	u := uuid.NewString()
	s, err := newMAIDCheckinJWT(u, privKey)
	if err != nil {
		t.Fatal(err)
	}
	tok, err := jwt.Parse(s, func(_ *jwt.Token) (interface{}, error) { return &privKey.PublicKey, nil })
	if err != nil {
		t.Fatal(err)
	}
	iss, err := tok.Claims.GetIssuer()
	if err != nil {
		t.Fatal(err)
	}
	if have, want := iss, u; have != want {
		t.Errorf("have %v, want %v", have, want)
	}
}
