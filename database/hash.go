package database

import "encoding/hex"

type Hash [32]byte

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h *Hash) UnmarshalText(data []byte) error {
	_, err :=  hex.Decode(h[:], data)
	return err
}