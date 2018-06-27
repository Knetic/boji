package boji

import (
	"archive/zip"
	"os"
	"io"
	"io/ioutil"
)

/*
	Represents a file contained inside a zip archive, 
	but which should transparently be used as a regular file as far as dav is concerned.
*/
type archiveFile struct {
	path string
	zfile *zip.File
	zreader io.ReadCloser
	parent archivableFS
}

func newArchiveFile(parent archivableFS, path string, zfile *zip.File) *archiveFile {
	return &archiveFile {
		parent: parent,
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

/*
	Extracts this file to a temporary directory and reads it (unless it has already been extracted)
*/
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
	return 0, nil	
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

func (this archiveFile) extractTempFile(zfile *zip.File) (string, error) {

	zreader, err := zfile.Open()
	if err !=nil {
		return "", err
	}

	tmpPath, err := ioutil.TempDir("/tmp/boji/", "b2")
	if err != nil {
		return "", err
	}

	outFile, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zfile.Mode())
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, zreader)
	return tmpPath, err
}