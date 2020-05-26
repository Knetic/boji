package boji

import (
	"archive/zip"
	"os"
	"io"
	"path/filepath"
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

	stats *telemetryStats
}

func newArchiveFileW(archivePath string, filename string, zreader *zip.ReadCloser, stats *telemetryStats) (*archiveFileW, error) {
	
	archiveDir := filepath.Dir(archivePath)
	tempfilePath := filepath.Join(archiveDir, filename)
	
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
		stats: stats,
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

func (this *archiveFileW) Write(p []byte) (int, error) {
	n, err := this.tempfile.Write(p)
	this.stats.bytesWritten += int64(n)
	return n, err
}

func (this *archiveFileW) Close() error {
	
	defer os.Remove(this.tempfilePath)
	this.tempfile.Close()
	
	stat, err := rewriteArchive(this.zreader, this.archivePath, this.filename, "", "")
	this.stat = stat
	return err
}

func (this *archiveFileW) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}
func (this *archiveFileW) Stat() (os.FileInfo, error) {
	
	this.stats.filesStatted++
	
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

/*
	Rewrites the archive.
	If `replaceFile` alone is specified, the given filename (in the same dir) will be added (or updated in) to the archive from a file in the same dir.
	If `renameWith` is also specified, the `replaceFile` will be kept the same as it currently exists in the archive, just with a new name.
	If neither are specified, nothing happens.
	If `deleteFrom` is specified, the given file will be ommitted during rewrites.
*/
func rewriteArchive(zreader *zip.ReadCloser, archivePath string, replaceFile, renameWith, deleteFrom string) (os.FileInfo, error) {

	var stat os.FileInfo

	// rewrite the zip archive, adding in the temp file
	archiveDir := filepath.Dir(archivePath)
	tempArchivePath := archivePath + "~"
	newArchive, err := os.Create(tempArchivePath)
	if err != nil {
		return nil, err
	}

	zwriter := zip.NewWriter(newArchive)
	defer zwriter.Close()

	// copy each extant file (except the old version of the file we're writing)
	for _, zipped := range zreader.File {

		name := zipped.Name

		if zipped.Name == deleteFrom {
			continue
		}
		if zipped.Name == replaceFile {
			if renameWith == "" {
				continue
			}
			name = renameWith
		}

		zippedReader, err := zipped.Open()
		if err != nil {
			return nil, err
		}

		err = compressFile(zipped.FileInfo(), name, zwriter, zippedReader)
		zippedReader.Close()
		if err != nil {
			return nil, err
		}
	}

	// add in the new one
	if replaceFile != "" && renameWith == "" {
		
		newFile, err := os.Open(filepath.Join(archiveDir, replaceFile))
		if err != nil {
			return nil, err
		}
		defer newFile.Close()

		stat, err = newFile.Stat()
		if err != nil {
			return nil, err
		}

		err = compressFile(stat, stat.Name(), zwriter, newFile)
		if err != nil {
			return nil, err
		}
	}

	// replace old with new
	err = os.Rename(tempArchivePath, archivePath)
	if err != nil {
		return nil, err
	}

	return stat, nil
}

func compressFile(stat os.FileInfo, name string, writer *zip.Writer, reader io.ReadCloser) error {

	header, err := zip.FileInfoHeader(stat)
	if err != nil {
		return err
	}
	header.Method = zip.Deflate
	header.Name = name

	compressWriter, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(compressWriter, reader)
	return err
}