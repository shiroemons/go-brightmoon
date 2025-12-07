package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shiroemons/go-brightmoon/internal/titles/fileutil"
	"github.com/shiroemons/go-brightmoon/internal/titles/models"
)

// AdditionalInfoParser は補足情報を解析します
type AdditionalInfoParser struct{}

// NewAdditionalInfoParser は新しいAdditionalInfoParserを作成します
func NewAdditionalInfoParser() *AdditionalInfoParser {
	return &AdditionalInfoParser{}
}

// CheckAdditionalInfo は補足情報の存在をチェックします
func (p *AdditionalInfoParser) CheckAdditionalInfo(archivePath string) models.AdditionalInfo {
	// アーカイブと同じディレクトリを取得
	dir := filepath.Dir(archivePath)

	// readme.txtのパス
	readmePath := filepath.Join(dir, "readme.txt")

	// thbgm.datとthbgm_tr.datのパス
	thbgmPath := filepath.Join(dir, "thbgm.dat")
	thbgmTrPath := filepath.Join(dir, "thbgm_tr.dat")

	// readme.txtの存在チェック
	if !fileutil.FileExists(readmePath) {
		return models.AdditionalInfo{HasAdditionalInfo: false}
	}

	// thbgm.datまたはthbgm_tr.datの存在チェック
	if !fileutil.FileExists(thbgmPath) && !fileutil.FileExists(thbgmTrPath) {
		return models.AdditionalInfo{HasAdditionalInfo: false}
	}

	// readme.txtを読み込む
	readmeData, err := os.ReadFile(readmePath)
	if err != nil {
		return models.AdditionalInfo{Error: fmt.Errorf("%w: %w", ErrReadmeRead, err)}
	}

	// ShiftJISからUTF-8に変換
	readmeText, err := fileutil.FromShiftJIS(string(readmeData))
	if err != nil {
		return models.AdditionalInfo{Error: fmt.Errorf("%w: %w", ErrReadmeEncodingConversion, err)}
	}

	// 2行目を取得
	lines := strings.Split(readmeText, "\n")
	if len(lines) < 2 {
		return models.AdditionalInfo{HasAdditionalInfo: false}
	}

	// 2行目からタイトルを抽出
	secondLine := strings.TrimSpace(lines[1])

	var title string
	if strings.HasPrefix(secondLine, "○") {
		// TH10以降の形式: ○東方風神録　～ Mountain of Faith.
		title = strings.TrimPrefix(secondLine, "○")
	} else if strings.HasPrefix(secondLine, "東方") {
		// TH07形式: 　東方妖々夢　〜 Perfect Cherry Blossom.
		title = secondLine
	} else {
		return models.AdditionalInfo{HasAdditionalInfo: false}
	}

	// 使用するthbgm.datのパスを決定
	var thbgmFilePath string
	if fileutil.FileExists(thbgmPath) {
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

	return models.AdditionalInfo{
		HasAdditionalInfo: true,
		TitleInfo:         fmt.Sprintf("@%s,%s", thbgmFilePath, title),
		DisplayTitle:      displayTitle,
		IsTrialVersion:    isTrialVer,
	}
}
