package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEntryInfo_IsSyncRequired(t *testing.T) {
	tests := []struct {
		name  string
		entry *EntryInfo
		want  bool
	}{
		{
			name:  "both not exist",
			entry: &EntryInfo{SrcPathInfo: PathInfo{Exists: false}, CopyPathInfo: PathInfo{Exists: false}},
			want:  false,
		},
		{
			name: "both are dirs",
			entry: &EntryInfo{
				SrcPathInfo:  PathInfo{Exists: true, IsDir: true},
				CopyPathInfo: PathInfo{Exists: true, IsDir: true},
			},
			want: false,
		},
		{
			name:  "only source exist",
			entry: &EntryInfo{SrcPathInfo: PathInfo{Exists: true}, CopyPathInfo: PathInfo{Exists: false}},
			want:  true,
		},
		{
			name:  "only copy exist",
			entry: &EntryInfo{SrcPathInfo: PathInfo{Exists: false}, CopyPathInfo: PathInfo{Exists: true}},
			want:  true,
		},
		{
			name: "only source is dir",
			entry: &EntryInfo{
				SrcPathInfo:  PathInfo{Exists: true, IsDir: true},
				CopyPathInfo: PathInfo{Exists: true, IsDir: false},
			},
			want: true,
		},
		{
			name: "only copy is dir",
			entry: &EntryInfo{
				SrcPathInfo:  PathInfo{Exists: true, IsDir: false},
				CopyPathInfo: PathInfo{Exists: true, IsDir: true},
			},
			want: true,
		},
		{
			name: "size differs",
			entry: &EntryInfo{
				SrcPathInfo:  PathInfo{Exists: true, Size: 10},
				CopyPathInfo: PathInfo{Exists: true, Size: 20},
			},
			want: true,
		},
		{
			name: "modTime differs",
			entry: &EntryInfo{
				SrcPathInfo:  PathInfo{Exists: true, ModTime: time.Unix(10000, 0)},
				CopyPathInfo: PathInfo{Exists: true, ModTime: time.Unix(10100, 0)},
			},
			want: true,
		},
		{
			name: "same files",
			entry: &EntryInfo{
				SrcPathInfo:  PathInfo{Exists: true, Size: 10, ModTime: time.Unix(10000, 0)},
				CopyPathInfo: PathInfo{Exists: true, Size: 10, ModTime: time.Unix(10000, 0)},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.entry.IsSyncRequired())
		})
	}
}
