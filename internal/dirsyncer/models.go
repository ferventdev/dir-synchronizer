package dirsyncer

import (
	"sync"
	"time"
)

type PathInfo struct {
	Exists  bool
	IsDir   bool
	Size    int64 // in bytes
	ModTime time.Time
}

type EntryInfo struct {
	SrcPathInfo, CopyPathInfo PathInfo
}

//setSrcPathInfo is a convenience setter for (d *dirScanner) walk
func (e *EntryInfo) setSrcPathInfo(pi PathInfo) {
	e.SrcPathInfo = pi
}

//setCopyPathInfo  is a convenience setter for (d *dirScanner) walk
func (e *EntryInfo) setCopyPathInfo(pi PathInfo) {
	e.CopyPathInfo = pi
}

type dirEntriesMap struct {
	mu   sync.Mutex
	eMap map[string]EntryInfo
}

func newDirEntriesMap() *dirEntriesMap {
	return &dirEntriesMap{eMap: make(map[string]EntryInfo, 10)}
}

func (m *dirEntriesMap) updateValueByKey(key string, valueUpdater func(entry *EntryInfo)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry := m.eMap[key]
	valueUpdater(&entry)
	m.eMap[key] = entry
}
