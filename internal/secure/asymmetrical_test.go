package secure

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

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
