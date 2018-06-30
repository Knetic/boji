package boji

import (
	"os"
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
	// try path-as-dir first
	archive := filepath.Join(path, "archive.zip")
	zreader, err := zip.OpenReader(archive)
	if err == nil {
		return &archivableDir {
			path: path,
			zreader: zreader,
		}, nil
	}

	// not looking for a dir, see if this is an archived dir with the file
	archive = filepath.Join(dir, "archive.zip")
	zreader, err = zip.OpenReader(archive)
	if err == nil {

		// writing something?
		if flag & os.O_CREATE != 0 || flag & os.O_RDWR != 0 || flag & os.O_WRONLY != 0 {
			return newArchiveFileW(archive, filename, zreader)
		}

		// reading existing file?
		for _, zfile := range zreader.File {
			if filepath.Base(zfile.Name) == filename {
				return newArchiveFile(dir, zfile), nil
			}
		}
	}

	// not found, try it straight
	return webdav.Dir(this).OpenFile(ctx, name, flag, perm)
}

func (this archivableFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	
	f, err := this.OpenFile(ctx, name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Stat()
}

func (this archivableFS) Rename(ctx context.Context, oldName, newName string) error {

	// TODO: archive implementation
	return webdav.Dir(this).Rename(ctx, oldName, newName)
}

func (this archivableFS) RemoveAll(ctx context.Context, name string) error {

	path := this.resolve(name)
	if path == "" {
		return errors.New("Unable to resolve local file")
	}

	filename := filepath.Base(path)
	dir := filepath.Dir(path)
	archive := filepath.Join(dir, "archive.zip")
	
	zreader, err := zip.OpenReader(archive)
	if err != nil {
		return webdav.Dir(this).RemoveAll(ctx, name)
	}

	_, err = rewriteArchive(zreader, archive, "", "", filename)	
	return err
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
