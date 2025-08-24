package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shiroemons/go-brightmoon/internal/titles/interfaces"
)

// OSFileSystem は実際のOSファイルシステムを使用する実装
type OSFileSystem struct{}

// NewOSFileSystem は新しいOSFileSystemを作成します
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// FileExists はファイルが存在するか確認します
func (fs *OSFileSystem) FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// ReadFile はファイルを読み込みます
func (fs *OSFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// WriteFile はファイルを書き込みます
func (fs *OSFileSystem) WriteFile(filename string, data []byte, perm uint32) error {
	return os.WriteFile(filename, data, os.FileMode(perm))
}

// MkdirAll はディレクトリを作成します
func (fs *OSFileSystem) MkdirAll(path string, perm uint32) error {
	return os.MkdirAll(path, os.FileMode(perm))
}

// Stat はファイル情報を取得します
func (fs *OSFileSystem) Stat(name string) (interfaces.FileInfo, error) {
	info, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	return &osFileInfo{info}, nil
}

// ReadDir はディレクトリを読み込みます
func (fs *OSFileSystem) ReadDir(dirname string) ([]interfaces.DirEntry, error) {
	entries, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	result := make([]interfaces.DirEntry, len(entries))
	for i, entry := range entries {
		result[i] = &osDirEntry{entry}
	}
	return result, nil
}

// Getwd は現在の作業ディレクトリを取得します
func (fs *OSFileSystem) Getwd() (string, error) {
	return os.Getwd()
}

// Executable は実行ファイルのパスを取得します
func (fs *OSFileSystem) Executable() (string, error) {
	return os.Executable()
}

// osFileInfo はos.FileInfoのラッパー
type osFileInfo struct {
	os.FileInfo
}

// Name はファイル名を返します
func (fi *osFileInfo) Name() string {
	return fi.FileInfo.Name()
}

// IsDir はディレクトリかどうかを返します
func (fi *osFileInfo) IsDir() bool {
	return fi.FileInfo.IsDir()
}

// osDirEntry はos.DirEntryのラッパー
type osDirEntry struct {
	os.DirEntry
}

// Name はエントリ名を返します
func (de *osDirEntry) Name() string {
	return de.DirEntry.Name()
}

// IsDir はディレクトリかどうかを返します
func (de *osDirEntry) IsDir() bool {
	return de.DirEntry.IsDir()
}

// DatFileFinderWithFS は.datファイルの検索を行います（FileSystemを使用）
type DatFileFinderWithFS struct {
	fs interfaces.FileSystem
}

// NewDatFileFinderWithFS は新しいDatFileFinderWithFSを作成します
func NewDatFileFinderWithFS(fs interfaces.FileSystem) *DatFileFinderWithFS {
	return &DatFileFinderWithFS{fs: fs}
}

// Find は実行ファイルと同じディレクトリおよびカレントディレクトリから.datファイルを検索します
func (f *DatFileFinderWithFS) Find() (string, error) {
	var datFiles []string

	// カレントディレクトリを取得
	currentDir, err := f.fs.Getwd()
	if err != nil {
		return "", err
	}

	// まずカレントディレクトリを検索
	currentDirFiles, err := f.findInDir(currentDir)
	if err != nil {
		return "", err
	}
	datFiles = append(datFiles, currentDirFiles...)

	// カレントディレクトリで見つかった場合
	if len(datFiles) > 0 {
		if len(datFiles) > 1 {
			return "", f.createMultipleFilesError(datFiles)
		}
		return datFiles[0], nil
	}

	// 実行ファイルのパスを取得
	execPath, err := f.fs.Executable()
	if err != nil {
		return "", err
	}

	// 実行ファイルのディレクトリを取得
	execDir := filepath.Dir(execPath)

	// 実行ファイルのディレクトリを検索
	execDirFiles, err := f.findInDir(execDir)
	if err != nil {
		return "", err
	}
	datFiles = append(datFiles, execDirFiles...)

	if len(datFiles) == 0 {
		return "", nil
	}

	if len(datFiles) > 1 {
		return "", f.createMultipleFilesError(datFiles)
	}

	return datFiles[0], nil
}

// findInDir は指定されたディレクトリ内のthxx.datやthxxtr.datファイルを検索します
func (f *DatFileFinderWithFS) findInDir(dir string) ([]string, error) {
	var datFiles []string

	files, err := f.fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		// thbgm.dat は除外
		if name == "thbgm.dat" {
			continue
		}

		if DatFilePattern.MatchString(name) {
			datFiles = append(datFiles, filepath.Join(dir, name))
		}
	}

	return datFiles, nil
}

// createMultipleFilesError は複数の.datファイルが見つかった場合のエラーを生成します
func (f *DatFileFinderWithFS) createMultipleFilesError(datFiles []string) error {
	fileNames := make([]string, len(datFiles))
	for i, path := range datFiles {
		fileNames[i] = filepath.Base(path)
	}
	return fmt.Errorf("複数の.datファイルが見つかりました: %s。-archive フラグで使用するファイルを指定してください", strings.Join(fileNames, ", "))
}
