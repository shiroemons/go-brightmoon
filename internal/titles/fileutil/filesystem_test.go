package fileutil

import (
	"errors"
	"os"
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

func TestOSFileSystem_ReadFile(t *testing.T) {
	fs := NewOSFileSystem()

	// 一時ファイルを作成
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	testContent := []byte("test content")
	if _, err := tmpfile.Write(testContent); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// ReadFile のテスト
	data, err := fs.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != string(testContent) {
		t.Errorf("ReadFile content mismatch: got %s, want %s", data, testContent)
	}

	// 存在しないファイルのテスト
	_, err = fs.ReadFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ReadFile should return error for non-existent file")
	}
}

func TestOSFileSystem_WriteFile(t *testing.T) {
	fs := NewOSFileSystem()

	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test write content")

	// WriteFile のテスト
	err = fs.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// ファイルが正しく書き込まれたか確認
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(data) != string(testContent) {
		t.Errorf("Written content mismatch: got %s, want %s", data, testContent)
	}
}

func TestOSFileSystem_MkdirAll(t *testing.T) {
	fs := NewOSFileSystem()

	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testDir := filepath.Join(tmpDir, "nested", "dirs", "test")

	// MkdirAll のテスト
	err = fs.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// ディレクトリが作成されたか確認
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Created path is not a directory")
	}
}

func TestOSFileSystem_Stat(t *testing.T) {
	fs := NewOSFileSystem()

	// 一時ファイルを作成
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// Stat のテスト
	info, err := fs.Stat(tmpfile.Name())
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// OSFileInfo のテスト
	if info.Name() == "" {
		t.Error("Name() should return non-empty string")
	}
	if info.IsDir() {
		t.Error("IsDir() should return false for file")
	}

	// 存在しないファイルのテスト
	_, err = fs.Stat("/nonexistent/file.txt")
	if err == nil {
		t.Error("Stat should return error for non-existent file")
	}
}

func TestOSFileSystem_ReadDir(t *testing.T) {
	fs := NewOSFileSystem()

	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// テストファイルとディレクトリを作成
	testFile := filepath.Join(tmpDir, "test.txt")
	testDir := filepath.Join(tmpDir, "subdir")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	// ReadDir のテスト
	entries, err := fs.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// OSDirEntry のテスト
	for _, entry := range entries {
		if entry.Name() == "" {
			t.Error("Entry Name() should return non-empty string")
		}
		if entry.Name() == "subdir" && !entry.IsDir() {
			t.Error("IsDir() should return true for directory")
		}
		if entry.Name() == "test.txt" && entry.IsDir() {
			t.Error("IsDir() should return false for file")
		}
	}

	// 存在しないディレクトリのテスト
	_, err = fs.ReadDir("/nonexistent/dir")
	if err == nil {
		t.Error("ReadDir should return error for non-existent directory")
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
