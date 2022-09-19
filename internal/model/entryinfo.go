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

func (pi *PathInfo) IsFile() bool {
	return !pi.IsDir
}

func (pi *PathInfo) IsSameAs(copy PathInfo) bool {
	if !pi.Exists && !copy.Exists {
		return true
	}
	//src and copy paths already refer to the same entry in DirEntriesMap, so we don't need to compare names (paths)
	if pi.Exists && copy.Exists && pi.IsDir && copy.IsDir {
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
	case src.Exists && src.IsFile() && !cp.Exists:
		return OpKindCopyFile
	case src.Exists && src.IsDir && !cp.Exists:
		// actually needed for empty dirs, because non-empty dirs are synced automatically as a part of files full path
		return OpKindCopyDir
	case (!src.Exists || src.IsDir) && cp.Exists && cp.IsFile():
		return OpKindRemoveFile
	case !src.Exists && cp.Exists && cp.IsDir: // actually works if cp is an empty dir
		return OpKindRemoveDir
	case src.Exists && cp.Exists && src.IsFile() && cp.IsDir: // actually works if cp is an empty dir
		return OpKindReplaceDirWithFile
	case src.Exists && cp.Exists && src.IsFile() && cp.IsFile() && (src.Size != cp.Size || src.ModTime != cp.ModTime):
		return OpKindReplaceFile
	default:
		return OpKindNone
	}
}
