// Package interfaces はtitles_thコマンドで使用するインターフェースを定義します
package interfaces

import (
	"context"
	"io"

	"github.com/shiroemons/go-brightmoon/internal/titles/models"
	"github.com/shiroemons/go-brightmoon/pkg/pbgarc"
)

// FileSystem はファイルシステム操作のインターフェース
type FileSystem interface {
	FileExists(filename string) bool
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm uint32) error
	MkdirAll(path string, perm uint32) error
	Stat(name string) (FileInfo, error)
	ReadDir(dirname string) ([]DirEntry, error)
	Getwd() (string, error)
	Executable() (string, error)
}

// FileInfo はファイル情報のインターフェース
type FileInfo interface {
	Name() string
	IsDir() bool
}

// DirEntry はディレクトリエントリのインターフェース
type DirEntry interface {
	Name() string
	IsDir() bool
}

// Extractor はアーカイブからファイルを抽出するインターフェースです
type Extractor interface {
	ExtractFiles(ctx context.Context, archivePath string, archiveType int, targetFiles []string) (map[string][]byte, error)
}

// DatFileFinder は.datファイルを検索するインターフェースです
type DatFileFinder interface {
	Find() (string, error)
}

// ArchiveExtractor はアーカイブからファイルを抽出するインターフェース
type ArchiveExtractor interface {
	ExtractFiles(archivePath string, archiveType int, targetFiles []string) (map[string][]byte, error)
}

// ArchiveOpener はアーカイブを開くためのインターフェース
type ArchiveOpener interface {
	Open(filename string) (pbgarc.PBGArchive, error)
	OpenWithType(filename string, archiveType int) (pbgarc.PBGArchive, error)
}

// Parser はデータを解析するインターフェース
type Parser interface {
	ParseTHFmt(data []byte) ([]*models.Record, error)
	ParseMusicCmt(data string) ([]*models.Track, error)
}

// AdditionalInfoChecker は補足情報をチェックするインターフェース
type AdditionalInfoChecker interface {
	CheckAdditionalInfo(archivePath string) models.AdditionalInfo
}

// FileFinder はファイルを検索するインターフェース
type FileFinder interface {
	Find() (string, error)
}

// Writer は出力を書き込むインターフェース
type Writer interface {
	io.Writer
}

// Logger はログ出力のインターフェース
type Logger interface {
	Printf(format string, a ...any)
}
