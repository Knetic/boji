package boji

import (
	"os"
	"context"
	"strings"
	"path"
	"path/filepath"
	"archive/zip"
	"errors"
	"io"
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

	var fromPath string

	oldPath := this.resolve(oldName)
	newPath := this.resolve(newName)

	zreaderFrom, archiveFrom, err := this.archiveAt(oldName)
	if err != nil {
		return err
	}

	// if it's renaming (not moving) within the same archive dir, just rewrite and short-circuit.
	oldDir := filepath.Dir(oldPath)
	newDir := filepath.Dir(newPath)
	if oldDir == newDir && zreaderFrom != nil {

		oldFilename := filepath.Base(oldPath)
		newFilename := filepath.Base(newPath)
		_, err = rewriteArchive(zreaderFrom, archiveFrom, oldFilename, newFilename, "")
		return err
	}

	zreaderTo, archiveTo, err := this.archiveAt(newName)
	if err != nil {
		return err
	}

	// is it also coming from an archive?
	if zreaderFrom != nil {
		
		// extract first
		fromFilename := filepath.Base(oldPath)
		fromPath = filepath.Join(filepath.Dir(oldPath), fromFilename)
		
		err = extractFile(zreaderFrom, fromFilename, fromPath)
		if err != nil {
			return err
		}
		defer os.Remove(fromPath)

		// at the end of this, delete from the old archive.
		defer func(){
			if err == nil {
				rewriteArchive(zreaderFrom, archiveFrom, "", "", fromFilename)
			}
		}()
	} else {
		fromPath = this.resolve(oldPath)
	}

	// move file to the archive dir (sibling to the actual archive)
	toFilename := filepath.Base(newName)
	toPath := filepath.Join(filepath.Dir(newPath), toFilename)
	
	err = os.Rename(fromPath, toPath)
	if err != nil {
		return err
	}

	// do we need to rewrite the target?
	if zreaderTo != nil {

		// delete it from the sibling
		defer os.Remove(toPath)

		// rewrite target archive with the new file
		_, err = rewriteArchive(zreaderTo, archiveTo, toFilename, "", "")
		return err
	}

	// 

	// it's not archived, just do it standard
	return webdav.Dir(this).Rename(ctx, oldName, newName)
}

func (this archivableFS) archiveAt(name string) (*zip.ReadCloser, string, error) {

	path := this.resolve(name)
	if path == "" {
		return nil, "", errors.New("Unable to resolve local file")
	}

	dir := filepath.Dir(path)
	archive := filepath.Join(dir, "archive.zip")
	
	zreader, err := zip.OpenReader(archive)
	if err != nil {
		return nil, "", err
	}

	return zreader, archive, nil
}

func extractFile(zreader *zip.ReadCloser, filename, path string) error {

	extractedFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer extractedFile.Close()

	for _, child := range zreader.File {
		if child.Name == filename {

			childReader, err := child.Open()
			if err != nil {
				return err
			}

			_, err = io.Copy(extractedFile, childReader)
			return err
		}
	}

	return errors.New("file not found to extract")
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
