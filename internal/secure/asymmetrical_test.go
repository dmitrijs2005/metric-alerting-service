package secure

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptRSAOAEPChunked(t *testing.T) {
	// Generate a test key pair (small for speed in tests)
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pub := &priv.PublicKey

	tests := []struct {
		name string
		data []byte
	}{
		{"short message", []byte("hello world")},
		{"exact block size", make([]byte, MaxPlainOAEP(pub))},
		{"multi-block message", make([]byte, MaxPlainOAEP(pub)*3+10)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fill with random data if empty or zero-filled
			if len(tt.data) > 0 && allZero(tt.data) {
				_, err := rand.Read(tt.data)
				require.NoError(t, err)
			}

			enc, err := EncryptRSAOAEPChunked(tt.data, pub)
			require.NoError(t, err, "encryption failed")

			dec, err := DecryptRSAOAEPChunked(enc, priv)
			require.NoError(t, err, "decryption failed")
			require.Equal(t, tt.data, dec, "decrypted data mismatch")
		})
	}
}

func TestDecryptFailsOnCorruptData(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// This is not valid base64
	_, err = DecryptRSAOAEPChunked("!!!notbase64!!!", priv)
	require.Error(t, err, "expected base64 decode error")

	// Valid base64 but wrong size (less than priv.Size())
	badData := "QUJD" // "ABC"
	_, err = DecryptRSAOAEPChunked(badData, priv)
	require.Error(t, err, "expected truncated RSA block stream error")
}

// allZero checks whether a slice contains only zero bytes.
func allZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

// helper: write PEM file
func writePEM(t *testing.T, dir, name, typ string, der []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	buf := &bytes.Buffer{}
	require.NoError(t, pem.Encode(buf, &pem.Block{Type: typ, Bytes: der}))
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o600))
	return path
}

func TestMaxPlainOAEP(t *testing.T) {
	// Max plaintext formula: k - 2*hLen - 2, with hLen = 32 for SHA-256
	// We test for a couple of key sizes and also do a real encrypt at boundary.
	sizes := []int{1024, 2048, 3072}
	for _, bits := range sizes {
		t.Run((func() string { return "bits_" + itoa(bits) })(), func(t *testing.T) {
			priv, err := rsa.GenerateKey(rand.Reader, bits)
			require.NoError(t, err)
			pub := &priv.PublicKey

			k := pub.Size()               // bytes
			want := k - 2*sha256.Size - 2 // formula
			got := MaxPlainOAEP(pub)
			assert.Equal(t, want, got)

			// Boundary behavior: exactly max bytes must encrypt; +1 must fail.
			h := sha256.New()
			okMsg := make([]byte, got)
			_, _ = rand.Read(okMsg)
			_, err = rsa.EncryptOAEP(h, rand.Reader, pub, okMsg, nil)
			assert.NoError(t, err, "encrypting exactly max bytes should succeed")

			failMsg := make([]byte, got+1)
			_, _ = rand.Read(failMsg)
			_, err = rsa.EncryptOAEP(h, rand.Reader, pub, failMsg, nil)
			assert.Error(t, err, "encrypting more than max bytes should fail")
		})
	}
}

func TestLoadRSAPublicKeyFromPEM_OK_and_Errors(t *testing.T) {
	dir := t.TempDir()

	// Generate RSA keypair; write PKCS#1 public key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubDER := x509.MarshalPKCS1PublicKey(&priv.PublicKey)
	pubPath := writePEM(t, dir, "pub_pkcs1.pem", "RSA PUBLIC KEY", pubDER)

	// OK: load valid PKCS#1 public
	pub, err := LoadRSAPublicKeyFromPEM(pubPath)
	require.NoError(t, err)
	assert.Equal(t, priv.PublicKey.N, pub.N)
	assert.Equal(t, priv.PublicKey.E, pub.E)

	// Error: file missing
	_, err = LoadRSAPublicKeyFromPEM(filepath.Join(dir, "missing.pem"))
	assert.Error(t, err)

	// Error: not a PEM file
	notPEM := filepath.Join(dir, "not_pem.txt")
	require.NoError(t, os.WriteFile(notPEM, []byte("hello"), 0o600))
	_, err = LoadRSAPublicKeyFromPEM(notPEM)
	assert.EqualError(t, err, "no PEM block")

	// Error: wrong PEM block (e.g., random bytes)
	badDER := []byte{1, 2, 3, 4}
	badPath := writePEM(t, dir, "bad_pub.pem", "RSA PUBLIC KEY", badDER)
	_, err = LoadRSAPublicKeyFromPEM(badPath)
	assert.Error(t, err)
}

func TestLoadRSAPrivateKeyFromPEM_PKCS1_and_PKCS8(t *testing.T) {
	dir := t.TempDir()

	// Generate RSA private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// PKCS#1
	pkcs1DER := x509.MarshalPKCS1PrivateKey(priv)
	pkcs1Path := writePEM(t, dir, "rsa_pkcs1.pem", "RSA PRIVATE KEY", pkcs1DER)

	// OK: load PKCS#1
	got1, err := LoadRSAPrivateKeyFromPEM(pkcs1Path)
	require.NoError(t, err)
	assert.Equal(t, priv.N, got1.N)
	assert.Equal(t, priv.E, got1.E)
	assert.Equal(t, priv.D, got1.D)

	// PKCS#8
	pkcs8DER, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)
	pkcs8Path := writePEM(t, dir, "rsa_pkcs8.pem", "PRIVATE KEY", pkcs8DER)

	// OK: load PKCS#8
	got2, err := LoadRSAPrivateKeyFromPEM(pkcs8Path)
	require.NoError(t, err)
	assert.Equal(t, priv.N, got2.N)
	assert.Equal(t, priv.E, got2.E)
	assert.Equal(t, priv.D, got2.D)

	// Error: not RSA in PKCS#8 (e.g., ECDSA key)
	ecPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	ecDER, err := x509.MarshalPKCS8PrivateKey(ecPriv)
	require.NoError(t, err)
	ecPath := writePEM(t, dir, "ec_pkcs8.pem", "PRIVATE KEY", ecDER)

	_, err = LoadRSAPrivateKeyFromPEM(ecPath)
	require.Error(t, err)
	assert.Equal(t, "not RSA private key", err.Error())

	// Error: invalid PEM
	notPEM := filepath.Join(dir, "not_pem.txt")
	require.NoError(t, os.WriteFile(notPEM, []byte("oops"), 0o600))
	_, err = LoadRSAPrivateKeyFromPEM(notPEM)
	assert.EqualError(t, err, "no PEM block")
}

// --- small helper: strconv.Itoa without importing strconv (keeps imports lean)
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
