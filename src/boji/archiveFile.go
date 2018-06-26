package boji
/*
import (
	"archive/zip"
	"os"
	"io"
)

type archiveFile struct {
	zfile *zip.File
	zreader io.ReadCloser
	parent *archiveDir
}

func newArchiveFile(zfile *zip.File) *archiveFile {
	return &archiveFile {
		zfile: zfile,
	}
}

//
func (this *archiveFile) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

func (this *archiveFile) Stat() (os.FileInfo, error) {
	return this.zfile.FileInfo(), nil
}

//
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
	
	if this.zreader == nil {
		this.zreader, err = this.zfile.Open()
		if err != nil {
			return 0, err
		}
	}

	return this.zreader.Seek(offset, whence)
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

*/