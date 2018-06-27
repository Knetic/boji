package boji

import (
	"os"
	"fmt"
	"context"
	"strings"
	"path"
	"path/filepath"
	"archive/zip"
	"errors"
	"golang.org/x/net/webdav"
)

/*
	A transparent-compression webdav filesystem.
	Any folder that only contains one zip archive will be considered compressed. 
	Further subdirectories are not part of that archive.

	Any operations that occur on a compressed directory will happen within that archive.
*/
type archivableFS string

func (this archivableFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return webdav.Dir(this).Mkdir(ctx, name, perm)
}

func (this archivableFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	
	// first try to see if it's archived
	path := this.resolve(name)
	if path == "" {
		return nil, errors.New("Unable to resolve local file")
	}

	filename := filepath.Base(path)
	dir := filepath.Dir(path)

	// we might either be browsing a directory which needs to be populated,
	// or we might be trying to access a specific file inside the dir.
	archive := filepath.Join(path, "archive.zip")

	zreader, err := zip.OpenReader(archive)
	if err == nil {

		// find the file, extract it to a temporary directory, and return that file.
		for _, zfile := range zreader.File {
			if zfile.Name == filename {
				return newArchiveFile(this, dir, zfile), nil
			}
		}

		// file not found, but this is definitely archived, so return just a list of files and directories.
		return &archivableDir {
			path: path,
			zreader: zreader,
		}, nil
	}

	fmt.Printf("Didn't find archive at %s\n", archive)

	// not found, try it straight
	return webdav.Dir(this).OpenFile(ctx, name, flag, perm)
}

func (this archivableFS) Rename(ctx context.Context, oldName, newName string) error {
	return webdav.Dir(this).Rename(ctx, oldName, newName)
}

func (this archivableFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return webdav.Dir(this).Stat(ctx, name) 
}

func (this archivableFS) RemoveAll(ctx context.Context, name string) error {
	return webdav.Dir(this).RemoveAll(ctx, name)
}

// stolen from the golang.org webdav implementation
func (this archivableFS) resolve(name string) string {
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
