package boji

import (
	"io"
	"os"
	"io/ioutil"
	"errors"
	"golang.org/x/crypto/openpgp"
)

// Represents an encrypted file that is decrypted as it is read
type encryptedFile struct {
	File *os.File
	
	path string
	key []byte
	
	encryptedReader io.Reader
	seekPos int64
	plaintextSize int64 // optimization. Only set after getPlaintextSize() is called.
	stats *telemetryStats

	flag int
	perm os.FileMode
}

func newEncryptedFile(path string, key []byte, flag int, perm os.FileMode, stats *telemetryStats) (*encryptedFile, error) {
	
	ret := &encryptedFile {
		path: path,
		key: key,
		flag: flag,
		perm: perm,
		stats: stats,
	}

	err := ret.open(false)
	return ret, err
}

func (this *encryptedFile) Read(p []byte) (n int, err error) {
	
	if this.encryptedReader == nil {
		err := this.open(true)
		if err != nil {
			return 0, err
		}

		// if there's no reader, but no error, empty file.
		if this.encryptedReader == nil {
			return 0, io.EOF
		}
	}

	read, err := this.encryptedReader.Read(p)
	if err == nil {
		this.seekPos += int64(read)
	}
	this.stats.bytesRead += int64(read)
	return read, err
}

func (this *encryptedFile) Seek(offset int64, whence int) (n int64, err error) {
	
	if this.encryptedReader == nil {
		err := this.open(true)
		if err != nil {
			return 0, err
		}
	}

	switch whence {
		case os.SEEK_SET:

			err := this.open(true)
			if err != nil {
				return -1, err
			}
			// if empty
			if this.encryptedReader == nil {
				return offset, nil
			}
			
			io.CopyN(ioutil.Discard, this.encryptedReader, offset)
			this.seekPos = offset
	
		case os.SEEK_CUR: 
			
			if this.encryptedReader != nil {
				io.CopyN(ioutil.Discard, this.encryptedReader, offset)
				this.seekPos += offset
			}
	
		case os.SEEK_END:

			// if empty
			if this.encryptedReader == nil {
				return offset, nil
			}

			totalSize, reopened, err := this.getPlaintextSize()
			if err != nil {
				return -1, err
			}

			if !reopened {
				err = this.open(true)
				if err != nil {
					return -1, err
				}
			}

			io.CopyN(ioutil.Discard, this.encryptedReader, totalSize + offset)
			this.seekPos = totalSize + offset
		}
	
		return this.seekPos, nil
}

func (this *encryptedFile) Stat() (os.FileInfo, error) {

	var size int64

	this.stats.filesStatted++

	// file stat isn't good enough, size the pgp headers (and block padding) inflate size.
	// so we have to _read the whole damn file_ to get full size, then return a revised fileinfo
	stat, err := this.File.Stat()
	if err != nil {
		return stat, err
	}

	// don't decrypt the entire file just to stat.
	if this.plaintextSize > 0 || this.encryptedReader != nil {
		size, _, err = this.getPlaintextSize()
		if err != nil {
			return stat, err
		}
	}else {
		size = stat.Size()
	}

	trimmed, _ := hideEncryptionExtension(stat.Name())

	return overrideFileInfo {
		FixedName: trimmed,
		FixedSize: size,
		wrapped: stat,
	}, nil
}

func (this *encryptedFile) Close() error {
	
	if this.File == nil {
		return nil
	}
	return this.File.Close()
}

//

func (this *encryptedFile) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

func (this *encryptedFile) Write(p []byte) (n int, err error) {
	return 0, errors.New("writing not supported on read-only encrypted file")
}

// 

func (this *encryptedFile) open(makeReader bool) error {

	if this.File != nil {
		this.File.Close()
	}

	fd, err := os.OpenFile(this.path, this.flag, this.perm)
	if err != nil {
		return err
	}
	this.File = fd

	// if we aren't requested to decrypt, don't.
	// used so that we don't need to decrypt every file in a dir just to stat them. 
	if !makeReader {
		return nil
	}

	// if we're supposed to decrypt, check if there's any data at all.
	stat, err := fd.Stat()
	if err != nil {
		return err
	}

	if stat.Size() <= 0 {
		this.plaintextSize = 0
		return nil
	}

	message, err := openpgp.ReadMessage(fd, defaultEmptyKeyring, newNoPromptKey(this.key).prompt, nil)
	if err != nil {
		return err
	}

	if !message.IsEncrypted {
		return errors.New("File is not encrypted, but has pgp extension")
	}
	if !message.IsSymmetricallyEncrypted {
		return errors.New("File is encrypted, but not symmetrically")
	}

	this.encryptedReader = message.UnverifiedBody
	this.seekPos = 0
	return nil
}

// return the plaintext size, whether or not the file had to be reopened to determine that size, and any errors.
func (this *encryptedFile) getPlaintextSize() (int64, bool, error) {
	
	// if we have a "cached" size, just use that.
	if this.plaintextSize > 0 {
		return this.plaintextSize, false, nil
	}

	if this.encryptedReader == nil {
		return -1, false, errors.New("no encrypted reader initialized")
	}

	written, err := io.Copy(ioutil.Discard, this.encryptedReader)
	if err != nil {
		return -1, false, err
	}

	totalSize := written + this.seekPos
	
	// reopen, since we've now changed the reader
	err = this.open(true)
	if err != nil {
		return -1, false, err
	}

	this.plaintextSize = totalSize
	return totalSize, true, nil
}
