package boji

import (
	"os"
	"io"
	"archive/zip"
)

/*
	A compressed directory.
*/
type archivableDir struct {
	path string
	zreader *zip.ReadCloser
	filesRead int

	stats *telemetryStats
}

func (this *archivableDir) Stat() (os.FileInfo, error) {
	
	this.stats.filesStatted++
	
	file, err := os.Open(this.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return file.Stat()
}

func (this *archivableDir) Readdir(count int) ([]os.FileInfo, error) {
	
	var children []os.FileInfo

	// add subdirectories
	file, err := os.Open(this.path)
	if err != nil {
		return []os.FileInfo{}, err
	}
	defer file.Close()

	// if this is the first call, add any subdirectories
	if this.filesRead == 0 {
		files, err := file.Readdir(0)
		if err != nil {
			return children, err
		}

		for _, child := range files {
			if child.IsDir() && len(children) < count {
				children = append(children, child)
			}
		}
	}

	// add files in archive
	for i, child := range this.zreader.File {

		// skip until we get to the page we're after
		if i < this.filesRead {
			continue
		}

		// stop if the page is full
		if count != 0 && len(children) >= count {
			break
		}

		children = append(children, child.FileInfo())
		this.filesRead++		
	}

	if count > 0 && len(this.zreader.File) == this.filesRead {
		err = io.EOF
	} else {
		err = nil
	}

	return children, err
}

func (this *archivableDir) Close() error {
	return nil
}
func (this *archivableDir) Read(p []byte) (n int, err error) {
	return 0, nil
}
func (this *archivableDir) Seek(offset int64, whence int) (n int64, err error) {
	return 0, nil	
}
func (this *archivableDir) Write(p []byte) (n int, err error) {
	return 0, nil
}