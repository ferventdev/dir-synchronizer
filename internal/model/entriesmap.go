package model

import (
	"sync"
)

//DirEntriesMap is the main data structure of the application.
//It holds a map with dir entries of the source and copy file trees.
//A key in this map is a relative path of one dir entry, and a value is this entry's info (in both source and copy file trees).
type DirEntriesMap struct {
	mu   sync.Mutex
	eMap map[string]EntryInfo
}

func NewDirEntriesMap() *DirEntriesMap {
	return &DirEntriesMap{eMap: make(map[string]EntryInfo, 10)}
}

func (m *DirEntriesMap) PrepareForScan() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.eMap {
		// we reset the existence flags at the beginning of each file trees scanning;
		// they will be set back to true for those entries which will be found during the file trees walks
		entry.SrcPathInfo.Exists = false
		entry.CopyPathInfo.Exists = false
	}
}

func (m *DirEntriesMap) UpdateValueByKey(key string, valueUpdater func(entry *EntryInfo)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry := m.eMap[key] // entry's zero value will be fine as well
	valueUpdater(&entry)
	m.eMap[key] = entry
}

func (m *DirEntriesMap) RemoveObsolete() {
	m.mu.Lock()
	defer m.mu.Unlock()
	keysForRemoval := make([]string, 0, len(m.eMap))
	for key, e := range m.eMap {
		isObsolete := !e.SrcPathInfo.Exists && !e.CopyPathInfo.Exists && e.Operation == nil
		if isObsolete {
			keysForRemoval = append(keysForRemoval, key)
		}
	}
	for _, key := range keysForRemoval {
		delete(m.eMap, key)
	}
}
