package boji

import (
	"io"
	"io/ioutil"
	"os"
	"bytes"
	"errors"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

func encryptDir(path string, key string) error {
	return nil
}

func decryptDir(path string, key string) error {
	return nil
}

/*
	Encrypts the given bytes with the given key, storing them at the given path.
*/
func encryptFile(contents []byte, path string, key string) error {

	buf := new(bytes.Buffer)

	// AES-256, no compression (users can already transparently compress)
	packetConfig := &packet.Config {
		DefaultCipher: packet.CipherAES256,
		CompressionConfig: &packet.CompressionConfig {
			Level: 0,
		},
	}

	encryptor, err := openpgp.SymmetricallyEncrypt(buf, []byte(key), nil, packetConfig)
	if err != nil {
		return err
	}

	// encrypt, first to memory (so we don't corrupt any existing data if this fails)
	_, err = encryptor.Write(contents)
	if err != nil {
		return err
	}

	err = encryptor.Close()
	if err != nil {
		return err
	}

	// write
	fd, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = io.Copy(buf, fd)
	if err != nil {
		return err
	}

	return nil
}

/*
	Decrypts the given local file with the given key, returning the contents.
*/
func decryptFile(path string, key string) ([]byte, error) {

	keyer := nopromptKey([]byte(key))

	fd, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer fd.Close()

	message, err := openpgp.ReadMessage(fd, nil, keyer.prompt, nil)
	if err != nil {
		return []byte{}, err
	}
	
	return ioutil.ReadAll(message.UnverifiedBody)
}

// Represents a known symmetric key
type nopromptKey []byte

func (this nopromptKey) prompt(keys []openpgp.Key, symmetric bool) ([]byte, error) {
	if !symmetric {
		return []byte{}, errors.New("Cannot decrypt, was not prompted for a symmetric key")
	}

	return this, nil
}