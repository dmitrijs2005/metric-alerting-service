package common

import (
	"crypto/hmac"
	"crypto/sha256"
)

func CreateAes256Signature(body []byte, key string) ([]byte, error) {
	h := hmac.New(sha256.New, []byte(key))
	_, err := h.Write(body)
	if err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
