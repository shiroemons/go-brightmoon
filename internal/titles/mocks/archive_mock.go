package mocks

import (
	"errors"
	"io"

	"github.com/shiroemons/go-brightmoon/pkg/pbgarc"
)

// MockArchiveFactory はテスト用のアーカイブファクトリ
type MockArchiveFactory struct {
	MockArchive pbgarc.PBGArchive
	Error       error
}

func (f *MockArchiveFactory) NewYumemiArchive() pbgarc.PBGArchive {
	if f.Error != nil {
		return nil
	}
	return f.MockArchive
}

func (f *MockArchiveFactory) NewKaguyaArchive() pbgarc.PBGArchive {
	if f.Error != nil {
		return nil
	}
	return f.MockArchive
}

func (f *MockArchiveFactory) NewSuicaArchive() pbgarc.PBGArchive {
	if f.Error != nil {
		return nil
	}
	return f.MockArchive
}

func (f *MockArchiveFactory) NewHinanawiArchive() pbgarc.PBGArchive {
	if f.Error != nil {
		return nil
	}
	return f.MockArchive
}

func (f *MockArchiveFactory) NewMarisaArchive() pbgarc.PBGArchive {
	if f.Error != nil {
		return nil
	}
	return f.MockArchive
}

func (f *MockArchiveFactory) NewKanakoArchive() pbgarc.PBGArchive {
	if f.Error != nil {
		return nil
	}
	return f.MockArchive
}

// MockMemoryExtractor はテスト用のメモリ抽出モック
type MockMemoryExtractor struct {
	Data  []byte
	Error error
}

func (e *MockMemoryExtractor) ExtractToMemory(archive pbgarc.PBGArchive) ([]byte, error) {
	if e.Error != nil {
		return nil, e.Error
	}
	return e.Data, nil
}

// SimpleMockArchive は最小限のPBGArchive実装
type SimpleMockArchive struct {
	Files           map[string][]byte
	CurrentFile     string
	FileNames       []string
	CurrentIndex    int
	OpenSuccess     bool
	ExtractSuccess  bool
	ShouldFailOpen  bool
	ShouldFailFirst bool
}

func NewSimpleMockArchive(files map[string][]byte) *SimpleMockArchive {
	fileNames := make([]string, 0, len(files))
	for name := range files {
		fileNames = append(fileNames, name)
	}
	return &SimpleMockArchive{
		Files:          files,
		FileNames:      fileNames,
		CurrentIndex:   -1,
		OpenSuccess:    true,
		ExtractSuccess: true,
	}
}

func (a *SimpleMockArchive) Open(filename string) (bool, error) {
	if a.ShouldFailOpen {
		return false, errors.New("open failed")
	}
	a.OpenSuccess = true
	return true, nil
}

func (a *SimpleMockArchive) Close() {}

func (a *SimpleMockArchive) EnumFirst() bool {
	if a.ShouldFailFirst || len(a.FileNames) == 0 {
		return false
	}
	a.CurrentIndex = 0
	a.CurrentFile = a.FileNames[0]
	return true
}

func (a *SimpleMockArchive) EnumNext() bool {
	a.CurrentIndex++
	if a.CurrentIndex >= len(a.FileNames) {
		return false
	}
	a.CurrentFile = a.FileNames[a.CurrentIndex]
	return true
}

func (a *SimpleMockArchive) GetEntryName() string {
	return a.CurrentFile
}

func (a *SimpleMockArchive) GetOriginalSize() uint32 {
	if data, ok := a.Files[a.CurrentFile]; ok {
		return uint32(len(data))
	}
	return 0
}

func (a *SimpleMockArchive) GetCompressedSize() uint32 {
	return a.GetOriginalSize()
}

func (a *SimpleMockArchive) Extract(w io.Writer, callback func(string, any) bool, user any) bool {
	if !a.ExtractSuccess {
		return false
	}
	if data, ok := a.Files[a.CurrentFile]; ok {
		// Writerに書き込み
		_, err := w.Write(data)
		return err == nil
	}
	return false
}

func (a *SimpleMockArchive) ExtractAll(callback func(string, any) bool, user any) bool {
	return true
}

func (a *SimpleMockArchive) GetTime() int64 {
	return 0
}

func (a *SimpleMockArchive) GetEntry() pbgarc.PBGArchiveEntry {
	return nil
}
