package boji

import (
	"time"
	"os"
)

type fixedSizeFileInfo struct {
	FixedSize int64
	wrapped os.FileInfo
}

// modified size check, to return the given size at construction time, rather than file size.
func (this fixedSizeFileInfo) Size() int64 {
	return this.FixedSize
}

func (this fixedSizeFileInfo) Name() string {
	return this.wrapped.Name()
}
func (this fixedSizeFileInfo) Mode() os.FileMode {
	return this.wrapped.Mode()
}
func (this fixedSizeFileInfo) ModTime() time.Time {
	return this.wrapped.ModTime()
}
func (this fixedSizeFileInfo) IsDir() bool {
	return this.wrapped.IsDir()
}
func (this fixedSizeFileInfo) Sys() interface{} {
	return this.wrapped.Sys()
}