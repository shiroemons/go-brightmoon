package archive

import (
	"errors"
	"fmt"

	"github.com/shiroemons/go-brightmoon/internal/titles/fileutil"
	"github.com/shiroemons/go-brightmoon/pkg/pbgarc"
)

// ArchiveTypeMapping はアーカイブタイプと生成関数のマッピング
type ArchiveTypeMapping struct {
	Name      string
	NewFunc   any
	NeedsType bool
	BaseType  int
}

// GetArchiveTypeMappings はアーカイブタイプのマッピングを返します
func GetArchiveTypeMappings() []ArchiveTypeMapping {
	return []ArchiveTypeMapping{
		{"Yumemi", pbgarc.NewYumemiArchive, false, 0},
		{"Kaguya", pbgarc.NewKaguyaArchive, true, 1},
		{"Suica", pbgarc.NewSuicaArchive, false, 0},
		{"Hinanawi", pbgarc.NewHinanawiArchive, false, 0},
		{"Marisa", pbgarc.NewMarisaArchive, false, 0},
		{"Kanako", pbgarc.NewKanakoArchive, true, 2},
	}
}

// openSpecificArchive は指定されたタイプのアーカイブを開きます
func (e *Extractor) openSpecificArchive(filename string, archiveType int) (pbgarc.PBGArchive, error) {
	var targetArchive pbgarc.PBGArchive
	var targetName string
	subType := -1

	// 指定されたarchiveTypeから適切なアーカイブを探す
	found := false
	for _, mapping := range GetArchiveTypeMappings() {
		if mapping.NeedsType {
			if mapping.BaseType == 1 { // Kaguya
				if archiveType == 0 || archiveType == 1 {
					if newFunc, ok := mapping.NewFunc.(func() *pbgarc.KaguyaArchive); ok {
						targetArchive = newFunc()
						targetName = mapping.Name
						subType = archiveType
						found = true
						break
					}
				}
			} else if mapping.BaseType == 2 { // Kanako
				if archiveType >= 0 && archiveType <= 2 {
					if newFunc, ok := mapping.NewFunc.(func() *pbgarc.KanakoArchive); ok {
						targetArchive = newFunc()
						targetName = mapping.Name
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

// archiveCandidate はアーカイブ候補
type archiveCandidate struct {
	name    string
	archive pbgarc.PBGArchive
	mapping *ArchiveTypeMapping
}

// openArchiveAuto はアーカイブ形式を自動判別してアーカイブを開きます
func (e *Extractor) openArchiveAuto(filename string) (pbgarc.PBGArchive, error) {
	candidates := []archiveCandidate{}

	e.logger.Printf("アーカイブ形式を自動検出中...\n")
	mappings := GetArchiveTypeMappings()
	for i := range mappings {
		mapping := &mappings[i]
		var archive pbgarc.PBGArchive

		// newFuncの型に応じてインスタンス化
		switch fn := mapping.NewFunc.(type) {
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
			e.logger.Printf("- %s: 候補として検出\n", mapping.Name)
			candidates = append(candidates, archiveCandidate{mapping.Name, archive, mapping})
		}
	}

	if len(candidates) == 0 {
		return nil, errors.New("対応するアーカイブ形式が見つかりませんでした")
	}

	// ファイル名からタイプを推測
	gameNum := fileutil.ExtractGameNumber(filename)

	var chosenArchive pbgarc.PBGArchive
	var archiveType int = -1    // 選択されたアーカイブのタイプ
	var archiveName string = "" // 選択されたアーカイブの名前

	// 単一の候補ならそれを使用
	if len(candidates) == 1 {
		e.logger.Printf("形式 %s を検出しました\n", candidates[0].name)
		chosenArchive = candidates[0].archive
		archiveName = candidates[0].name
	} else {
		// 複数候補がある場合、ファイル名から推測
		chosenArchive, archiveName, archiveType = e.chooseFromCandidates(candidates, gameNum)
		if chosenArchive == nil {
			e.logger.Printf("ゲーム番号 %d に基づく自動判別ができませんでした。最初の候補を使用します。\n", gameNum)
			chosenArchive = candidates[0].archive
			archiveName = candidates[0].name
		} else {
			e.logger.Printf("ゲーム番号 %d に基づいてアーカイブ形式を選択しました\n", gameNum)
		}
	}

	// 重要: th20tr.datなどの新しいファイルでは、Kanakoアーカイブのタイプを明示的に再設定
	if archiveType >= 0 && archiveName != "" {
		e.logger.Printf("自動判別の結果: %s (Type %d)\n", archiveName, archiveType)

		// タイプ設定後に問題が発生する場合は、明示的に再オープン
		if archiveName == "Kanako" && gameNum >= 13 {
			e.logger.Printf("Kanakoアーカイブを再初期化します（タイプ2を適用）\n")
			// 新しいインスタンスを作成
			newArchive := pbgarc.NewKanakoArchive()
			newArchive.SetArchiveType(2) // 明示的にタイプ2を設定

			// 再度開く
			ok, err := newArchive.Open(filename)
			if err != nil || !ok || !newArchive.EnumFirst() {
				e.logger.Printf("再初期化に失敗しました: %v\n", err)
				// 元のアーカイブを返す
			} else {
				e.logger.Printf("再初期化に成功しました\n")
				return newArchive, nil // 成功したら新しいアーカイブを返す
			}
		}
	}

	return chosenArchive, nil
}

// chooseFromCandidates は複数の候補から適切なアーカイブを選択します
func (e *Extractor) chooseFromCandidates(candidates []archiveCandidate, gameNum int) (pbgarc.PBGArchive, string, int) {
	var chosenArchive pbgarc.PBGArchive
	var archiveName string
	var archiveType int = -1

	switch {
	case gameNum >= 6 && gameNum <= 7:
		// th06, th07
		for _, c := range candidates {
			if (gameNum == 6 && c.name == "Hinanawi") ||
				(gameNum == 7 && c.name == "Yumemi") {
				chosenArchive = c.archive
				archiveName = c.name
				break
			}
		}

	case gameNum == 8 || gameNum == 9:
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
						e.logger.Printf("Kaguya サブタイプを 0 に設定しました\n")
					} else {
						kaguyaArchive.SetArchiveType(1) // Shoot the Bullet
						archiveType = 1
						e.logger.Printf("Kaguya サブタイプを 1 に設定しました\n")
					}
				}
				break
			}
		}

	case gameNum >= 10:
		// th10+ (Kanako)
		for _, c := range candidates {
			if c.name == "Kanako" {
				chosenArchive = c.archive
				archiveName = c.name

				// サブタイプを設定
				if kanakoArchive, ok := chosenArchive.(*pbgarc.KanakoArchive); ok {
					if gameNum >= 10 && gameNum <= 11 || gameNum == 95 {
						kanakoArchive.SetArchiveType(0) // MoF/SA
						archiveType = 0
						e.logger.Printf("Kanako サブタイプを 0 に設定しました\n")
					} else if gameNum == 12 {
						kanakoArchive.SetArchiveType(1) // UFO/DS/FW
						archiveType = 1
						e.logger.Printf("Kanako サブタイプを 1 に設定しました\n")
					} else if gameNum >= 13 {
						// 東方神霊廟以降はタイプ2を使用
						kanakoArchive.SetArchiveType(2) // TD+
						archiveType = 2
						e.logger.Printf("Kanako サブタイプを 2 に設定しました（TH13以降）\n")
					}
				}
				break
			}
		}
	}

	return chosenArchive, archiveName, archiveType
}
