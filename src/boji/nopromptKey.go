package boji

import (
	"errors"
	"golang.org/x/crypto/openpgp"
)

// Represents a known symmetric key
type nopromptKey []byte

func (this nopromptKey) prompt(keys []openpgp.Key, symmetric bool) ([]byte, error) {
	if !symmetric {
		return []byte{}, errors.New("Cannot decrypt, was not prompted for a symmetric key")
	}

	return this, nil
}