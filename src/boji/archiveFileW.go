package boji

import (
	"archive/zip"
	"os"
	"io"
	"fmt"
)

/*
	Represents a _writeable_ file inside a zip archive.
	Once written and closed, this will rewrite the archive, containing the changes to this file.
	Requires locking the entire directory, since archive writes are, by nature, fairly synchronous.
*/
type archiveFileW struct {

	zreader *zip.ReadCloser
	filename string
	archivePath string

	tempfile *os.File
	tempfilePath string

	stat os.FileInfo
	seekPos int64
}

func newArchiveFileW(archivePath string, filename string, zreader *zip.ReadCloser) (*archiveFileW, error) {
	
	tempfilePath := filename
	f, err := os.Create(tempfilePath)
	if err != nil {
		return nil, err
	}

	return &archiveFileW {
		zreader: zreader,
		tempfile: f,
		archivePath: archivePath,
		filename: filename,
		tempfilePath: tempfilePath,
	}, nil
}

func (this *archiveFileW) Seek(offset int64, whence int) (n int64, err error) {
	
	switch whence {
	case os.SEEK_END:
		fallthrough // a bug i know, but this should never really be a thing
	case os.SEEK_SET: this.seekPos = offset
	case os.SEEK_CUR: this.seekPos += offset
	}

	return this.seekPos, nil	
}

func (this *archiveFileW) Write(p []byte) (n int, err error) {
	return this.tempfile.Write(p)
}

func (this *archiveFileW) Close() error {
	
	defer os.Remove(this.tempfilePath)

	// wrap up the writing
	this.tempfile.Close()

	// rewrite the zip archive, adding in the temp file
	tempArchivePath := this.archivePath + "~"
	newArchive, err := os.Create(tempArchivePath)
	if err != nil {
		fmt.Printf("cant make new archive\n")
		return err
	}
	defer os.Remove(this.tempfilePath)

	zwriter := zip.NewWriter(newArchive)
	defer zwriter.Close()

	// copy each extant file (except the old version of the file we're writing)
	for _, zipped := range this.zreader.File {

		if zipped.Name == this.filename {
			continue
		}

		zippedReader, err := zipped.Open()
		if err != nil {
			fmt.Printf("cant open %s\n", zipped.Name)
			return err
		}

		err = compressFile(zipped.FileInfo(), zwriter, zippedReader)
		zippedReader.Close()
		if err != nil {
			fmt.Printf("cant compress %s: %v\n", zipped.Name, err)
			return err
		}
	}

	// add in the new one
	newFile, err := os.Open(this.tempfilePath)
	if err != nil {
		fmt.Printf("cant open up the new file: %v\n", err)
		return err
	}
	defer newFile.Close()

	stat, err := newFile.Stat()
	if err != nil {
		fmt.Printf("cant stat new file: %v\n", err)
		return err
	}

	err = compressFile(stat, zwriter, newFile)
	if err != nil {
		return err
	}

	// replace old with new
	err = os.Rename(tempArchivePath, this.archivePath)
	if err != nil {
		return err
	}

	this.stat = stat
	return nil
}

func (this *archiveFileW) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}
func (this *archiveFileW) Stat() (os.FileInfo, error) {
	
	// some implementations (such as gnome/Mint's dav:// file reader)
	// will stat the same file immediately after writing. 
	if this.stat != nil {
		return this.stat, nil
	}
	return this.tempfile.Stat()
}

func (this *archiveFileW) Read(p []byte) (n int, err error) {
	return 0, nil
}

func compressFile(stat os.FileInfo, writer *zip.Writer, reader io.ReadCloser) error {

	header, err := zip.FileInfoHeader(stat)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate

	compressWriter, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(compressWriter, reader)
	return err
}