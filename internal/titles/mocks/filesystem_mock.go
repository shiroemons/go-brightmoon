// Package mocks はテスト用のモック実装を提供します
package mocks

import (
	"errors"
	"path/filepath"

	"github.com/shiroemons/go-brightmoon/internal/titles/interfaces"
)

// MockFileSystem はテスト用のファイルシステムモック
type MockFileSystem struct {
	Files      map[string][]byte
	Dirs       map[string]bool
	WorkingDir string
	ExecPath   string
	Error      error
}

// NewMockFileSystem は新しいMockFileSystemを作成します
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files:      make(map[string][]byte),
		Dirs:       make(map[string]bool),
		WorkingDir: "/test/dir",
		ExecPath:   "/test/exec/program",
	}
}

// FileExists はファイルが存在するか確認します
func (fs *MockFileSystem) FileExists(filename string) bool {
	_, exists := fs.Files[filename]
	return exists
}

// ReadFile はファイルを読み込みます
func (fs *MockFileSystem) ReadFile(filename string) ([]byte, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	data, exists := fs.Files[filename]
	if !exists {
		return nil, errors.New("file not found")
	}
	return data, nil
}

// WriteFile はファイルを書き込みます
func (fs *MockFileSystem) WriteFile(filename string, data []byte, perm uint32) error {
	if fs.Error != nil {
		return fs.Error
	}
	fs.Files[filename] = data
	return nil
}

// MkdirAll はディレクトリを作成します
func (fs *MockFileSystem) MkdirAll(path string, perm uint32) error {
	if fs.Error != nil {
		return fs.Error
	}
	fs.Dirs[path] = true
	return nil
}

// Stat はファイル情報を取得します
func (fs *MockFileSystem) Stat(name string) (interfaces.FileInfo, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}
	if _, exists := fs.Files[name]; exists {
		return &MockFileInfo{name: filepath.Base(name), isDir: false}, nil
	}
	if _, exists := fs.Dirs[name]; exists {
		return &MockFileInfo{name: filepath.Base(name), isDir: true}, nil
	}
	return nil, errors.New("file not found")
}

// ReadDir はディレクトリを読み込みます
func (fs *MockFileSystem) ReadDir(dirname string) ([]interfaces.DirEntry, error) {
	if fs.Error != nil {
		return nil, fs.Error
	}

	// ディレクトリが存在するか確認
	if !fs.Dirs[dirname] {
		// ディレクトリが明示的に設定されていない場合でも、
		// ファイルが存在する場合は空のエントリリストを返す
		hasFiles := false
		for path := range fs.Files {
			if filepath.Dir(path) == dirname {
				hasFiles = true
				break
			}
		}
		if !hasFiles {
			return nil, errors.New("directory not found")
		}
	}

	var entries []interfaces.DirEntry
	// ファイルをディレクトリエントリとして追加
	for path := range fs.Files {
		dir := filepath.Dir(path)
		if dir == dirname {
			entries = append(entries, &MockDirEntry{
				name:  filepath.Base(path),
				isDir: false,
			})
		}
	}
	// サブディレクトリをエントリとして追加
	for path := range fs.Dirs {
		dir := filepath.Dir(path)
		if dir == dirname && path != dirname {
			entries = append(entries, &MockDirEntry{
				name:  filepath.Base(path),
				isDir: true,
			})
		}
	}

	return entries, nil
}

// Getwd は現在の作業ディレクトリを返します
func (fs *MockFileSystem) Getwd() (string, error) {
	if fs.Error != nil {
		return "", fs.Error
	}
	return fs.WorkingDir, nil
}

// Executable は実行ファイルのパスを返します
func (fs *MockFileSystem) Executable() (string, error) {
	if fs.Error != nil {
		return "", fs.Error
	}
	return fs.ExecPath, nil
}

// MockFileInfo はテスト用のFileInfo実装
type MockFileInfo struct {
	name  string
	isDir bool
}

// Name はファイル名を返します
func (fi *MockFileInfo) Name() string {
	return fi.name
}

// IsDir はディレクトリかどうかを返します
func (fi *MockFileInfo) IsDir() bool {
	return fi.isDir
}

// MockDirEntry はテスト用のDirEntry実装
type MockDirEntry struct {
	name  string
	isDir bool
}

// Name はエントリ名を返します
func (de *MockDirEntry) Name() string {
	return de.name
}

// IsDir はディレクトリかどうかを返します
func (de *MockDirEntry) IsDir() bool {
	return de.isDir
}
