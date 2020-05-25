package boji

import (
	"fmt"
	"io"
	"os"
	"errors"
	"golang.org/x/crypto/openpgp"
)

// Represents a file that can be written to with transparent encryption
type encryptedFileW struct {
	Path string

	fd *os.File
	plaintextBytes int64
	encryptedWriter io.WriteCloser

	key []byte
	flag int
	perm os.FileMode
}

func newEncryptedFileW(path string, key []byte, flag int, perm os.FileMode) (*encryptedFileW, error) {

	fmt.Printf("newEncryptedFileW %s\n", path)

	// open temporary file to write to.
	fd, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return nil, err
	}

	return &encryptedFileW {
		Path: path,
		flag: flag,
		perm: perm,
		fd: fd,
		key: key,
	}, nil
}

func (this *encryptedFileW) Write(p []byte) (int, error) {
	
	var err error

	// only open encrypted writer when we have something to write
	if this.encryptedWriter == nil {
		
		this.encryptedWriter, err = openpgp.SymmetricallyEncrypt(this.fd, this.key, nil, defaultPacketConfig())
		
		// null out key once it's used. Never keep it if we can help it.
		this.key = []byte{}
		if err != nil {
			return 0, err
		}
	}

	n, err := this.encryptedWriter.Write(p)
	this.plaintextBytes += int64(n)

	fmt.Printf("encryptedFileW.Write (%d) (%d)\n", len(p), n)
	return n, err
}

func (this *encryptedFileW) Close() error {
	
	fmt.Printf("newEncryptedFileW.Close\n")
	
	// if we haven't written anything, return the fd closure err
	if this.encryptedWriter == nil {
		return this.fd.Close()
	}

	// otherwise, make sure fd closes, but preferentially return encrypted writer closure
	defer this.fd.Close()
	return this.encryptedWriter.Close()
}

func (this *encryptedFileW) Stat() (os.FileInfo, error) {
	
	var info os.FileInfo
	var err error

	if this.fd != nil {
		info, err = this.fd.Stat()
	} else {
		info, err = os.Stat(this.Path)
	}

	if err != nil {
		return nil, err
	}

	fmt.Printf("newEncryptedFileW.Stat (%s), (%d)\n", (info.Name() + encryptedExtension), this.plaintextBytes)

	return overrideFileInfo {
		FixedSize: this.plaintextBytes,
		FixedName: info.Name() + encryptedExtension,
		wrapped: info,
	}, nil
}

func (this *encryptedFileW) Seek(offset int64, whence int) (n int64, err error) {
	fmt.Printf("encryptedFileW.Seek\n")
	return -1, errors.New("Cannot seek an encrypted file")
}
func (this *encryptedFileW) Readdir(count int) ([]os.FileInfo, error) {
	fmt.Printf("encryptedFileW.Readdir\n")
	return []os.FileInfo{}, nil
}
func (this *encryptedFileW) Read(p []byte) (n int, err error) {
	fmt.Printf("encryptedFileW.Read\n")
	return 0, nil
}