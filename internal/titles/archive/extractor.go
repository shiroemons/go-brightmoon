// Package archive はアーカイブの操作を行います
package archive

import (
	"context"
	"fmt"
	"strings"

	"github.com/shiroemons/go-brightmoon/internal/titles/config"
	"github.com/shiroemons/go-brightmoon/internal/titles/fileutil"
	"github.com/shiroemons/go-brightmoon/pkg/pbgarc"
)

// Extractor はアーカイブからファイルを抽出します
type Extractor struct {
	logger          *config.DebugLogger
	factory         ArchiveFactory
	memoryExtractor MemoryExtractor
}

// NewExtractor は新しいExtractorを作成します
func NewExtractor(logger *config.DebugLogger) *Extractor {
	return &Extractor{
		logger:          logger,
		factory:         &DefaultArchiveFactory{},
		memoryExtractor: &DefaultMemoryExtractor{},
	}
}

// NewExtractorWithFactory は新しいExtractorをファクトリー付きで作成します
func NewExtractorWithFactory(logger *config.DebugLogger, factory ArchiveFactory, extractor MemoryExtractor) *Extractor {
	return &Extractor{
		logger:          logger,
		factory:         factory,
		memoryExtractor: extractor,
	}
}

// ExtractFiles は.datアーカイブから特定のファイルをメモリに展開します
func (e *Extractor) ExtractFiles(ctx context.Context, archivePath string, archiveType int, targetFiles []string) (map[string][]byte, error) {
	// コンテキストのキャンセルチェック
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	results := make(map[string][]byte)

	// アーカイブを開く
	archive, err := e.openArchive(archivePath, archiveType)
	if err != nil {
		return nil, err
	}

	// ファイルを検索して展開
	findCount := 0
	if !archive.EnumFirst() {
		return nil, ErrNoFilesFound
	}

	do := true
	for do {
		// コンテキストのキャンセルチェック
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		entryName := archive.GetEntryName()

		// 対象ファイルか確認
		for _, target := range targetFiles {
			if strings.EqualFold(entryName, target) {
				// ファイルをメモリに展開
				data, err := e.extractToMemory(archive)
				if err != nil {
					return results, fmt.Errorf("%w: %s: %w", ErrExtractFailed, entryName, err)
				}

				results[entryName] = data
				findCount++
				e.logger.Printf("ファイル %s をメモリに展開しました（%d バイト）\n", entryName, len(data))
				break
			}
		}

		// すべてのファイルが見つかったら終了
		if findCount == len(targetFiles) {
			break
		}

		do = archive.EnumNext()
	}

	return results, nil
}

// extractToMemory はアーカイブエントリの内容をメモリに展開します
func (e *Extractor) extractToMemory(archive pbgarc.PBGArchive) ([]byte, error) {
	return e.memoryExtractor.ExtractToMemory(archive)
}

// openArchive はアーカイブを開きます
func (e *Extractor) openArchive(archivePath string, archiveType int) (pbgarc.PBGArchive, error) {
	// ファイル名からゲーム番号を判断して、より直接的にアーカイブタイプを設定
	if archiveType == -1 && strings.HasSuffix(strings.ToLower(archivePath), ".dat") {
		gameNum := fileutil.ExtractGameNumber(archivePath)
		if gameNum > 0 {
			return e.openByGameNumber(archivePath, gameNum)
		}
	}

	if archiveType != -1 {
		// タイプが指定されている場合
		return e.openSpecificArchive(archivePath, archiveType)
	}

	// タイプが指定されていない場合（自動判別）
	return e.openArchiveAuto(archivePath)
}

// openByGameNumber はゲーム番号に基づいてアーカイブを開きます
func (e *Extractor) openByGameNumber(archivePath string, gameNum int) (pbgarc.PBGArchive, error) {
	switch {
	case gameNum == 6:
		e.logger.Printf("Hinanawi形式を強制適用します\n")
		archive := e.factory.NewHinanawiArchive()
		ok, err := archive.Open(archivePath)
		if !ok || err != nil {
			e.logger.Printf("Hinanawi形式でのオープンに失敗しました: %v\n", err)
			return e.openArchiveAuto(archivePath)
		}
		e.logger.Printf("Hinanawi形式での強制オープンに成功しました\n")
		return archive, nil

	case gameNum == 7:
		e.logger.Printf("Yumemi形式を強制適用します\n")
		archive := e.factory.NewYumemiArchive()
		ok, err := archive.Open(archivePath)
		if !ok || err != nil {
			e.logger.Printf("Yumemi形式でのオープンに失敗しました: %v\n", err)
			return e.openArchiveAuto(archivePath)
		}
		e.logger.Printf("Yumemi形式での強制オープンに成功しました\n")
		return archive, nil

	case gameNum == 8 || gameNum == 9:
		e.logger.Printf("Kaguya形式（タイプ %d）を強制適用します\n", gameNum-8)
		kaguya := e.factory.NewKaguyaArchive()
		if k, ok := kaguya.(*pbgarc.KaguyaArchive); ok {
			k.SetArchiveType(gameNum - 8) // 8→0, 9→1
		}
		ok, err := kaguya.Open(archivePath)
		if !ok || err != nil {
			e.logger.Printf("Kaguya形式でのオープンに失敗しました: %v\n", err)
			return e.openArchiveAuto(archivePath)
		}
		e.logger.Printf("Kaguya形式での強制オープンに成功しました\n")
		return kaguya, nil

	case gameNum >= 10:
		var typeNum int
		if gameNum >= 10 && gameNum <= 11 || gameNum == 95 {
			typeNum = 0
		} else if gameNum == 12 {
			typeNum = 1
		} else {
			typeNum = 2 // 13以降はタイプ2
		}

		e.logger.Printf("Kanako形式（タイプ %d）を強制適用します\n", typeNum)
		kanako := e.factory.NewKanakoArchive()
		if k, ok := kanako.(*pbgarc.KanakoArchive); ok {
			k.SetArchiveType(typeNum)
		}
		ok, err := kanako.Open(archivePath)
		if !ok || err != nil {
			e.logger.Printf("Kanako形式でのオープンに失敗しました: %v\n", err)
			return e.openArchiveAuto(archivePath)
		}
		e.logger.Printf("Kanako形式での強制オープンに成功しました\n")
		return kanako, nil

	default:
		return e.openArchiveAuto(archivePath)
	}
}
