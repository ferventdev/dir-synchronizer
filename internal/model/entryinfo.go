package model

import "time"

//PathInfo holds info about one dir entry in a file tree (of either source OR copy directory).
type PathInfo struct {
	Exists   bool
	FullPath string
	IsDir    bool
	Size     int64 // in bytes
	ModTime  time.Time
}

//EntryInfo holds info about same dir entry in BOTH files trees (source and copy) and the sync operation between them.
type EntryInfo struct {
	SrcPathInfo, CopyPathInfo PathInfo
	operation                 *Operation
}

//SetSrcPathInfo is a convenience setter for (d *dirScanner) walk.
func (e *EntryInfo) SetSrcPathInfo(pi PathInfo) {
	e.SrcPathInfo = pi
}

//SetCopyPathInfo  is a convenience setter for (d *dirScanner) walk.
func (e *EntryInfo) SetCopyPathInfo(pi PathInfo) {
	e.CopyPathInfo = pi
}
