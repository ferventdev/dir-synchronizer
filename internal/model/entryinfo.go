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

func (pi *PathInfo) IsSameAs(copy PathInfo) bool {
	if !pi.Exists && !copy.Exists {
		return true
	}
	return pi.Exists && copy.Exists && (pi.IsDir == copy.IsDir) && (pi.Size == copy.Size) && (pi.ModTime == copy.ModTime)
}

//EntryInfo holds info about same dir entry in BOTH files trees (source and copy) and the sync operation between them.
type EntryInfo struct {
	SrcPathInfo  PathInfo
	CopyPathInfo PathInfo
	Operation    *Operation
}

//SetSrcPathInfo is a convenience setter for (d *dirScanner) walk.
func (ei *EntryInfo) SetSrcPathInfo(pi PathInfo) {
	ei.SrcPathInfo = pi
}

//SetCopyPathInfo  is a convenience setter for (d *dirScanner) walk.
func (ei *EntryInfo) SetCopyPathInfo(pi PathInfo) {
	ei.CopyPathInfo = pi
}
