package model

import "time"

//PathInfo holds info about one dir entry in a file tree (of either source OR copy directory).
type PathInfo struct {
	Exists   bool      `json:"exists"`
	FullPath string    `json:"-"`
	IsDir    bool      `json:"isDir,omitempty"`
	Size     int64     `json:"size,omitempty"` // in bytes
	ModTime  time.Time `json:"modTime"`
}

func (pi *PathInfo) IsSameAs(copy PathInfo) bool {
	if !pi.Exists && !copy.Exists {
		return true
	}
	return pi.Exists && copy.Exists && (pi.IsDir == copy.IsDir) && (pi.Size == copy.Size) && (pi.ModTime == copy.ModTime)
}

//EntryInfo holds info about same dir entry in BOTH files trees (source and copy) and the sync operation between them.
type EntryInfo struct {
	SrcPathInfo  PathInfo   `json:"src"`
	CopyPathInfo PathInfo   `json:"copy"`
	OperationPtr *Operation `json:"operation,omitempty"`
}

//SetSrcPathInfo is a convenience setter for (d *dirScanner) walk.
func (ei *EntryInfo) SetSrcPathInfo(pi PathInfo) {
	ei.SrcPathInfo = pi
}

//SetCopyPathInfo  is a convenience setter for (d *dirScanner) walk.
func (ei *EntryInfo) SetCopyPathInfo(pi PathInfo) {
	ei.CopyPathInfo = pi
}

func (ei *EntryInfo) SetOperation(op *Operation) {
	ei.OperationPtr = op
}

func (ei *EntryInfo) IsSyncRequired() bool {
	return !ei.SrcPathInfo.IsSameAs(ei.CopyPathInfo)
}

func (ei *EntryInfo) ResolveOperationKind() OperationKind {
	src, cp := ei.SrcPathInfo, ei.CopyPathInfo
	switch {
	case src.Exists && !cp.Exists:
		return OpKindCopy
	case !src.Exists && cp.Exists:
		return OpKindRemove
	case src.Exists && cp.Exists && !src.IsDir && !cp.IsDir && (src.Size != cp.Size || src.ModTime != cp.ModTime):
		// so far, I decided not to sync empty dirs, i.e. only all files (recursively) are synchronized
		// non-empty dirs will be synced automatically as a part of files full path
		return OpKindReplace
	default: // normally this will never happen
		return OpKindNone
	}
}
