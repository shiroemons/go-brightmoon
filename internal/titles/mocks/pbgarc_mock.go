package mocks

import (
	"errors"
	"io"

	"github.com/shiroemons/go-brightmoon/pkg/pbgarc"
)

var (
	// ErrMockOpenFailed はオープン失敗エラー
	ErrMockOpenFailed = errors.New("mock: failed to open archive")
	// ErrMockExtractFailed は抽出失敗エラー
	ErrMockExtractFailed = errors.New("mock: failed to extract file")
	// ErrMockFileNotFound はファイル未発見エラー
	ErrMockFileNotFound = errors.New("mock: file not found in archive")
)

// MockPBGArchive はPBGArchiveインターフェースのモック実装
type MockPBGArchive struct {
	Files          map[string][]byte
	CurrentFile    string
	CurrentIndex   int
	FileList       []string
	OpenSuccess    bool
	OpenError      error
	ExtractSuccess bool
	ExtractError   error
	CloseError     error
}

// NewMockPBGArchive は新しいMockPBGArchiveを作成
func NewMockPBGArchive(files map[string][]byte) *MockPBGArchive {
	fileList := make([]string, 0, len(files))
	for name := range files {
		fileList = append(fileList, name)
	}
	
	return &MockPBGArchive{
		Files:          files,
		FileList:       fileList,
		CurrentIndex:   -1,
		OpenSuccess:    true,
		ExtractSuccess: true,
	}
}

// Open はモック実装
func (m *MockPBGArchive) Open(filename string) (bool, error) {
	if m.OpenError != nil {
		return false, m.OpenError
	}
	if !m.OpenSuccess {
		return false, ErrMockOpenFailed
	}
	return true, nil
}

// Close はモック実装
func (m *MockPBGArchive) Close() error {
	if m.CloseError != nil {
		return m.CloseError
	}
	return nil
}

// EnumFirst はモック実装
func (m *MockPBGArchive) EnumFirst() bool {
	if len(m.FileList) == 0 {
		return false
	}
	m.CurrentIndex = 0
	m.CurrentFile = m.FileList[0]
	return true
}

// EnumNext はモック実装
func (m *MockPBGArchive) EnumNext() bool {
	if m.CurrentIndex >= len(m.FileList)-1 {
		return false
	}
	m.CurrentIndex++
	m.CurrentFile = m.FileList[m.CurrentIndex]
	return true
}

// GetEntryName はモック実装
func (m *MockPBGArchive) GetEntryName() string {
	return m.CurrentFile
}

// GetOriginalSize はモック実装
func (m *MockPBGArchive) GetOriginalSize() uint32 {
	if data, ok := m.Files[m.CurrentFile]; ok {
		return uint32(len(data))
	}
	return 0
}

// GetCompressedSize はモック実装
func (m *MockPBGArchive) GetCompressedSize() uint32 {
	if data, ok := m.Files[m.CurrentFile]; ok {
		return uint32(len(data))
	}
	return 0
}

// GetEntry はモック実装
func (m *MockPBGArchive) GetEntry() pbgarc.PBGArchiveEntry {
	return &MockPBGArchiveEntry{
		Name: m.CurrentFile,
		Size: m.GetOriginalSize(),
	}
}

// Extract はモック実装
func (m *MockPBGArchive) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if m.ExtractError != nil {
		return false
	}
	if !m.ExtractSuccess {
		return false
	}
	
	if data, ok := m.Files[m.CurrentFile]; ok {
		if callback != nil {
			if !callback(m.CurrentFile, user) {
				return false
			}
		}
		_, err := w.Write(data)
		return err == nil
	}
	return false
}

// ExtractAll はモック実装
func (m *MockPBGArchive) ExtractAll(callback func(string, interface{}) bool, user interface{}) bool {
	for name, data := range m.Files {
		if callback != nil {
			if !callback(name, user) {
				return false
			}
		}
		_ = data // 実際には何もしない
	}
	return true
}

// MockPBGArchiveEntry はPBGArchiveEntryのモック実装
type MockPBGArchiveEntry struct {
	Name string
	Size uint32
}

// GetEntryName はモック実装
func (e *MockPBGArchiveEntry) GetEntryName() string {
	return e.Name
}

// GetOriginalSize はモック実装
func (e *MockPBGArchiveEntry) GetOriginalSize() uint32 {
	return e.Size
}

// GetCompressedSize はモック実装
func (e *MockPBGArchiveEntry) GetCompressedSize() uint32 {
	return e.Size
}

// Extract はモック実装
func (e *MockPBGArchiveEntry) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if callback != nil {
		return callback(e.Name, user)
	}
	return true
}

// MockPBGArchiveFactory はアーカイブファクトリのモック
type MockPBGArchiveFactory struct {
	Archive      pbgarc.PBGArchive
	OpenError    error
	OpenAutoArch pbgarc.PBGArchive
	OpenAutoErr  error
}

// Open はモック実装
func (f *MockPBGArchiveFactory) Open(filename string, archiveType int) (pbgarc.PBGArchive, error) {
	if f.OpenError != nil {
		return nil, f.OpenError
	}
	if f.Archive != nil {
		return f.Archive, nil
	}
	return NewMockPBGArchive(map[string][]byte{}), nil
}

// OpenAuto はモック実装
func (f *MockPBGArchiveFactory) OpenAuto(filename string) (pbgarc.PBGArchive, error) {
	if f.OpenAutoErr != nil {
		return nil, f.OpenAutoErr
	}
	if f.OpenAutoArch != nil {
		return f.OpenAutoArch, nil
	}
	return NewMockPBGArchive(map[string][]byte{}), nil
}