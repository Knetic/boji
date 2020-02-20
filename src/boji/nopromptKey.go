package boji

import (
	"errors"
	"golang.org/x/crypto/openpgp"
)

// Represents a known symmetric key
type nopromptKey struct {
	key []byte
	prompted bool
}

func newNoPromptKey(key []byte) *nopromptKey {
	return &nopromptKey {
		key: key,
	}
}

func (this *nopromptKey) prompt(keys []openpgp.Key, symmetric bool) ([]byte, error) {
	
	if !symmetric {
		return []byte{}, errors.New("Cannot decrypt, was not prompted for a symmetric key")
	}
	if this.prompted {
		return []byte{}, errors.New("Key given was incorrect")
	}

	this.prompted = true
	return this.key, nil
}