package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportPEM(t *testing.T) {
	priv, pub, err := generateKeyPair(2048)
	require.NoError(t, err, "failed to generate key pair")

	pubPEM := exportPubKeyAsPEMStr(pub)
	require.NotEmpty(t, pubPEM, "public key PEM should not be empty")

	privPEM := exportPrivKeyAsPEMStr(priv)
	require.NotEmpty(t, privPEM, "private key PEM should not be empty")

	// Decode and parse public key
	block, _ := pem.Decode([]byte(pubPEM))
	require.NotNil(t, block, "failed to decode public key PEM")
	assert.Equal(t, "RSA PUBLIC KEY", block.Type, "unexpected PEM block type for public key")

	_, err = x509.ParsePKCS1PublicKey(block.Bytes)
	require.NoError(t, err, "failed to parse public key")

	// Decode and parse private key
	block, _ = pem.Decode([]byte(privPEM))
	require.NotNil(t, block, "failed to decode private key PEM")
	assert.Equal(t, "RSA PRIVATE KEY", block.Type, "unexpected PEM block type for private key")

	_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	require.NoError(t, err, "failed to parse private key")
}

func TestWritePEMFile(t *testing.T) {
	tmpFile := t.TempDir() + "/key.pem"
	content := []byte("test content")

	err := writePEMFile(tmpFile, content, 0o600)
	require.NoError(t, err, "writePEMFile failed")

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err, "failed to read temp file")
	assert.Equal(t, content, data, "file content mismatch")
}

func TestGenerateKeyPair_RoundTripAndProperties(t *testing.T) {
	tests := []int{1024, 2048}
	for _, bits := range tests {
		t.Run((func() string {
			return "bits_" + strconv.Itoa(bits)
		})(), func(t *testing.T) {
			priv, pub, err := generateKeyPair(bits)
			require.NoError(t, err, "failed to generate key pair")

			// basic sanity checks
			require.NotNil(t, priv, "private key must not be nil")
			require.NotNil(t, pub, "public key must not be nil")
			require.NoError(t, priv.Validate(), "private key should validate")

			// modulus size matches requested bits (RSA modulus is 'bits' long)
			assert.Equal(t, bits, priv.N.BitLen(), "private key modulus size")
			assert.Equal(t, priv.PublicKey.N.BitLen(), pub.N.BitLen(), "public key modulus size")

			// the returned public key corresponds to the private key
			assert.Equal(t, priv.PublicKey.N, pub.N)
			assert.Equal(t, priv.PublicKey.E, pub.E)

			// round-trip: encrypt with pub, decrypt with priv
			msg := []byte("hello, rsa-oaep")
			h := sha256.New()

			ciphertext, err := rsa.EncryptOAEP(h, rand.Reader, pub, msg, nil)
			require.NoError(t, err, "encrypt should succeed")

			plaintext, err := rsa.DecryptOAEP(h, rand.Reader, priv, ciphertext, nil)
			require.NoError(t, err, "decrypt should succeed")
			assert.Equal(t, msg, plaintext, "round-trip should preserve message")
		})
	}
}
