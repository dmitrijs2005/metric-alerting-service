package main

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportPEM(t *testing.T) {
	priv, pub := generateKeyPair(2048)

	pubPEM := exportPubKeyAsPEMStr(pub)
	require.NotEmpty(t, pubPEM, "public key PEM should not be empty")

	privPEM := exportPrivKeyAsPEMStr(priv)
	require.NotEmpty(t, privPEM, "private key PEM should not be empty")

	// Decode and parse public key
	block, _ := pem.Decode([]byte(pubPEM))
	require.NotNil(t, block, "failed to decode public key PEM")
	assert.Equal(t, "RSA PUBLIC KEY", block.Type, "unexpected PEM block type for public key")

	_, err := x509.ParsePKCS1PublicKey(block.Bytes)
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
