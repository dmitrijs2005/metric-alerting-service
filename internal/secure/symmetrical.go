package secure

import (
	"crypto/hmac"
	"crypto/sha256"
)

// CreateAes256Signature generates an HMAC-SHA256 signature for the given body
// using the provided key. Despite the name, this function does not perform AES
// encryption — the "AES256" in the name refers to the common practice of using
// a 256-bit (32-byte) key for HMAC signing.
//
// The returned byte slice is the raw 32-byte HMAC digest.
//
// Parameters:
//   - body: The message to sign.
//   - key:  The secret key used to generate the HMAC.
//
// Returns:
//   - The generated HMAC digest as a byte slice.
//   - An error if writing to the hash fails (rare).
func CreateAes256Signature(body []byte, key string) ([]byte, error) {
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(body)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
