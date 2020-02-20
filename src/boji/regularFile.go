package boji

import (
	"os"
	"context"
	"strings"
	"golang.org/x/net/webdav"
)

/*
	Wrapped version of default webdav.File implementation.
	Mostly used so that directory reading can show "regular" names of transparently encrypted files.
*/
type regularFile struct {
	wrapped webdav.File
}

func newRegularFile(base string, ctx context.Context, path string, flag int, perm os.FileMode) (*regularFile, error) {
	
	wrapped, err := webdav.Dir(base).OpenFile(ctx, path, flag, perm)
	if err != nil {
		return nil, err
	}

	return &regularFile {
		wrapped: wrapped,
	}, nil
}

func (this *regularFile) Readdir(count int) ([]os.FileInfo, error) {
	
	ret, err := this.wrapped.Readdir(count)
	if err != nil {
		return ret, err
	}

	// go through each FileInfo, replace with wrapped if encrypted.
	for i, info := range ret {

		name := info.Name()
		if strings.HasSuffix(name, encryptedExtension) {

			trimmed := name[:len(name)-len(encryptedExtension)]

			ret[i] = overrideFileInfo {
				FixedName: trimmed,
				wrapped: info,
			}
		}
	}

	return ret, nil
}

//

func (this *regularFile) Read(p []byte) (n int, err error) {
	return this.wrapped.Read(p)
}

func (this *regularFile) Seek(offset int64, whence int) (n int64, err error) {
	return this.wrapped.Seek(offset, whence)
}

func (this *regularFile) Stat() (os.FileInfo, error) {
	return this.wrapped.Stat()
}

func (this *regularFile) Close() error {
	return this.wrapped.Close()
}

func (this *regularFile) Write(p []byte) (n int, err error) {
	return this.wrapped.Write(p)
}