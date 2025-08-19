package secure

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAes256Signature(t *testing.T) {
	body := []byte("hello world")
	key := "secret-key"

	gotSig, err := CreateAes256Signature(body, key)
	assert.NoErrorf(t, err, "CreateAes256Signature returned error: %v", err)

	// Compute expected signature directly
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	wantSig := h.Sum(nil)

	assert.Equalf(t, gotSig, wantSig, "signature mismatch:\n got:  %s\n want: %s",
		hex.EncodeToString(gotSig), hex.EncodeToString(wantSig))

	// Ensure signature changes with a different key
	gotSig2, err := CreateAes256Signature(body, "different-key")
	assert.NoErrorf(t, err, "unexpected error with different key: %v", err)

	assert.NotEqual(t, gotSig, gotSig2, "expected different signature for different key, got same value")
}
