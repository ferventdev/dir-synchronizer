package settings

import (
	"dsync/internal/log"
	"flag"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		commandArgs []string
		panic       bool
		wantErr     bool
		want        *Settings
	}{
		{name: "flag panic 1", commandArgs: []string{"-undefined"}, panic: true, wantErr: false, want: nil},
		{name: "flag panic 2", commandArgs: []string{"-once=123"}, panic: true, wantErr: false, want: nil},
		{name: "flag panic 3", commandArgs: []string{"-loglvl"}, panic: true, wantErr: false, want: nil},
		{name: "flag panic 4", commandArgs: []string{"-workers=a"}, panic: true, wantErr: false, want: nil},
		{name: "flag panic 5", commandArgs: []string{"-scanperiod=b"}, panic: true, wantErr: false, want: nil},
		{name: "no args", commandArgs: nil, panic: false, wantErr: true, want: nil},
		{name: "not enough args", commandArgs: []string{"a"}, panic: false, wantErr: true, want: nil},
		{name: "bad level", commandArgs: []string{"-loglvl=nope", "d1", "d2"}, panic: false, wantErr: true, want: nil},
		{name: "same dirs", commandArgs: []string{"dir", "dir"}, panic: false, wantErr: true, want: nil},
		{
			name: "valid args",
			commandArgs: []string{"-hidden", "-copydirs", "-log2std", "-once", "-pid",
				"-loglvl=debug", "-scanperiod=3s", "-workers=10", "dir1", "dir2"},
			panic:   false,
			wantErr: false,
			want: &Settings{
				SrcDir:           abs("dir1"),
				CopyDir:          abs("dir2"),
				ScanPeriod:       3 * time.Second,
				IncludeHidden:    true,
				IncludeEmptyDirs: true,
				LogLevel:         log.DebugLevel,
				LogToStd:         true,
				Once:             true,
				PrintPID:         true,
				WorkersCount:     10,
			},
		},
		{
			name:        "default args",
			commandArgs: []string{"dir1", "dir2"},
			panic:       false,
			wantErr:     false,
			want: &Settings{
				SrcDir:           abs("dir1"),
				CopyDir:          abs("dir2"),
				ScanPeriod:       time.Second,
				IncludeHidden:    false,
				IncludeEmptyDirs: false,
				LogLevel:         log.InfoLevel,
				LogToStd:         false,
				Once:             false,
				PrintPID:         false,
				WorkersCount:     runtime.NumCPU(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				stg *Settings
				err error
			)

			requires := require.New(t)
			if tt.panic {
				requires.Panics(func() {
					stg, err = New(tt.commandArgs, flag.PanicOnError)
				})
				return
			}

			requires.NotPanics(func() {
				stg, err = New(tt.commandArgs, flag.PanicOnError)
			})

			if tt.wantErr {
				requires.Error(err)
				requires.Nil(stg)
				return
			}

			requires.NoError(err)
			requires.NotNil(stg)
			requires.Equal(*tt.want, *stg)
		})
	}
}

func TestSettings_Validate(t *testing.T) {
	type fields struct {
		SrcDir       string
		CopyDir      string
		ScanPeriod   time.Duration
		WorkersCount int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		errText string
	}{
		{
			name:    "no source dir",
			fields:  fields{SrcDir: "noSuchDir", CopyDir: "cpDir"},
			wantErr: true,
			errText: "the first (source) directory is invalid",
		},
		{
			name:    "source not dir",
			fields:  fields{SrcDir: "settings.go", CopyDir: "cpDir"},
			wantErr: true,
			errText: "is not a directory path",
		},
		{
			name:    "no copy dir",
			fields:  fields{SrcDir: "../settings", CopyDir: "noSuchDir"},
			wantErr: true,
			errText: "the second (copy) directory is invalid",
		},
		{
			name:    "copy not dir",
			fields:  fields{SrcDir: "../settings", CopyDir: "settings.go"},
			wantErr: true,
			errText: "is not a directory path",
		},
		{
			name:    "bad scan period",
			fields:  fields{SrcDir: "../settings", CopyDir: "../model", ScanPeriod: maxScanPeriod + time.Second},
			wantErr: true,
			errText: "period of directories scanning must be a value between",
		},
		{
			name:    "bad workers count",
			fields:  fields{SrcDir: "../settings", CopyDir: "../model", ScanPeriod: minScanPeriod, WorkersCount: maxWorkersCount + 1},
			wantErr: true,
			errText: "number of workers must be a value between",
		},
		{
			name:    "ok",
			fields:  fields{SrcDir: "../settings", CopyDir: "../model", ScanPeriod: minScanPeriod, WorkersCount: minWorkersCount},
			wantErr: false,
			errText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (&Settings{
				SrcDir:       tt.fields.SrcDir,
				CopyDir:      tt.fields.CopyDir,
				ScanPeriod:   tt.fields.ScanPeriod,
				WorkersCount: tt.fields.WorkersCount,
			}).Validate()

			requires := require.New(t)
			if tt.wantErr {
				requires.ErrorContains(err, tt.errText)
				return
			}

			requires.NoError(err)
		})
	}
}

func abs(path string) string {
	s, _ := filepath.Abs(path)
	return s
}
