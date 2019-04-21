package boji

import (
	"io"
	"io/ioutil"
	"os"
	"bytes"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

func encryptDir(path string, key []byte) error {
	return nil
}

func decryptDir(path string, key []byte) error {
	return nil
}

/*
	Encrypts the given bytes with the given key, storing them at the given path.
*/
func encryptFile(contents []byte, path string, key []byte) error {

	buf := new(bytes.Buffer)

	encryptor, err := defaultEncryptor(buf, key)
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
func decryptFile(path string, key []byte) ([]byte, error) {

	keyer := nopromptKey(key)

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

// Returns a writer that will encrypt the contents with AES-256, no compression.
func defaultEncryptor(cipherText io.Writer, key []byte) (io.WriteCloser, error) {

	// AES-256, no compression (users can already transparently compress)
	packetConfig := &packet.Config {
		DefaultCipher: packet.CipherAES256,
		CompressionConfig: &packet.CompressionConfig {
			Level: 0,
		},
	}

	return openpgp.SymmetricallyEncrypt(cipherText, key, nil, packetConfig)
}