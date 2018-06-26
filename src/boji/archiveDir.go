package boji

import (
	"os"
	"context"
	"strings"
	"path"
	"path/filepath"
	"archive/zip"
	"io/ioutil"
	"io"
	"golang.org/x/net/webdav"
)

/*
	A transparent-compression webdav filesystem.
	Any folder that only contains one zip archive will be considered compressed. 
	Further subdirectories are not part of that archive.

	Any operations that occur on a compressed directory will happen within that archive.
*/
type archiveDir string

func (this archiveDir) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return webdav.Dir(this).Mkdir(ctx, name, perm)
}

func (this archiveDir) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	
	// try it straight first
	file, derr := webdav.Dir(this).OpenFile(ctx, name, flag, perm)
	return file, derr
	
	if derr != nil && file != nil {
		return file, derr
	}

	// the file is inaccessible, check if this is a compressed dir
	
	//
	path := this.resolve(name)
	if path == "" {
		return file, derr
	}

	filename := filepath.Base(path)
	dir := filepath.Dir(path)
	archive := filepath.Join(dir, "archive.zip")

	zreader, err := zip.OpenReader(archive)
	if err != nil {
		return nil, err
	}

	// find the file, extract it to a temporary directory, and return that file.
	for _, zfile := range zreader.File {
		if zfile.Name == filename {

			extractPath, err := this.extractTempFile(path, zfile)
			if err != nil {
				return nil, err
			}

			return os.Open(extractPath)
		}
	}
	
	// not found, give back what the standard implementation would have given.
	return nil, derr
}

func (this archiveDir) Rename(ctx context.Context, oldName, newName string) error {
	return webdav.Dir(this).Rename(ctx, oldName, newName)
}

func (this archiveDir) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return webdav.Dir(this).Stat(ctx, name) 
}

func (this archiveDir) RemoveAll(ctx context.Context, name string) error {
	return webdav.Dir(this).RemoveAll(ctx, name)
}

func (this archiveDir) extractTempFile(path string, zfile *zip.File) (string, error) {

	zreader, err := zfile.Open()
	if err !=nil {
		return "", err
	}

	tmpPath, err := ioutil.TempDir("/tmp/boji", path)
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

// stolen from the golang.org webdav implementation
func (this archiveDir) resolve(name string) string {
	if filepath.Separator != '/' && strings.IndexRune(name, filepath.Separator) >= 0 ||
		strings.Contains(name, "\x00") {
		return ""
	}
	dir := string(this)
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, filepath.FromSlash(slashClean(name)))
}
func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}