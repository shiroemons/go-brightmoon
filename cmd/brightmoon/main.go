package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/shiroemons/go-brightmoon/pkg/pbgarc"
)

var (
	extractFlag  = flag.Bool("x", false, "extract files")
	listFlag     = flag.Bool("l", false, "list files")
	outputDir    = flag.String("o", ".", "output directory")
	useType      = flag.Int("t", -1, "archive type (e.g., 0 for Imperishable Night, see README for details). If omitted, auto-detection is attempted.")
	debugFlag    = flag.Bool("d", false, "debug mode (show more info)")
	parallelFlag = flag.Bool("p", false, "use parallel extraction")
	workerCount  = flag.Int("w", 4, "number of worker threads for parallel extraction")
)

// コールバック関数
func callback(msg string, user interface{}) bool {
	fmt.Print(msg)
	return true
}

func main() {
	flag.Parse()

	// 引数チェック
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("使用方法: brightmoon [オプション] <アーカイブファイル>")
		fmt.Println("オプション:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// ファイル名
	filename := args[0]

	// デバッグモードの場合、ファイル情報を表示
	if *debugFlag {
		fileInfo, err := os.Stat(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ファイル情報の取得に失敗: %v\n", err)
		} else {
			fmt.Printf("ファイル: %s\n", filename)
			fmt.Printf("サイズ: %d バイト\n", fileInfo.Size())
			fmt.Printf("更新時間: %v\n", fileInfo.ModTime())

			// ファイルの先頭数バイトを表示
			file, err := os.Open(filename)
			if err == nil {
				defer file.Close()
				header := make([]byte, 16)
				n, err := file.Read(header)
				if err == nil && n > 0 {
					fmt.Printf("ファイルヘッダ (hex): ")
					for i := 0; i < n; i++ {
						fmt.Printf("%02x ", header[i])
					}
					fmt.Println()
				}
			}
		}
		fmt.Println()
	}

	var archive pbgarc.PBGArchive
	var err error

	if *useType != -1 {
		// タイプが指定されている場合
		archive, err = openSpecificArchive(filename, *useType)
	} else {
		// タイプが指定されていない場合 (自動判別 + ユーザー選択)
		archive, err = openArchiveAuto(filename)
	}

	// アーカイブを開く (エラーチェックは共通)
	if err != nil {
		if *debugFlag {
			fmt.Fprintf(os.Stderr, "エラー詳細:\n%v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "エラー: %v\n", err) // エラーメッセージを具体的に表示
		}
		os.Exit(1)
	}

	// KaguyaArchiveの場合はタイプを設定 (openArchiveAuto内で処理されるため、ここでは不要になる)
	/*
		if kaguyaArchive, ok := archive.(*pbgarc.KaguyaArchive); ok {
			// openArchiveAuto で設定済みのはず
			// kaguyaArchive.SetArchiveType(*useType)
			// fmt.Printf("アーカイブタイプを %d に設定しました\n", *useType)
		}
	*/

	// リストを表示する
	if *listFlag {
		listArchive(archive)
	}

	// 抽出対象ファイル名を取得 (アーカイブファイル名の後の引数)
	filesToExtract := []string{}
	if len(args) > 1 {
		filesToExtract = args[1:]
	}

	// 抽出する (-x フラグまたはファイル指定がある場合)
	if *extractFlag || len(filesToExtract) > 0 {
		if len(filesToExtract) > 0 {
			fmt.Printf("%d 個の指定されたファイルを抽出中...\n", len(filesToExtract))
		} else {
			fmt.Println("アーカイブ内の全ファイルを抽出中...")
		}

		var count int
		var notFound []string
		var extractErr error

		if *parallelFlag {
			// 並列処理で抽出
			count, notFound, extractErr = extractArchiveParallel(archive, *outputDir, *workerCount, filesToExtract)
		} else {
			// 順次処理で抽出
			count, notFound, extractErr = extractArchiveSequential(archive, *outputDir, filesToExtract)
		}

		if extractErr != nil {
			// エラーメッセージは抽出関数内で表示される想定だが、ここでも表示
			fmt.Fprintf(os.Stderr, "抽出処理中にエラーが発生しました: %v\n", extractErr)
		}

		if len(notFound) > 0 {
			fmt.Fprintf(os.Stderr, "\n警告: 指定されたファイルのうち、以下は見つかりませんでした:\n")
			for _, f := range notFound {
				fmt.Fprintf(os.Stderr, "- %s\n", f)
			}
		}

		if extractErr == nil || count > 0 { // エラーがあっても一部成功していれば表示
			fmt.Printf("\n%d 個のファイルを抽出しました\n", count)
		}
		// エラーがあり、かつ何も抽出できなかった場合は os.Exit(1) した方が良いかもしれない
		if extractErr != nil && count == 0 {
			os.Exit(1)
		}
	}
}

// 指定されたタイプのアーカイブを開くヘルパー関数 (実装を拡充)
func openSpecificArchive(filename string, archiveType int) (pbgarc.PBGArchive, error) {
	var targetArchive pbgarc.PBGArchive // インターフェース型
	var targetName string
	// var requiresSubTypeSelection bool = false // 未使用のため削除
	subType := -1 // Kaguya/Kanako のサブタイプ

	// アーカイブタイプと生成関数のマッピング (具体的な型を返すように変更)
	typeMapping := []struct {
		name      string
		newFunc   interface{} // 型を interface{} にして後でアサーション
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

	// 指定された archiveType から適切なアーカイブを探す
	found := false
	for _, mapping := range typeMapping {
		if mapping.needsType {
			if mapping.baseType == 1 { // Kaguya
				if archiveType == 0 || archiveType == 1 {
					if newFunc, ok := mapping.newFunc.(func() *pbgarc.KaguyaArchive); ok {
						targetArchive = newFunc() // 具体的な型で生成しインターフェースに代入
						targetName = mapping.name
						subType = archiveType
						found = true
						break
					}
				}
			} else if mapping.baseType == 2 { // Kanako
				if archiveType >= 0 && archiveType <= 2 { // Kanakoのサブタイプ範囲
					if newFunc, ok := mapping.newFunc.(func() *pbgarc.KanakoArchive); ok {
						targetArchive = newFunc()
						targetName = mapping.name
						subType = archiveType
						found = true
						break
					}
				}
			}
		} else {
			// タイプ指定不要な形式 - ここでは何もしない
		}
	}

	// `-t` でタイプ指定不要な形式 (Yumemi等) を指定した場合の処理
	// 例: `-t 3` など -> default でエラーになる想定だが、もし番号を割り振るならここで処理
	if !found && archiveType >= 0 {
		// ここで Yumemi 等の固定タイプを開く処理を追加することも可能
		// 例: if archiveType == 3 { targetArchive = pbgarc.NewYumemiArchive(); ... }
		return nil, fmt.Errorf("指定されたアーカイブタイプ %d は不明か、タイプ指定不要な形式です", archiveType)
	}

	if targetArchive == nil {
		return nil, fmt.Errorf("指定されたアーカイブタイプ %d に対応する実装が見つかりません", archiveType)
	}

	// サブタイプを設定 (Kaguya/Kanako)
	if kaguyaArchive, ok := targetArchive.(*pbgarc.KaguyaArchive); ok && subType != -1 {
		kaguyaArchive.SetArchiveType(subType)
		targetName = fmt.Sprintf("%s (Type %d)", targetName, subType)
	} else if kanakoArchive, ok := targetArchive.(*pbgarc.KanakoArchive); ok && subType != -1 {
		kanakoArchive.SetArchiveType(subType)
		options := pbgarc.GetArchiveTypeOptions()
		if subType >= 0 && subType < len(options) {
			targetName = fmt.Sprintf("%s (%s)", targetName, options[subType])
		} else {
			targetName = fmt.Sprintf("%s (Type %d)", targetName, subType)
		}
	}

	// ファイルを開く
	ok, err := targetArchive.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("%s としてアーカイブを開けませんでした: %w", targetName, err)
	}
	if !ok || !targetArchive.EnumFirst() {
		return nil, fmt.Errorf("%s としてアーカイブを開きましたが、無効か空のようです", targetName)
	}

	fmt.Printf("%s アーカイブを開きました: %s\n", targetName, filename)
	return targetArchive, nil
}

// guessArchiveInfoFromName はファイル名からアーカイブ形式とサブタイプを推測します
func guessArchiveInfoFromName(filename string) (expectedFormatName string, expectedSubType int, err error) {
	// デフォルトは未定義
	expectedSubType = -1

	// ファイル名のベース部分を取得 (例: th08bgm.dat -> th08)
	baseName := strings.ToLower(filepath.Base(filename))
	re := regexp.MustCompile(`^(th[0-9]{2,3}(?:[a-z]*)?)\..*$`) // 例: th08, th085, th125, th128
	matches := re.FindStringSubmatch(baseName)

	if len(matches) < 2 {
		// 特殊なケース (Suicaなど)
		if strings.Contains(baseName, "th06") { // Hinanawi (th06)
			expectedFormatName = "Hinanawi"
		} else if strings.Contains(baseName, "th07") { // Yumemi (th07)
			expectedFormatName = "Yumemi"
		} else if strings.Contains(baseName, "th095") { // Marisa (th095)
			expectedFormatName = "Marisa"
		} else {
			err = errors.New("ファイル名からゲームバージョンを特定できませんでした")
			return
		}
		return
	}

	gameStr := matches[1] // 例: "th08", "th085", "th13"

	switch gameStr {
	case "th06":
		expectedFormatName = "Hinanawi"
	case "th07":
		expectedFormatName = "Yumemi"
	case "th08":
		expectedFormatName = "Kaguya"
		expectedSubType = 0 // Imperishable Night
	case "th085": // 弾幕アマノジャクはファイル名パターンが異なる場合があるため注意
		expectedFormatName = "Kaguya"
		expectedSubType = 1 // Shoot the Bullet / Impossible Spell Card
	case "th095":
		expectedFormatName = "Marisa"
	case "th10", "th11":
		expectedFormatName = "Kanako"
		expectedSubType = 0 // MoF / SA
	case "th12", "th125", "th128":
		expectedFormatName = "Kanako"
		expectedSubType = 1 // UFO / DS / FW
	default:
		// th13 以降をKanako Type 2と仮定
		if strings.HasPrefix(gameStr, "th") {
			numStr := ""
			for i := 2; i < len(gameStr); i++ {
				if gameStr[i] >= '0' && gameStr[i] <= '9' {
					numStr += string(gameStr[i])
				} else {
					break // Stop at the first non-digit character
				}
			}
			if len(numStr) > 0 {
				num, atoiErr := strconv.Atoi(numStr)
				if atoiErr == nil && num >= 13 {
					expectedFormatName = "Kanako"
					expectedSubType = 2 // TD and later
				} else {
					// Handle cases where Atoi fails (shouldn't happen with the loop logic)
					// or num < 13 (which should have been caught by the switch)
					err = fmt.Errorf("未対応または不明なゲームバージョンです: %s (数値解析エラーまたは範囲外)", gameStr)
				}
			} else {
				err = fmt.Errorf("ファイル名からゲームバージョン番号を抽出できませんでした: %s", gameStr)
			}
		} else {
			// Should be caught by regex, but good practice
			err = fmt.Errorf("ファイル名が 'th' で始まりません: %s", gameStr)
		}
	}

	return
}

// アーカイブを開く (自動判別)
func openArchiveAuto(filename string) (pbgarc.PBGArchive, error) {
	// 各アーカイブタイプを試す
	archiveMappings := []struct {
		name      string
		newFunc   interface{} // 型を interface{} に
		needsType bool
		baseType  int
	}{
		{"Yumemi", pbgarc.NewYumemiArchive, false, 0},
		{"Suica", pbgarc.NewSuicaArchive, false, 0},
		{"Hinanawi", pbgarc.NewHinanawiArchive, false, 0},
		{"Marisa", pbgarc.NewMarisaArchive, false, 0},
		{"Kaguya", pbgarc.NewKaguyaArchive, true, 1},
		{"Kanako", pbgarc.NewKanakoArchive, true, 2},
	}

	// 候補リストの型も変更
	candidates := []struct {
		name    string
		archive pbgarc.PBGArchive // インターフェース型で保持
		mapping *struct {         // mapping情報も保持
			name      string
			newFunc   interface{}
			needsType bool
			baseType  int
		}
	}{}
	var errorsDetected []string

	fmt.Println("アーカイブ形式を自動検出中...")
	for i := range archiveMappings {
		mapping := &archiveMappings[i]
		var archive pbgarc.PBGArchive

		// newFunc の型に応じてインスタンス化
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
			// 予期しない型
			continue
		}

		ok, err := archive.Open(filename)

		if err != nil {
			errorsDetected = append(errorsDetected, fmt.Sprintf("- %s (Open): %v", mapping.name, err))
			continue
		}
		if !ok {
			// Open returned false, but no error. Treat as non-candidate.
			errorsDetected = append(errorsDetected, fmt.Sprintf("- %s (Open): returned false without error", mapping.name))
			continue
		}

		// Open succeeded (ok=true), now check EnumFirst
		if archive.EnumFirst() {
			fmt.Printf("- %s: 候補として検出\n", mapping.name)
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
		} else {
			// EnumFirst failed, record this
			errorsDetected = append(errorsDetected, fmt.Sprintf("- %s: 開けましたが無効か空のようです (EnumFirst failed)", mapping.name))
		}
	}

	// ---- 自動選択ロジック ----
	if len(candidates) == 0 {
		errorMsg := "対応するアーカイブ形式が見つかりませんでした。"
		// Always show detailed errors if detection failed
		if len(errorsDetected) > 0 {
			errorMsg += "\n検出時のエラー詳細:\n" + strings.Join(errorsDetected, "\n")
		}
		return nil, errors.New(errorMsg)
	}

	var chosenArchive pbgarc.PBGArchive
	var chosenMapping *struct {
		name      string
		newFunc   interface{}
		needsType bool
		baseType  int
	}

	guessedFormat, guessedSubType, guessErr := guessArchiveInfoFromName(filename)

	if len(candidates) == 1 {
		fmt.Printf("形式 %s を検出しました。\n", candidates[0].name)
		chosenArchive = candidates[0].archive
		chosenMapping = candidates[0].mapping

		// 候補が一つでも、推測と異なる場合は警告 (デバッグ用)
		if guessErr == nil && chosenMapping.name != guessedFormat {
			fmt.Printf("警告: 検出された形式 (%s) はファイル名から推測される形式 (%s) と異なります。\n", chosenMapping.name, guessedFormat)
		} else if guessErr != nil && *debugFlag {
			fmt.Printf("デバッグ情報: ファイル名からの形式推測に失敗: %v\n", guessErr)
		}

	} else {
		// 複数の候補が見つかった場合、ファイル名から推測した形式を優先
		fmt.Println("\n複数の候補が見つかりました:")
		for _, c := range candidates {
			fmt.Printf("- %s\n", c.name)
		}

		if guessErr != nil {
			return nil, fmt.Errorf("複数の形式候補が見つかりましたが、ファイル名から形式を特定できませんでした: %w。 `-t` オプションで形式を明示的に指定してください", guessErr)
		}

		fmt.Printf("ファイル名から %s 形式と推測します...\n", guessedFormat)
		foundMatch := false
		for _, c := range candidates {
			if c.mapping.name == guessedFormat {
				chosenArchive = c.archive
				chosenMapping = c.mapping
				foundMatch = true
				fmt.Printf("%s を選択しました。\n", chosenMapping.name)
				break
			}
		}

		if !foundMatch {
			return nil, fmt.Errorf("複数の形式候補が見つかりましたが、ファイル名から推測された形式 (%s) が候補内にありません。 `-t` オプションで形式を明示的に指定してください", guessedFormat)
		}
	}

	// 選ばれた形式が Kaguya または Kanako の場合、サブタイプを自動設定
	if chosenMapping.needsType {
		if guessErr != nil || guessedSubType == -1 {
			// ファイル名からサブタイプを推測できなかった場合
			errMsg := "選択された形式はサブタイプ指定が必要ですが、ファイル名から自動特定できませんでした。"
			if guessErr != nil {
				errMsg += fmt.Sprintf(" (エラー: %v)", guessErr)
			}
			return nil, fmt.Errorf("%s `-t` オプションでタイプを明示的に指定してください", errMsg)
		}

		// サブタイプを設定
		if chosenMapping.baseType == 1 { // Kaguya
			if kaguyaArchive, ok := chosenArchive.(*pbgarc.KaguyaArchive); ok {
				kaguyaArchive.SetArchiveType(guessedSubType)
				fmt.Printf("Kaguya サブタイプを %d (ファイル名から自動設定) に設定しました。\n", guessedSubType)
			} else {
				return nil, errors.New("内部エラー: KaguyaArchive への型アサーションに失敗しました")
			}
		} else if chosenMapping.baseType == 2 { // Kanako
			if kanakoArchive, ok := chosenArchive.(*pbgarc.KanakoArchive); ok {
				options := pbgarc.GetArchiveTypeOptions()
				if guessedSubType >= 0 && guessedSubType < len(options) {
					kanakoArchive.SetArchiveType(guessedSubType)
					fmt.Printf("Kanako サブタイプを %d (%s) (ファイル名から自動設定) に設定しました。\n", guessedSubType, options[guessedSubType])
				} else {
					return nil, fmt.Errorf("内部エラー: ファイル名から推測された Kanako サブタイプ %d が無効です", guessedSubType)
				}
			} else {
				return nil, errors.New("内部エラー: KanakoArchive への型アサーションに失敗しました")
			}
		}
	}

	fmt.Printf("%s アーカイブとして開きました: %s\n", chosenMapping.name, filename) // 最終的な形式名を表示
	return chosenArchive, nil
}

// アーカイブのリストを表示
func listArchive(archive pbgarc.PBGArchive) {
	fmt.Println("アーカイブ内のファイル一覧:")
	fmt.Println("----------------------------")
	fmt.Printf("%-32s %10s %10s\n", "ファイル名", "元サイズ", "圧縮サイズ")
	fmt.Println("----------------------------")

	if !archive.EnumFirst() {
		fmt.Println("ファイルがありません")
		return
	}

	do := true
	for do {
		fmt.Printf("%-32s %10d %10d\n",
			archive.GetEntryName(),
			archive.GetOriginalSize(),
			archive.GetCompressedSize())
		do = archive.EnumNext()
	}
	fmt.Println("----------------------------")
}
