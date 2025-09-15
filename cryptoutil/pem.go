package cryptoutil

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// PEMRSAPrivateKey returns key as a PEM block.
func PEMRSAPrivateKey(key *rsa.PrivateKey) []byte {
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(block)
}

// RSAKeyFromPEM decodes a PEM RSA private key.
func RSAKeyFromPEM(key []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(key)
	if block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("PEM type is not RSA PRIVATE KEY")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// PEMCertificate returns derBytes encoded as a PEM block.
func PEMCertificate(derBytes []byte) []byte {
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}
	return pem.EncodeToMemory(block)
}

// CertificateFromPEM decodes a PEM certificate.
func CertificateFromPEM(cert []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(cert)
	if block.Type != "CERTIFICATE" {
		return nil, errors.New("PEM type is not CERTIFICATE")
	}
	return x509.ParseCertificate(block.Bytes)
}
