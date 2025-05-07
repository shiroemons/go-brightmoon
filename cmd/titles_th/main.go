//
// GOOS=windows GOARCH=amd64 go build -o titles_th.exe main.go
//

package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"github.com/shiroemons/go-brightmoon/pkg/pbgarc"
)

var (
	version     = "0.0.3"
	archivePath string
	useType     = flag.Int("t", -1, "archive type (e.g., 0 for Imperishable Night, see README for details). If omitted, auto-detection is attempted.")
	outputDir   = flag.String("o", ".", "output directory for the generated files")
	debugMode   bool
	showVersion bool

	// 正規表現パターンを事前コンパイル
	datFilePattern = regexp.MustCompile(`^th\d+(?:tr)?\.dat$`)

	// アーカイブタイプと生成関数のマッピング
	archiveTypeMapping = []struct {
		name      string
		newFunc   interface{}
		needsType bool
		baseType  int
	}{
		{"Yumemi", pbgarc.NewYumemiArchive, false, 0},
		{"Kaguya", pbgarc.NewKaguyaArchive, true, 1},
		{"Suica", pbgarc.NewSuicaArchive, false, 0},
		{"Hinanawi", pbgarc.NewHinanawiArchive, false, 0},
		{"Marisa", pbgarc.NewMarisaArchive, false, 0},
		{"Kanako", pbgarc.NewKanakoArchive, true, 2},
	}
)

// init はフラグの設定を行います
func init() {
	// archiveフラグとそのエイリアスを設定
	flag.StringVar(&archivePath, "archive", "", "path to .dat archive file (e.g. th08.dat)")
	flag.StringVar(&archivePath, "a", "", "path to .dat archive file (e.g. th08.dat) (shorthand)")

	// -d と -debug の両方でデバッグモードを有効にする
	flag.BoolVar(&debugMode, "debug", false, "enable debug output")
	flag.BoolVar(&debugMode, "d", false, "enable debug output (shorthand)")

	// バージョン表示フラグ
	flag.BoolVar(&showVersion, "version", false, "show version information")
	flag.BoolVar(&showVersion, "v", false, "show version information (shorthand)")
}

type Record struct {
	FileName string
	Start    string
	Intro    string
	Loop     string
	Length   string
}

type Track struct {
	FileName string
	Title    string
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func toHex(i uint32) string {
	return fmt.Sprintf("%08X", i)
}

func transformEncoding(rawReader io.Reader, trans transform.Transformer) (string, error) {
	ret, err := io.ReadAll(transform.NewReader(rawReader, trans))
	if err == nil {
		return string(ret), nil
	} else {
		return "", err
	}
}

func FromShiftJIS(str string) (string, error) {
	return transformEncoding(strings.NewReader(str), japanese.ShiftJIS.NewDecoder())
}

// debugPrintf はデバッグモードが有効な場合のみメッセージを表示します
func debugPrintf(format string, a ...interface{}) {
	if debugMode {
		fmt.Printf(format, a...)
	}
}

// extractGameNumber はファイル名からゲーム番号を抽出します
func extractGameNumber(filename string) int {
	baseName := strings.ToLower(filepath.Base(filename))
	gameNum := -1

	// thで始まるdatファイルなら、ゲーム番号を抽出
	if strings.HasPrefix(baseName, "th") {
		re := regexp.MustCompile(`^th(\d+)`)
		matches := re.FindStringSubmatch(baseName)
		if len(matches) > 1 {
			gameNum, _ = strconv.Atoi(matches[1])
			debugPrintf("ファイル名からゲーム番号 %d を抽出しました\n", gameNum)
		}
	}

	return gameNum
}

// openArchiveAndExtract は.datアーカイブから特定のファイルをメモリに展開します
func openArchiveAndExtract(archivePath string, archiveType int, targetFiles []string) (map[string][]byte, error) {
	results := make(map[string][]byte)

	// アーカイブを開く
	var archive pbgarc.PBGArchive
	var err error

	// ファイル名からゲーム番号を判断して、より直接的にアーカイブタイプを設定
	if archiveType == -1 && strings.HasSuffix(strings.ToLower(archivePath), ".dat") {
		gameNum := extractGameNumber(archivePath)
		if gameNum > 0 {
			// ゲーム番号に基づいてタイプを強制設定
			if gameNum >= 6 && gameNum <= 7 {
				// 6, 7: 特殊な形式を使用
				if gameNum == 6 {
					debugPrintf("Hinanawi形式を強制適用します\n")
					archive = pbgarc.NewHinanawiArchive()
					ok, openErr := archive.Open(archivePath)
					if !ok || openErr != nil {
						debugPrintf("Hinanawi形式でのオープンに失敗しました: %v\n", openErr)
					} else {
						debugPrintf("Hinanawi形式での強制オープンに成功しました\n")
						goto EXTRACT_FILES
					}
				} else if gameNum == 7 {
					debugPrintf("Yumemi形式を強制適用します\n")
					archive = pbgarc.NewYumemiArchive()
					ok, openErr := archive.Open(archivePath)
					if !ok || openErr != nil {
						debugPrintf("Yumemi形式でのオープンに失敗しました: %v\n", openErr)
					} else {
						debugPrintf("Yumemi形式での強制オープンに成功しました\n")
						goto EXTRACT_FILES
					}
				}
			} else if gameNum == 8 || gameNum == 9 {
				// 8, 9: Kaguya形式
				debugPrintf("Kaguya形式（タイプ %d）を強制適用します\n", gameNum-8)
				kaguya := pbgarc.NewKaguyaArchive()
				kaguya.SetArchiveType(gameNum - 8) // 8→0, 9→1
				ok, openErr := kaguya.Open(archivePath)
				if !ok || openErr != nil {
					debugPrintf("Kaguya形式でのオープンに失敗しました: %v\n", openErr)
				} else {
					archive = kaguya
					debugPrintf("Kaguya形式での強制オープンに成功しました\n")
					goto EXTRACT_FILES
				}
			} else if gameNum >= 10 {
				// 10以降: Kanako形式
				var typeNum int
				if gameNum >= 10 && gameNum <= 11 || gameNum == 95 {
					typeNum = 0
				} else if gameNum == 12 {
					typeNum = 1
				} else {
					typeNum = 2 // 13以降はタイプ2
				}

				debugPrintf("Kanako形式（タイプ %d）を強制適用します\n", typeNum)
				kanako := pbgarc.NewKanakoArchive()
				kanako.SetArchiveType(typeNum)
				ok, openErr := kanako.Open(archivePath)
				if !ok || openErr != nil {
					debugPrintf("Kanako形式でのオープンに失敗しました: %v\n", openErr)
				} else {
					archive = kanako
					debugPrintf("Kanako形式での強制オープンに成功しました\n")
					goto EXTRACT_FILES
				}
			}
		}
	}

	if archiveType != -1 {
		// タイプが指定されている場合
		archive, err = openSpecificArchive(archivePath, archiveType)
	} else {
		// タイプが指定されていない場合（自動判別）
		archive, err = openArchiveAuto(archivePath)
	}

	if err != nil {
		return nil, fmt.Errorf("アーカイブを開けませんでした: %w", err)
	}

EXTRACT_FILES:
	// ファイルを検索して展開
	findCount := 0
	if !archive.EnumFirst() {
		return nil, errors.New("アーカイブ内にファイルが見つかりません")
	}

	do := true
	for do {
		entryName := archive.GetEntryName()

		// 対象ファイルか確認
		for _, target := range targetFiles {
			if strings.EqualFold(entryName, target) {
				// ファイルをメモリに展開
				data, err := extractFileToMemory(archive)
				if err != nil {
					return results, fmt.Errorf("%sの展開に失敗しました: %w", entryName, err)
				}

				results[entryName] = data
				findCount++
				debugPrintf("ファイル %s をメモリに展開しました（%d バイト）\n", entryName, len(data))
				break
			}
		}

		// すべてのファイルが見つかったら終了
		if findCount == len(targetFiles) {
			break
		}

		do = archive.EnumNext()
	}

	// 見つからなかったファイルは特定の条件でのみ警告を表示
	// TR版ファイルが見つからない場合は警告しない
	return results, nil
}

// extractFileToMemory はアーカイブエントリの内容をメモリに展開します
func extractFileToMemory(archive pbgarc.PBGArchive) ([]byte, error) {
	origSize := archive.GetOriginalSize()
	if origSize == 0 {
		return nil, fmt.Errorf("ファイルサイズが0です")
	}

	// バッファに書き込むためのio.Writerを作成
	buf := bytes.NewBuffer(make([]byte, 0, origSize))

	// ファイルを展開
	success := archive.Extract(buf, nil, nil)
	if !success {
		return nil, fmt.Errorf("ファイルの展開に失敗しました")
	}

	// バッファからデータを取得
	return buf.Bytes(), nil
}

// openSpecificArchive は指定されたタイプのアーカイブを開きます
func openSpecificArchive(filename string, archiveType int) (pbgarc.PBGArchive, error) {
	var targetArchive pbgarc.PBGArchive
	var targetName string
	subType := -1

	// 指定されたarchiveTypeから適切なアーカイブを探す
	found := false
	for _, mapping := range archiveTypeMapping {
		if mapping.needsType {
			if mapping.baseType == 1 { // Kaguya
				if archiveType == 0 || archiveType == 1 {
					if newFunc, ok := mapping.newFunc.(func() *pbgarc.KaguyaArchive); ok {
						targetArchive = newFunc()
						targetName = mapping.name
						subType = archiveType
						found = true
						break
					}
				}
			} else if mapping.baseType == 2 { // Kanako
				if archiveType >= 0 && archiveType <= 2 {
					if newFunc, ok := mapping.newFunc.(func() *pbgarc.KanakoArchive); ok {
						targetArchive = newFunc()
						targetName = mapping.name
						subType = archiveType
						found = true
						break
					}
				}
			}
		}
	}

	if !found && archiveType >= 0 {
		return nil, fmt.Errorf("指定されたアーカイブタイプ %d は不明か、タイプ指定不要な形式です", archiveType)
	}

	if targetArchive == nil {
		return nil, fmt.Errorf("指定されたアーカイブタイプ %d に対応する実装が見つかりません", archiveType)
	}

	// サブタイプを設定 (Kaguya/Kanako)
	if kaguyaArchive, ok := targetArchive.(*pbgarc.KaguyaArchive); ok && subType != -1 {
		kaguyaArchive.SetArchiveType(subType)
	} else if kanakoArchive, ok := targetArchive.(*pbgarc.KanakoArchive); ok && subType != -1 {
		kanakoArchive.SetArchiveType(subType)
	}

	// ファイルを開く
	ok, err := targetArchive.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("%s としてアーカイブを開けませんでした: %w", targetName, err)
	}
	if !ok || !targetArchive.EnumFirst() {
		return nil, fmt.Errorf("%s としてアーカイブを開きましたが、無効か空のようです", targetName)
	}

	return targetArchive, nil
}

// openArchiveAuto はアーカイブ形式を自動判別してアーカイブを開きます
func openArchiveAuto(filename string) (pbgarc.PBGArchive, error) {
	candidates := []struct {
		name    string
		archive pbgarc.PBGArchive
		mapping *struct {
			name      string
			newFunc   interface{}
			needsType bool
			baseType  int
		}
	}{}

	debugPrintf("アーカイブ形式を自動検出中...\n")
	for i := range archiveTypeMapping {
		mapping := &archiveTypeMapping[i]
		var archive pbgarc.PBGArchive

		// newFuncの型に応じてインスタンス化
		switch fn := mapping.newFunc.(type) {
		case func() *pbgarc.YumemiArchive:
			archive = fn()
		case func() *pbgarc.SuicaArchive:
			archive = fn()
		case func() *pbgarc.HinanawiArchive:
			archive = fn()
		case func() *pbgarc.MarisaArchive:
			archive = fn()
		case func() *pbgarc.KaguyaArchive:
			archive = fn()
		case func() *pbgarc.KanakoArchive:
			archive = fn()
		default:
			continue
		}

		ok, err := archive.Open(filename)
		if err != nil || !ok {
			continue
		}

		if archive.EnumFirst() {
			debugPrintf("- %s: 候補として検出\n", mapping.name)
			candidates = append(candidates, struct {
				name    string
				archive pbgarc.PBGArchive
				mapping *struct {
					name      string
					newFunc   interface{}
					needsType bool
					baseType  int
				}
			}{mapping.name, archive, mapping})
		}
	}

	if len(candidates) == 0 {
		return nil, errors.New("対応するアーカイブ形式が見つかりませんでした")
	}

	// ファイル名からタイプを推測
	gameNum := extractGameNumber(filename)

	var chosenArchive pbgarc.PBGArchive
	var archiveType int = -1    // 選択されたアーカイブのタイプ
	var archiveName string = "" // 選択されたアーカイブの名前

	// 単一の候補ならそれを使用
	if len(candidates) == 1 {
		debugPrintf("形式 %s を検出しました\n", candidates[0].name)
		chosenArchive = candidates[0].archive
		archiveName = candidates[0].name
	} else {
		// 複数候補がある場合、ファイル名から推測
		if gameNum >= 6 && gameNum <= 7 {
			// th06, th07
			for _, c := range candidates {
				if (gameNum == 6 && c.name == "Hinanawi") ||
					(gameNum == 7 && c.name == "Yumemi") {
					chosenArchive = c.archive
					archiveName = c.name
					break
				}
			}
		} else if gameNum == 8 || gameNum == 9 {
			// th08, th09 (Kaguya)
			for _, c := range candidates {
				if c.name == "Kaguya" {
					chosenArchive = c.archive
					archiveName = c.name

					// サブタイプを設定
					if kaguyaArchive, ok := chosenArchive.(*pbgarc.KaguyaArchive); ok {
						if gameNum == 8 {
							kaguyaArchive.SetArchiveType(0) // Imperishable Night
							archiveType = 0
							debugPrintf("Kaguya サブタイプを 0 に設定しました\n")
						} else {
							kaguyaArchive.SetArchiveType(1) // Shoot the Bullet
							archiveType = 1
							debugPrintf("Kaguya サブタイプを 1 に設定しました\n")
						}
					}
					break
				}
			}
		} else if gameNum >= 10 {
			// th10+ (Kanako)
			for _, c := range candidates {
				if c.name == "Kanako" {
					chosenArchive = c.archive
					archiveName = c.name

					// サブタイプを設定
					if kanakoArchive, ok := chosenArchive.(*pbgarc.KanakoArchive); ok {
						if gameNum >= 10 && gameNum <= 11 {
							kanakoArchive.SetArchiveType(0) // MoF/SA
							archiveType = 0
							debugPrintf("Kanako サブタイプを 0 に設定しました\n")
						} else if gameNum == 12 {
							kanakoArchive.SetArchiveType(1) // UFO/DS/FW
							archiveType = 1
							debugPrintf("Kanako サブタイプを 1 に設定しました\n")
						} else if gameNum >= 13 {
							// 東方神霊廟以降はタイプ2を使用
							kanakoArchive.SetArchiveType(2) // TD+
							archiveType = 2
							debugPrintf("Kanako サブタイプを 2 に設定しました（TH13以降）\n")
						}
					}
					break
				}
			}
		}

		// 選択できなかった場合は最初の候補を使用
		if chosenArchive == nil {
			debugPrintf("ゲーム番号 %d に基づく自動判別ができませんでした。最初の候補を使用します。\n", gameNum)
			chosenArchive = candidates[0].archive
			archiveName = candidates[0].name
		} else {
			debugPrintf("ゲーム番号 %d に基づいてアーカイブ形式を選択しました\n", gameNum)
		}
	}

	// 重要: th20tr.datなどの新しいファイルでは、Kanakoアーカイブのタイプを明示的に再設定
	// アーカイブタイプが設定されている場合は、一度閉じて再度開き直す
	if archiveType >= 0 && archiveName != "" {
		debugPrintf("自動判別の結果: %s (Type %d)\n", archiveName, archiveType)

		// タイプ設定後に問題が発生する場合は、明示的に再オープン
		if archiveName == "Kanako" && gameNum >= 13 {
			debugPrintf("Kanakoアーカイブを再初期化します（タイプ2を適用）\n")
			// 新しいインスタンスを作成
			newArchive := pbgarc.NewKanakoArchive()
			newArchive.SetArchiveType(2) // 明示的にタイプ2を設定

			// 再度開く
			ok, err := newArchive.Open(filename)
			if err != nil || !ok || !newArchive.EnumFirst() {
				debugPrintf("再初期化に失敗しました: %v\n", err)
				// 元のアーカイブを返す
			} else {
				debugPrintf("再初期化に成功しました\n")
				return newArchive, nil // 成功したら新しいアーカイブを返す
			}
		}
	}

	return chosenArchive, nil
}

// createMultipleDatFilesError は複数の.datファイルが見つかった場合のエラーを生成します
func createMultipleDatFilesError(datFiles []string) error {
	fileNames := make([]string, len(datFiles))
	for i, path := range datFiles {
		fileNames[i] = filepath.Base(path)
	}
	return fmt.Errorf("複数の.datファイルが見つかりました: %s。-archive フラグで使用するファイルを指定してください", strings.Join(fileNames, ", "))
}

// findDatFiles は実行ファイルと同じディレクトリおよびカレントディレクトリにある thxx.dat や thxxtr.dat ファイルを検索します
func findDatFiles() (string, error) {
	var datFiles []string

	// カレントディレクトリを取得
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("カレントディレクトリを取得できませんでした: %w", err)
	}

	// まずカレントディレクトリを検索
	currentDirFiles, err := findDatFilesInDir(currentDir)
	if err != nil {
		return "", err
	}
	datFiles = append(datFiles, currentDirFiles...)

	// カレントディレクトリで見つかった場合は他のディレクトリは検索しない
	if len(datFiles) > 0 {
		// 一致するファイルが2つ以上ある場合
		if len(datFiles) > 1 {
			return "", createMultipleDatFilesError(datFiles)
		}

		// 1つだけ見つかった場合はそのファイルパスを返す
		return datFiles[0], nil
	}

	// 実行ファイルのパスを取得
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("実行ファイルのパスを取得できませんでした: %w", err)
	}

	// 実行ファイルのディレクトリを取得
	execDir := filepath.Dir(execPath)

	// 実行ファイルのディレクトリを検索
	execDirFiles, err := findDatFilesInDir(execDir)
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
		return "", createMultipleDatFilesError(datFiles)
	}

	// 1つだけ見つかった場合はそのファイルパスを返す
	return datFiles[0], nil
}

// findDatFilesInDir は指定されたディレクトリ内のthxx.datやthxxtr.datファイルを検索します
func findDatFilesInDir(dir string) ([]string, error) {
	var datFiles []string

	// ディレクトリ内のファイル一覧を取得
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("ディレクトリ '%s' 内のファイル一覧を取得できませんでした: %w", dir, err)
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

		if datFilePattern.MatchString(name) {
			datFiles = append(datFiles, filepath.Join(dir, name))
		}
	}

	return datFiles, nil
}

// 結果をUTF-8 BOMありでファイルに保存する関数
func saveToFile(outputPath string, content string) error {
	// 出力先ディレクトリを作成（存在しない場合）
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("出力先ディレクトリの作成に失敗しました: %w", err)
	}

	// ファイルを作成
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました: %w", err)
	}
	defer file.Close()

	// UTF-8 BOMを書き込む
	utf8bom := []byte{0xEF, 0xBB, 0xBF}
	if _, err := file.Write(utf8bom); err != nil {
		return fmt.Errorf("bomの書き込みに失敗しました: %w", err)
	}

	// 内容を書き込む
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("内容の書き込みに失敗しました: %w", err)
	}

	debugPrintf("ファイル %s を作成しました\n", outputPath)
	return nil
}

// 入力ファイル名から出力ファイル名を生成する
func generateOutputFilename(inputPath string) string {
	// ファイル名の部分だけを取得（拡張子なし）
	baseName := filepath.Base(inputPath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// titles_XXX.txt 形式の名前を生成
	return fmt.Sprintf("titles_%s.txt", baseName)
}

// AdditionalInfo は補足情報を保持する構造体
type AdditionalInfo struct {
	HasAdditionalInfo bool
	TitleInfo         string
	DisplayTitle      string
	IsTrialVersion    bool
	Error             error
}

// isTrialVersion はファイル名から体験版かどうかを判定します
func isTrialVersion(filename string) bool {
	return strings.Contains(strings.ToLower(filepath.Base(filename)), "tr")
}

// processArchive はアーカイブからファイルを抽出して処理する共通関数です
func processArchive(archivePath string, archiveType int) ([]byte, string, string, error) {
	// アーカイブが体験版かどうか判定
	isTrial := isTrialVersion(archivePath)

	// 検索するファイル名（体験版アーカイブの場合は _trファイルのみを検索）
	var targetFiles []string
	var fmtFile, cmtFile string
	if isTrial {
		targetFiles = []string{"thbgm_tr.fmt", "musiccmt_tr.txt"}
		fmtFile = "thbgm_tr.fmt"
		cmtFile = "musiccmt_tr.txt"
	} else {
		targetFiles = []string{"thbgm.fmt", "musiccmt.txt"}
		fmtFile = "thbgm.fmt"
		cmtFile = "musiccmt.txt"
	}

	// アーカイブからファイルを抽出
	fileData, err := openArchiveAndExtract(archivePath, archiveType, targetFiles)
	if err != nil {
		return nil, "", "", fmt.Errorf("アーカイブからのファイル抽出中にエラーが発生しました: %w", err)
	}

	// データの取得
	var thfmt []byte
	var musiccmt string

	if data, ok := fileData[fmtFile]; ok && len(data) > 0 {
		thfmt = data
		if cmtData, ok := fileData[cmtFile]; ok {
			musiccmt = string(cmtData)
		} else {
			return nil, "", "", fmt.Errorf("警告: %s が見つかりませんでした", cmtFile)
		}
	} else {
		return nil, "", "", fmt.Errorf("アーカイブ内に %s が見つかりませんでした", fmtFile)
	}

	return thfmt, musiccmt, archivePath, nil
}

// processLocalFiles はローカルファイルシステムからファイルを読み込む共通関数です
func processLocalFiles() ([]byte, string, string, error) {
	var thfmt []byte
	var musiccmt string
	var inputFile string
	var err error

	// 製品版のファイルをチェック
	if fileExists("thbgm.fmt") && fileExists("musiccmt.txt") {
		inputFile = "thbgm" // 特別な名前を使用
		thfmt, err = os.ReadFile("thbgm.fmt")
		if err != nil {
			return nil, "", "", fmt.Errorf("thbgm.fmtの読み込みに失敗しました: %w", err)
		}

		musiccmtBytes, err := os.ReadFile("musiccmt.txt")
		if err != nil {
			return nil, "", "", fmt.Errorf("musiccmt.txtの読み込みに失敗しました: %w", err)
		}
		musiccmt = string(musiccmtBytes)
		return thfmt, musiccmt, inputFile, nil
	}

	// 体験版のファイルをチェック
	if fileExists("thbgm_tr.fmt") && fileExists("musiccmt_tr.txt") {
		inputFile = "thbgm_tr" // 特別な名前を使用
		thfmt, err = os.ReadFile("thbgm_tr.fmt")
		if err != nil {
			return nil, "", "", fmt.Errorf("thbgm_tr.fmtの読み込みに失敗しました: %w", err)
		}

		musiccmtBytes, err := os.ReadFile("musiccmt_tr.txt")
		if err != nil {
			return nil, "", "", fmt.Errorf("musiccmt_tr.txtの読み込みに失敗しました: %w", err)
		}
		musiccmt = string(musiccmtBytes)
		return thfmt, musiccmt, inputFile, nil
	}

	return nil, "", "", fmt.Errorf("thbgm.fmt、musiccmt.txt または thbgm_tr.fmt、musiccmt_tr.txt のファイルがありません")
}

func main() {
	flag.Parse()

	// バージョン情報の表示
	if showVersion {
		fmt.Printf("titles_th version %s\n", version)
		os.Exit(0)
	}

	var (
		err       error
		thfmt     []byte
		musiccmt  string
		inputFile string // 入力ファイル名を保存するための変数
	)

	// アーカイブが指定されている場合
	if archivePath != "" {
		debugPrintf("アーカイブファイル %s からデータを読み込みます...\n", archivePath)

		// 補足情報の存在チェックとタイトル取得
		additionalInfo := checkAdditionalInfo(archivePath)
		if additionalInfo.Error != nil {
			fmt.Fprintf(os.Stderr, "警告: 補足情報の読み込みに失敗しました: %v\n", additionalInfo.Error)
		}

		// アーカイブからファイルを抽出して処理
		thfmt, musiccmt, inputFile, err = processArchive(archivePath, *useType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	} else {
		// コマンドラインでアーカイブが指定されていない場合は自動検出を試みる
		datFile, err := findDatFiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
			os.Exit(1)
		}

		if datFile != "" {
			// .datファイルが見つかった場合、そこからデータを抽出
			debugPrintf("自動検出したアーカイブファイル %s からデータを読み込みます...\n", filepath.Base(datFile))

			// 補足情報の存在チェックとタイトル取得
			additionalInfo := checkAdditionalInfo(datFile)
			if additionalInfo.Error != nil {
				fmt.Fprintf(os.Stderr, "警告: 補足情報の読み込みに失敗しました: %v\n", additionalInfo.Error)
			}

			// アーカイブからファイルを抽出して処理
			thfmt, musiccmt, inputFile, err = processArchive(datFile, *useType)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		} else {
			// 従来の処理: ファイルシステムからの読み込み
			thfmt, musiccmt, inputFile, err = processLocalFiles()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		}
	}

	offset := 0
	var records []*Record
	for offset < len(thfmt) {
		if offset+52 > len(thfmt) {
			break
		}
		pcmFmt := thfmt[offset : offset+52]
		fileBytes := pcmFmt[0:16]
		n := bytes.IndexByte(fileBytes, 0)
		file := string(fileBytes[:n])
		if file != "" {
			start := binary.LittleEndian.Uint32(pcmFmt[16:])
			intro := binary.LittleEndian.Uint32(pcmFmt[24:])
			length := binary.LittleEndian.Uint32(pcmFmt[28:])
			loop := length - intro
			record := &Record{
				FileName: file,
				Start:    toHex(start),
				Intro:    toHex(intro),
				Loop:     toHex(loop),
				Length:   toHex(length),
			}
			records = append(records, record)
		}
		offset += 52
	}

	text, err := FromShiftJIS(musiccmt)
	check(err)

	var (
		fileName string
		title    string
	)

	var tracks []*Track

	buf := bytes.NewBufferString(text)
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "@bgm/") {
			fileName = strings.Replace(scanner.Text(), "@bgm/", "", -1)
			if !strings.HasSuffix(fileName, ".wav") {
				fileName = fileName + ".wav"
			}
		}
		if strings.HasPrefix(scanner.Text(), "♪") {
			title = strings.Replace(scanner.Text(), "♪", "", -1)
			track := &Track{
				FileName: fileName,
				Title:    title,
			}
			tracks = append(tracks, track)
		}
	}

	// 曲データの出力内容を構築
	var outputBuilder strings.Builder

	// 補足情報の存在チェックとタイトル取得
	additionalInfo := checkAdditionalInfo(inputFile)
	if additionalInfo.Error != nil {
		fmt.Fprintf(os.Stderr, "警告: 補足情報の読み込みに失敗しました: %v\n", additionalInfo.Error)
	}

	// 補足情報が存在する場合のみタイトル情報を出力
	if additionalInfo.HasAdditionalInfo {
		if additionalInfo.IsTrialVersion {
			outputBuilder.WriteString(fmt.Sprintf("#「%s」体験版曲データ\n", additionalInfo.DisplayTitle))
		} else {
			outputBuilder.WriteString(fmt.Sprintf("#「%s」製品版曲データ\n", additionalInfo.DisplayTitle))
		}
		outputBuilder.WriteString("#デフォルトのパスと製品名\n")
		outputBuilder.WriteString(additionalInfo.TitleInfo + "\n")
	}

	// ヘッダー情報
	outputBuilder.WriteString("#曲データ\n")
	outputBuilder.WriteString("#開始位置[Bytes]、イントロ部の長さ[Bytes]、ループ部の長さ[Bytes]、曲名\n")
	outputBuilder.WriteString("#位置・長さは16進値として記述する\n")

	// トラックデータ
	for _, t := range tracks {
		for _, r := range records {
			if t.FileName == r.FileName {
				outputBuilder.WriteString(fmt.Sprintf("%s,%s,%s,%s\n", r.Start, r.Intro, r.Loop, t.Title))
			}
		}
	}
	for i, r := range records {
		if len(tracks) <= i {
			if r.FileName == "th128_08.wav" {
				outputBuilder.WriteString(fmt.Sprintf("%s,%s,%s,%s\n", r.Start, r.Intro, r.Loop, "プレイヤーズスコア"))
			} else {
				outputBuilder.WriteString(fmt.Sprintf("%s,%s,%s,%s\n", r.Start, r.Intro, r.Loop, r.FileName))
			}
		}
	}

	// 出力ファイル名を生成
	outputFilename := generateOutputFilename(inputFile)
	outputPath := filepath.Join(*outputDir, outputFilename)

	// ファイルに保存
	err = saveToFile(outputPath, outputBuilder.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "ファイルの保存に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// デバッグモードの場合のみ保存メッセージを表示
	debugPrintf("データを %s に保存しました\n", outputPath)

	// 標準出力にも表示
	fmt.Println(outputBuilder.String())
}

// checkAdditionalInfo は補足情報の存在をチェックします
func checkAdditionalInfo(archivePath string) AdditionalInfo {
	// アーカイブと同じディレクトリを取得
	dir := filepath.Dir(archivePath)

	// readme.txtのパス
	readmePath := filepath.Join(dir, "readme.txt")

	// thbgm.datとthbgm_tr.datのパス
	thbgmPath := filepath.Join(dir, "thbgm.dat")
	thbgmTrPath := filepath.Join(dir, "thbgm_tr.dat")

	// readme.txtの存在チェック
	if !fileExists(readmePath) {
		return AdditionalInfo{HasAdditionalInfo: false}
	}

	// thbgm.datまたはthbgm_tr.datの存在チェック
	if !fileExists(thbgmPath) && !fileExists(thbgmTrPath) {
		return AdditionalInfo{HasAdditionalInfo: false}
	}

	// readme.txtを読み込む
	readmeData, err := os.ReadFile(readmePath)
	if err != nil {
		return AdditionalInfo{Error: fmt.Errorf("readme.txtの読み込みに失敗しました: %w", err)}
	}

	// ShiftJISからUTF-8に変換
	readmeText, err := FromShiftJIS(string(readmeData))
	if err != nil {
		return AdditionalInfo{Error: fmt.Errorf("readme.txtの文字コード変換に失敗しました: %w", err)}
	}

	// 2行目を取得
	lines := strings.Split(readmeText, "\n")
	if len(lines) < 2 {
		return AdditionalInfo{HasAdditionalInfo: false}
	}

	// 2行目の最初の文字が○かチェック
	secondLine := strings.TrimSpace(lines[1])
	if !strings.HasPrefix(secondLine, "○") {
		return AdditionalInfo{HasAdditionalInfo: false}
	}

	// ○以降の文字列を取得
	title := strings.TrimPrefix(secondLine, "○")

	// 使用するthbgm.datのパスを決定
	var thbgmFilePath string
	if fileExists(thbgmPath) {
		thbgmFilePath = thbgmPath
	} else {
		thbgmFilePath = thbgmTrPath
	}

	// 体験版かどうかチェック
	isTrialVer := strings.Contains(title, " 体験版")

	// 体験版の場合は「 体験版」の表記を除外
	displayTitle := title
	if isTrialVer {
		displayTitle = strings.Replace(title, " 体験版", "", 1)
	}

	return AdditionalInfo{
		HasAdditionalInfo: true,
		TitleInfo:         fmt.Sprintf("@%s,%s", thbgmFilePath, title),
		DisplayTitle:      displayTitle,
		IsTrialVersion:    isTrialVer,
	}
}
