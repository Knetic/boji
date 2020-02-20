package boji

import (
	"io"
	"os"
	"errors"
	"golang.org/x/crypto/openpgp"
)

// Represents a file that can be written to with transparent encryption
type encryptedFileW struct {
	Path string

	tempPath string
	tempFile *os.File
	encryptedWriter io.WriteCloser

	flag int
	perm os.FileMode
}

func newEncryptedFileW(path string, key []byte, flag int, perm os.FileMode) (*encryptedFileW, error) {

	// open temporary file to write to.
	tempPath := path + ".tmp"
	tempFile, err := os.OpenFile(tempPath, flag, perm)
	if err != nil {
		return nil, err
	}

	encryptedWriter, err := openpgp.SymmetricallyEncrypt(tempFile, key, nil, defaultPacketConfig())
	if err != nil {
		return nil, err
	}

	return &encryptedFileW {
		Path: path,
		flag: flag,
		perm: perm,
		tempPath: tempPath,
		tempFile: tempFile,
		encryptedWriter: encryptedWriter,
	}, nil
}

func (this *encryptedFileW) Write(p []byte) (n int, err error) {
	return this.encryptedWriter.Write(p)
}

func (this *encryptedFileW) Close() error {
	
	err := this.encryptedWriter.Close()
	if err != nil {
		this.tempFile.Close()
		return err
	}

	this.tempFile.Close()
	return os.Rename(this.tempPath, this.Path)
}

func (this *encryptedFileW) Stat() (os.FileInfo, error) {
	
	if this.tempFile != nil {
		return this.tempFile.Stat()
	}

	return os.Stat(this.Path)
}

func (this *encryptedFileW) Seek(offset int64, whence int) (n int64, err error) {
	return -1, errors.New("Cannot seek an encrypted file")
}
func (this *encryptedFileW) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}
func (this *encryptedFileW) Read(p []byte) (n int, err error) {
	return 0, nil
}