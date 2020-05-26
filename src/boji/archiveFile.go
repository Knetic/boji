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
	stats *telemetryStats

	seekPos int64
}

func newArchiveFile(path string, zfile *zip.File, stats *telemetryStats) *archiveFile {
	return &archiveFile {
		zfile: zfile,
		path: path,
		stats: stats,
	}
}

//
func (this *archiveFile) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

func (this *archiveFile) Stat() (os.FileInfo, error) {
	return this.zfile.FileInfo(), nil
}

func (this *archiveFile) Read(p []byte) (int, error) {

	var err error

	if this.zreader == nil {
		this.zreader, err = this.zfile.Open()
		if err != nil {
			return 0, err
		}
	}
	
	n, err := this.zreader.Read(p)
	this.stats.bytesRead += int64(n)
	return n, err
}

func (this *archiveFile) Seek(offset int64, whence int) (n int64, err error) {
	
	switch whence {
	case os.SEEK_SET: this.seekPos = offset
		
		// reset the zip reader
		if this.zreader != nil {
			this.zreader.Close()
		}

		this.zreader, err = this.zfile.Open()
		if err != nil {
			return 0, err
		}
		
		io.CopyN(ioutil.Discard, this.zreader, offset)

	case os.SEEK_CUR: this.seekPos += offset
		if this.zreader != nil {
			io.CopyN(ioutil.Discard, this.zreader, offset)
		}

	case os.SEEK_END: 
		stat, err := this.Stat()
		if err != nil {
			return -1, err
		}
		this.seekPos = stat.Size() + offset
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