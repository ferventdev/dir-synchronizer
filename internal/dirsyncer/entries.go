package dirsyncer

import "time"

type PathInfo struct {
	Exists  bool
	IsDir   bool
	Size    int64 // in bytes
	ModTime time.Time
}

type EntryInfo struct {
	SrcPathInfo, CopyPathInfo PathInfo
}
