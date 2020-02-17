package boji

import (
	"io"
	"fmt"
	"os"
	"errors"
	"golang.org/x/crypto/openpgp"
)

// Represents an encrypted file that is decrypted as it is read
type encryptedFile struct {
	File *os.File
	encryptedReader io.Reader
}

func newEncryptedFile(fd *os.File, key []byte) (*encryptedFile, error) {

	fmt.Printf("read 1\n")
	message, err := openpgp.ReadMessage(fd, emptyKeyring(0), nopromptKey(key).prompt, nil)
	if err != nil {
		return nil, err
	}

	if !message.IsSymmetricallyEncrypted {
		return nil, errors.New("File is encrypted, but not symmetrically")
	}

	return &encryptedFile {
		File: fd,
		encryptedReader: message.UnverifiedBody,
	}, nil
}

func (this *encryptedFile) Read(p []byte) (n int, err error) {
	fmt.Printf("read 2\n")
	/ TODO: hangs here. intentional syntax error.
	a, b := this.encryptedReader.Read(p)
	fmt.Printf("read 3\n")
	return a,b
}

func (this *encryptedFile) Seek(offset int64, whence int) (n int64, err error) {
	
	// TODO: steal psuedo-seeking from archiveFile
	fmt.Printf("seek 1\n")
	return 0, nil
}

func (this *encryptedFile) Stat() (os.FileInfo, error) {
	return this.File.Stat()
}

//

func (this *encryptedFile) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

func (this *encryptedFile) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (this *encryptedFile) Close() error {
	
	if this.File == nil {
		return nil
	}
	return this.File.Close()
}

