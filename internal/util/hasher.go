package util

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
)

type GobHasher struct {
	encBuf bytes.Buffer
	decBuf bytes.Buffer
}

func NewGobHasher() *GobHasher {
	return &GobHasher{}
}

// NOT THREAD SAFE since we use same buffer
func (h *GobHasher) Encode(input []string) (string, error) {
	// Reset the buffer for reuse
	h.encBuf.Reset()
	enc := gob.NewEncoder(&h.encBuf)

	if err := enc.Encode(input); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(h.encBuf.Bytes()), nil
}

func (h *GobHasher) Decode(encoded string) ([]string, error) {
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	// Reset the buffer and fill it with decoded data
	h.decBuf.Reset()
	h.decBuf.Write(decoded)
	dec := gob.NewDecoder(&h.decBuf)

	var result []string
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}

	if result == nil {
		return []string{}, nil
	}

	return result, nil
}
