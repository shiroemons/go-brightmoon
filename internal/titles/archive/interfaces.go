package archive

import (
	"github.com/shiroemons/go-brightmoon/pkg/pbgarc"
)

// ArchiveOpener はアーカイブを開くためのインターフェース
type ArchiveOpener interface {
	Open(filename string) (pbgarc.PBGArchive, error)
	OpenWithType(filename string, archiveType int) (pbgarc.PBGArchive, error)
	AutoDetect(filename string) (pbgarc.PBGArchive, error)
}

// ArchiveFactory はアーカイブインスタンスを生成するインターフェース
type ArchiveFactory interface {
	NewYumemiArchive() pbgarc.PBGArchive
	NewKaguyaArchive() pbgarc.PBGArchive
	NewSuicaArchive() pbgarc.PBGArchive
	NewHinanawiArchive() pbgarc.PBGArchive
	NewMarisaArchive() pbgarc.PBGArchive
	NewKanakoArchive() pbgarc.PBGArchive
}

// DefaultArchiveFactory はデフォルトのアーカイブファクトリ実装
type DefaultArchiveFactory struct{}

func (f *DefaultArchiveFactory) NewYumemiArchive() pbgarc.PBGArchive {
	return pbgarc.NewYumemiArchive()
}

func (f *DefaultArchiveFactory) NewKaguyaArchive() pbgarc.PBGArchive {
	return pbgarc.NewKaguyaArchive()
}

func (f *DefaultArchiveFactory) NewSuicaArchive() pbgarc.PBGArchive {
	return pbgarc.NewSuicaArchive()
}

func (f *DefaultArchiveFactory) NewHinanawiArchive() pbgarc.PBGArchive {
	return pbgarc.NewHinanawiArchive()
}

func (f *DefaultArchiveFactory) NewMarisaArchive() pbgarc.PBGArchive {
	return pbgarc.NewMarisaArchive()
}

func (f *DefaultArchiveFactory) NewKanakoArchive() pbgarc.PBGArchive {
	return pbgarc.NewKanakoArchive()
}

// MemoryExtractor はメモリへの抽出を行うインターフェース
type MemoryExtractor interface {
	ExtractToMemory(archive pbgarc.PBGArchive) ([]byte, error)
}

// DefaultMemoryExtractor はデフォルトのメモリ抽出実装
type DefaultMemoryExtractor struct{}

func (e *DefaultMemoryExtractor) ExtractToMemory(archive pbgarc.PBGArchive) ([]byte, error) {
	origSize := archive.GetOriginalSize()
	if origSize == 0 {
		return nil, ErrEmptyFile
	}

	buf := make([]byte, 0, origSize)
	writer := &memoryWriter{buf: &buf}

	success := archive.Extract(writer, nil, nil)
	if !success {
		return nil, ErrExtractFailed
	}

	return *writer.buf, nil
}

// memoryWriter はメモリへの書き込み用Writer
type memoryWriter struct {
	buf *[]byte
}

func (w *memoryWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
