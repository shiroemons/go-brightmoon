// Package fileutil はファイル操作のユーティリティ関数を提供します
package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var (
	// DatFilePattern は thXX.dat や thXXtr.dat ファイルのパターン
	DatFilePattern = regexp.MustCompile(`^th\d+(?:tr)?\.dat$`)
)

// FileExists はファイルが存在するか確認します
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// FromShiftJIS はShift-JISからUTF-8に変換します
func FromShiftJIS(str string) (string, error) {
	reader := strings.NewReader(str)
	transformer := japanese.ShiftJIS.NewDecoder()
	ret, err := io.ReadAll(transform.NewReader(reader, transformer))
	if err != nil {
		return "", err
	}
	return string(ret), nil
}

// SaveToFileWithBOM はUTF-8 BOMありでファイルに保存します
func SaveToFileWithBOM(outputPath string, content string) error {
	// 出力先ディレクトリを作成（存在しない場合）
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("%w: %w", ErrCreateDirectory, err)
	}

	// ファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateFile, err)
	}
	defer file.Close()

	// UTF-8 BOMを書き込む
	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	if _, err := file.Write(utf8bom); err != nil {
		return fmt.Errorf("%w: %w", ErrWriteBOM, err)
	}

	// 内容を書き込む
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("%w: %w", ErrWriteContent, err)
	}

	return nil
}

// GenerateOutputFilename は入力ファイル名から出力ファイル名を生成します
func GenerateOutputFilename(inputPath string) string {
	// ファイル名の部分だけを取得（拡張子なし）
	baseName := filepath.Base(inputPath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// titles_XXX.txt 形式の名前を生成
	return fmt.Sprintf("titles_%s.txt", baseName)
}

// ExtractGameNumber はファイル名からゲーム番号を抽出します
func ExtractGameNumber(filename string) int {
	baseName := strings.ToLower(filepath.Base(filename))
	gameNum := -1

	// thで始まるdatファイルなら、ゲーム番号を抽出
	if strings.HasPrefix(baseName, "th") {
		re := regexp.MustCompile(`^th(\d+)`)
		matches := re.FindStringSubmatch(baseName)
		if len(matches) > 1 {
			gameNum = 0
			for _, c := range matches[1] {
				gameNum = gameNum*10 + int(c-'0')
			}
		}
	}

	return gameNum
}

// IsTrialVersion はファイル名から体験版かどうかを判定します
func IsTrialVersion(filename string) bool {
	return strings.Contains(strings.ToLower(filepath.Base(filename)), "tr")
}

// DatFileFinder は.datファイルの検索を行います
type DatFileFinder struct{}

// NewDatFileFinder は新しいDatFileFinderを作成します
func NewDatFileFinder() *DatFileFinder {
	return &DatFileFinder{}
}

// Find は実行ファイルと同じディレクトリおよびカレントディレクトリから.datファイルを検索します
func (f *DatFileFinder) Find() (string, error) {
	var datFiles []string

	// カレントディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGetCurrentDirectory, err)
	}

	// まずカレントディレクトリを検索
	currentDirFiles, err := f.findInDir(currentDir)
	if err != nil {
		return "", err
	}
	datFiles = append(datFiles, currentDirFiles...)

	// カレントディレクトリで見つかった場合は他のディレクトリは検索しない
	if len(datFiles) > 0 {
		// 一致するファイルが2つ以上ある場合
		if len(datFiles) > 1 {
			return "", f.createMultipleFilesError(datFiles)
		}
		// 1つだけ見つかった場合はそのファイルパスを返す
		return datFiles[0], nil
	}

	// 実行ファイルのパスを取得
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrGetExecutablePath, err)
	}

	// 実行ファイルのディレクトリを取得
	execDir := filepath.Dir(execPath)

	// 実行ファイルのディレクトリを検索
	execDirFiles, err := f.findInDir(execDir)
	if err != nil {
		return "", err
	}
	datFiles = append(datFiles, execDirFiles...)

	// 一致するファイルがない場合
	if len(datFiles) == 0 {
		return "", nil
	}

	// 一致するファイルが2つ以上ある場合
	if len(datFiles) > 1 {
		return "", f.createMultipleFilesError(datFiles)
	}

	// 1つだけ見つかった場合はそのファイルパスを返す
	return datFiles[0], nil
}

// findInDir は指定されたディレクトリ内のthxx.datやthxxtr.datファイルを検索します
func (f *DatFileFinder) findInDir(dir string) ([]string, error) {
	var datFiles []string

	// ディレクトリ内のファイル一覧を取得
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrReadDirectory, dir, err)
	}

	// thxx.dat や thxxtr.dat パターンに一致するファイルを検索
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
func (f *DatFileFinder) createMultipleFilesError(datFiles []string) error {
	fileNames := make([]string, len(datFiles))
	for i, path := range datFiles {
		fileNames[i] = filepath.Base(path)
	}
	return fmt.Errorf("%w: %s", ErrMultipleDatFiles, strings.Join(fileNames, ", "))
}
