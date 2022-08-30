package model

import (
	"sync"
)

//DirEntriesMap is the main data structure of the application.
//It holds a map with dir entries of the source and copy file trees.
//A key in this map is a relative path of one dir entry, and a value is this entry's info (in both source and copy file trees).
//For the safety, concurrent access to the inner map is protected and controlled by a mutex.
type DirEntriesMap struct {
	mu   sync.Mutex
	eMap map[string]EntryInfo
}

func NewDirEntriesMap() *DirEntriesMap {
	return &DirEntriesMap{eMap: make(map[string]EntryInfo, 10)}
}

func (m *DirEntriesMap) PrepareForScan() error {
	return m.ForEach(
		func(key string, eMap map[string]EntryInfo) error {
			// we reset the existence flags at the beginning of each file trees scanning;
			// they will be set back to true for those entries which will be found during the file trees walks
			entry := eMap[key] // entry's zero value will be fine as well
			entry.SrcPathInfo.Exists = false
			entry.CopyPathInfo.Exists = false
			eMap[key] = entry
			return nil
		},
	)
}

func (m *DirEntriesMap) UpdateValueByKey(key string, valueUpdater func(*EntryInfo)) {
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
	for k, e := range m.eMap {
		isObsolete := !e.SrcPathInfo.Exists && !e.CopyPathInfo.Exists && e.OperationPtr == nil
		if isObsolete {
			keysForRemoval = append(keysForRemoval, k)
		}
	}
	for _, key := range keysForRemoval {
		delete(m.eMap, key)
	}
}

func (m *DirEntriesMap) ForEach(fn func(key string, eMap map[string]EntryInfo) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, _ := range m.eMap {
		if err := fn(k, m.eMap); err != nil {
			return err
		}
	}
	return nil
}
