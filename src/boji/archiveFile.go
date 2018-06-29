package boji

import (
	"archive/zip"
	"os"
	"io"
)

/*
	Represents a file contained inside a zip archive, 
	but which should transparently be used as a regular file as far as dav is concerned.
*/
type archiveFile struct {
	path string
	zfile *zip.File
	zreader io.ReadCloser

	seekPos int64
}

func newArchiveFile(path string, zfile *zip.File) *archiveFile {
	return &archiveFile {
		zfile: zfile,
		path: path,
	}
}

//
func (this *archiveFile) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

func (this *archiveFile) Stat() (os.FileInfo, error) {
	return this.zfile.FileInfo(), nil
}

func (this *archiveFile) Read(p []byte) (n int, err error) {

	if this.zreader == nil {
		this.zreader, err = this.zfile.Open()
		if err != nil {
			return 0, err
		}
	}
	
	return this.zreader.Read(p)
}

func (this *archiveFile) Seek(offset int64, whence int) (n int64, err error) {
	
	switch whence {
	case os.SEEK_SET: this.seekPos = offset
	case os.SEEK_CUR: this.seekPos += offset
	case os.SEEK_END: 
		stat, err := this.Stat()
		if err != nil {
			return -1, err
		}
		this.seekPos = stat.Size() - offset
	}

	return this.seekPos, nil	
}

func (this *archiveFile) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (this *archiveFile) Close() error {
	
	if this.zreader == nil {
		return nil
	}
	return this.zreader.Close()
}