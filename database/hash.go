package database

import (
	"bytes"
	"encoding/hex"
)

type Hash [32]byte

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h *Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)
	return err
}

func (h *Hash) Clone() Hash {
	var hash Hash
	copy(hash[:], h[:])
	return hash
}

func (h Hash) IsEmpty() bool {
	emptyHash := Hash{}

	return bytes.Equal(emptyHash[:], h[:])
}