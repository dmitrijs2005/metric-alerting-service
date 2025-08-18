package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// exportPubKeyAsPEMStr returns the given RSA public key encoded as a
// PEM-formatted string using PKCS#1 bytes with the type "RSA PUBLIC KEY".
func exportPubKeyAsPEMStr(pubkey *rsa.PublicKey) string {
	pubKeyPem := string(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pubkey),
		},
	))
	return pubKeyPem
}

// exportPrivKeyAsPEMStr returns the given RSA private key encoded as a
// PEM-formatted string using PKCS#1 bytes with the type "RSA PRIVATE KEY".
func exportPrivKeyAsPEMStr(privkey *rsa.PrivateKey) string {
	privKeyPem := string(pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privkey),
		},
	))
	return privKeyPem

}

// generateKeyPair creates a new RSA key pair with the requested key size in bits.
// It returns the private key and the corresponding public key. If key generation
// fails, the error is printed to stdout and the returned keys may be nil.
func generateKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	return privateKey, &privateKey.PublicKey, nil
}

// writePEMFile writes raw bytes to a file named fn with the provided permissions.
// It returns any error encountered while writing.
func writePEMFile(fn string, b []byte, perm uint32) error {
	return os.WriteFile(fn, b, os.FileMode(perm))
}
