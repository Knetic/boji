package boji

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

type singleWalkFunc func(path string, key []byte) error

func singleWalk(path string, key []byte, walkfunc singleWalkFunc) error {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		err = walkfunc(path + "/" + f.Name(), key)
		if err != nil {
			return err
		}
	}
	return nil
}

func encryptDir(path string, key []byte) error {
	return singleWalk(path, key, encryptFile)
}

func decryptDir(path string, key []byte) error {
	return singleWalk(path, key, decryptFile)
}

/*
	Encrypts the given bytes with the given key, storing them at the given path +".pgp"
*/
func encryptFile(path string, key []byte) error {

	if strings.HasSuffix(path, encryptedExtension) {
		return nil
	}

	encryptedPath := path + encryptedExtension
	
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(encryptedPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	plaintext, err := defaultEncryptor(dst, key)
	if err != nil {
		return err
	}
	defer plaintext.Close()

	_, err = io.Copy(plaintext, src)
	if err != nil {
		return err
	}
	
	err = plaintext.Close()
	if err != nil {
		return err
	}

	return os.Remove(path)
}

/*
	Decrypts the given local file with the given key, returning the contents.
	"path" is assumed to include the ".pgp" postfix.
*/
func decryptFile(path string, key []byte) error {

	if !strings.HasSuffix(path, encryptedExtension) {
		return nil
	}

	decryptPath := strings.TrimSuffix(path, encryptedExtension)

	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(decryptPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	message, err := openpgp.ReadMessage(src, defaultEmptyKeyring, nopromptKey(key).prompt, nil)
	if err != nil {
		return err
	}
	
	_, err = io.Copy(dst, message.UnverifiedBody)
	if err != nil {
		return err
	}

	return os.Remove(path)
}

// Returns a writer that will encrypt the contents with AES-256, no compression.
func defaultEncryptor(cipherText io.Writer, key []byte) (io.WriteCloser, error) {
	return openpgp.SymmetricallyEncrypt(cipherText, key, nil, defaultPacketConfig())
}

// AES-256, no compression (users can already transparently compress)
func defaultPacketConfig() *packet.Config {
	return &packet.Config {
		DefaultCipher: packet.CipherAES256,
		CompressionConfig: &packet.CompressionConfig {
			Level: 0,
		},
	}
}