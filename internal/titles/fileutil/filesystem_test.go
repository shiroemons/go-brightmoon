package fileutil

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/shiroemons/go-brightmoon/internal/titles/mocks"
)

func TestDatFileFinderWithFS_Find(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockFileSystem)
		wantFile  string
		wantError bool
		errorMsg  string
	}{
		{
			name: "カレントディレクトリに1つのdatファイル",
			setupMock: func(fs *mocks.MockFileSystem) {
				fs.WorkingDir = "/current"
				fs.Dirs["/current"] = true
				fs.Files["/current/th08.dat"] = []byte("test")
			},
			wantFile: "/current/th08.dat",
		},
		{
			name: "カレントディレクトリに複数のdatファイル",
			setupMock: func(fs *mocks.MockFileSystem) {
				fs.WorkingDir = "/current"
				fs.Dirs["/current"] = true
				fs.Files["/current/th08.dat"] = []byte("test")
				fs.Files["/current/th09.dat"] = []byte("test")
			},
			wantError: true,
			errorMsg:  "複数の.datファイル",
		},
		{
			name: "実行ファイルディレクトリにdatファイル",
			setupMock: func(fs *mocks.MockFileSystem) {
				fs.WorkingDir = "/current"
				fs.ExecPath = "/exec/program"
				fs.Dirs["/current"] = true
				fs.Dirs["/exec"] = true
				fs.Files["/exec/th06.dat"] = []byte("test")
			},
			wantFile: "/exec/th06.dat",
		},
		{
			name: "体験版のdatファイル",
			setupMock: func(fs *mocks.MockFileSystem) {
				fs.WorkingDir = "/current"
				fs.Dirs["/current"] = true
				fs.Files["/current/th06tr.dat"] = []byte("test")
			},
			wantFile: "/current/th06tr.dat",
		},
		{
			name: "thbgm.datは除外される",
			setupMock: func(fs *mocks.MockFileSystem) {
				fs.WorkingDir = "/current"
				fs.ExecPath = "/exec/program"
				fs.Dirs["/current"] = true
				fs.Dirs["/exec"] = true
				fs.Files["/current/thbgm.dat"] = []byte("test")
			},
			wantFile: "",
		},
		{
			name: "datファイルが見つからない",
			setupMock: func(fs *mocks.MockFileSystem) {
				fs.WorkingDir = "/current"
				fs.ExecPath = "/exec/program"
				fs.Dirs["/current"] = true
				fs.Dirs["/exec"] = true
			},
			wantFile: "",
		},
		{
			name: "ReadDirエラー",
			setupMock: func(fs *mocks.MockFileSystem) {
				fs.WorkingDir = "/current"
				fs.Error = errors.New("read error")
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := mocks.NewMockFileSystem()
			tt.setupMock(fs)

			finder := NewDatFileFinderWithFS(fs)
			result, err := finder.Find()

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Error message should contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result != tt.wantFile {
					t.Errorf("Expected file '%s', got '%s'", tt.wantFile, result)
				}
			}
		})
	}
}

func TestOSFileSystem(t *testing.T) {
	fs := NewOSFileSystem()

	// FileExists のテスト（このファイル自体を使用）
	if !fs.FileExists("filesystem_test.go") {
		t.Error("FileExists should return true for existing file")
	}

	if fs.FileExists("nonexistent_file_xyz.go") {
		t.Error("FileExists should return false for non-existing file")
	}

	// Getwd のテスト
	wd, err := fs.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	if wd == "" {
		t.Error("Getwd should return non-empty string")
	}

	// Executable のテスト
	exec, err := fs.Executable()
	if err != nil {
		t.Fatalf("Executable failed: %v", err)
	}
	if exec == "" {
		t.Error("Executable should return non-empty string")
	}
}

func TestSaveToFileWithBOM_WithMockFS(t *testing.T) {
	fs := mocks.NewMockFileSystem()

	// テストデータ
	filename := "/test/output.txt"

	// MkdirAllとWriteFileをモック
	fs.Dirs[filepath.Dir(filename)] = true

	// OSFileSystemの代わりにMockFileSystemを使用したい場合の例
	// 実際のSaveToFileWithBOMは現在ファイルシステムに直接アクセスしているので、
	// インターフェース化が必要

	// モックファイルシステムにファイルが書き込まれたことを確認
	if len(fs.Files) == 0 {
		// 現状ではSaveToFileWithBOMは直接os.WriteFileを使用しているため、
		// このテストは実際には動作しません
		// 将来的にインターフェース化する必要があります
		t.Skip("SaveToFileWithBOM needs to be refactored to use FileSystem interface")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsString(s[1:], substr) || (len(s) > 0 && s[0:len(substr)] == substr))
}
