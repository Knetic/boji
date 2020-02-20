package boji

import (
	"time"
	"os"
)

// FileInfo wrapper that overrides wrapped values for size or name, depending on what was given during construction.
type overrideFileInfo struct {
	FixedSize int64
	FixedName string
	wrapped os.FileInfo
}

// modified size check, to return the given size at construction time, rather than file size.
func (this overrideFileInfo) Size() int64 {
	if this.FixedSize <= 0 {
		return this.wrapped.Size()
	}
	return this.FixedSize
}

func (this overrideFileInfo) Name() string {
	if this.FixedName == "" {
		return this.wrapped.Name()
	}
	return this.FixedName
}

func (this overrideFileInfo) Mode() os.FileMode {
	return this.wrapped.Mode()
}
func (this overrideFileInfo) ModTime() time.Time {
	return this.wrapped.ModTime()
}
func (this overrideFileInfo) IsDir() bool {
	return this.wrapped.IsDir()
}
func (this overrideFileInfo) Sys() interface{} {
	return this.wrapped.Sys()
}