package boji

import (
	"os"
	"net/http"
)

type plainFilesystem struct {
	fs http.FileSystem
}

func (fs plainFilesystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return dirfile{f}, nil
}

type dirfile struct {
	http.File
}

func (f dirfile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}