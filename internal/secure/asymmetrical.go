package secure

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
)

// oaepLabel is an optional label used with RSA-OAEP encryption and decryption.
// It must be the same for both encryption and decryption.
var oaepLabel = []byte("my-oaep-label")

// MaxPlainOAEP returns the maximum plaintext size in bytes that can be
// encrypted in a single RSA-OAEP(SHA-256) operation for the given public key.
func MaxPlainOAEP(pub *rsa.PublicKey) int {
	k := pub.Size()     // bytes
	hLen := sha256.Size // 32
	return k - 2*hLen - 2
}

// LoadRSAPublicKeyFromPEM reads and parses an RSA public key from a PEM file
// containing a PKCS#1 public key block.
//
// Returns the public key or an error if the file cannot be read, the PEM
// cannot be decoded, or the key cannot be parsed.
func LoadRSAPublicKeyFromPEM(path string) (*rsa.PublicKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("no PEM block")
	}
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pub, nil
}

// LoadRSAPrivateKeyFromPEM reads and parses an RSA private key from a PEM file.
// Supports PKCS#1 and PKCS#8 encoded private keys.
//
// Returns the private key or an error if the file cannot be read, the PEM
// cannot be decoded, the key cannot be parsed, or it is not an RSA key.
func LoadRSAPrivateKeyFromPEM(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("no PEM block")
	}
	if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	any, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	k, ok := any.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not RSA private key")
	}
	return k, nil
}

// EncryptRSAOAEPChunked encrypts a message of arbitrary length using
// RSA-OAEP with SHA-256 by splitting it into chunks small enough to fit
// into individual RSA-OAEP encryption operations.
//
// The result is a base64-encoded string containing the concatenated RSA
// ciphertext blocks.
//
// The same oaepLabel constant must be used for decryption.
func EncryptRSAOAEPChunked(msg []byte, pub *rsa.PublicKey) (string, error) {
	max := MaxPlainOAEP(pub)
	hash := sha256.New()
	var out []byte
	for off := 0; off < len(msg); {
		end := off + max
		if end > len(msg) {
			end = len(msg)
		}
		ct, err := rsa.EncryptOAEP(hash, rand.Reader, pub, msg[off:end], oaepLabel)
		if err != nil {
			return "", err
		}
		out = append(out, ct...)
		off = end
	}
	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptRSAOAEPChunked decrypts data previously produced by EncryptRSAOAEPChunked.
// The input must be a base64-encoded string containing concatenated RSA ciphertext
// blocks, each of size equal to priv.Size().
//
// The same oaepLabel constant must be used for encryption.
func DecryptRSAOAEPChunked(b64 string, priv *rsa.PrivateKey) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	k := priv.Size()
	hash := sha256.New()
	var plain []byte
	for off := 0; off < len(data); {
		if off+k > len(data) {
			return nil, errors.New("truncated RSA block stream")
		}
		block := data[off : off+k]
		pt, err := rsa.DecryptOAEP(hash, rand.Reader, priv, block, oaepLabel)
		if err != nil {
			return nil, err
		}
		plain = append(plain, pt...)
		off += k
	}
	return plain, nil
}
