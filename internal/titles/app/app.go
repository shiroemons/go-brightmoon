// Package app はアプリケーションのメインロジックを実装します
package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shiroemons/go-brightmoon/internal/titles/archive"
	"github.com/shiroemons/go-brightmoon/internal/titles/config"
	"github.com/shiroemons/go-brightmoon/internal/titles/fileutil"
	"github.com/shiroemons/go-brightmoon/internal/titles/interfaces"
	"github.com/shiroemons/go-brightmoon/internal/titles/models"
	"github.com/shiroemons/go-brightmoon/internal/titles/parser"
)

// App はアプリケーションのメインロジックを管理します
type App struct {
	config               *config.Config
	logger               *config.DebugLogger
	extractor            interfaces.Extractor
	thbgmParser          *parser.THBGMParser
	additionalInfoParser *parser.AdditionalInfoParser
	datFileFinder        interfaces.DatFileFinder
	fs                   interfaces.FileSystem
}

// Options はAppの設定オプション
type Options struct {
	FileSystem    interfaces.FileSystem
	Extractor     interfaces.Extractor
	DatFileFinder interfaces.DatFileFinder
}

// New は新しいAppを作成します
func New(cfg *config.Config) *App {
	return NewWithOptions(cfg, Options{})
}

// NewWithOptions は新しいAppをオプション付きで作成します
func NewWithOptions(cfg *config.Config, opts Options) *App {
	logger := config.NewDebugLogger(cfg.DebugMode)

	// デフォルトのファイルシステムを設定
	fs := opts.FileSystem
	if fs == nil {
		fs = fileutil.NewOSFileSystem()
	}

	// デフォルトのExtractorを設定
	var extractor interfaces.Extractor
	if opts.Extractor != nil {
		extractor = opts.Extractor
	} else {
		extractor = archive.NewExtractor(logger)
	}

	// デフォルトのDatFileFinderを設定
	var datFileFinder interfaces.DatFileFinder
	if opts.DatFileFinder != nil {
		datFileFinder = opts.DatFileFinder
	} else {
		datFileFinder = fileutil.NewDatFileFinder()
	}

	return &App{
		config:               cfg,
		logger:               logger,
		extractor:            extractor,
		thbgmParser:          parser.NewTHBGMParser(),
		additionalInfoParser: parser.NewAdditionalInfoParser(),
		datFileFinder:        datFileFinder,
		fs:                   fs,
	}
}

// Run はアプリケーションを実行します
func (a *App) Run(ctx context.Context) error {
	var extractedData models.ExtractedData
	var err error

	// データソースの決定と抽出
	if a.config.ArchivePath != "" {
		// アーカイブが指定されている場合
		extractedData, err = a.processArchive(ctx, a.config.ArchivePath)
	} else {
		// アーカイブが指定されていない場合、自動検出またはローカルファイル
		extractedData, err = a.processAutoDetect(ctx)
	}

	if err != nil {
		return err
	}

	// データの解析
	records, err := a.thbgmParser.ParseTHFmt(extractedData.THFmt)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrParseTHFmt, err)
	}

	tracks, err := a.thbgmParser.ParseMusicCmt(extractedData.MusicCmt)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrParseMusicCmt, err)
	}

	// 補足情報の取得
	additionalInfo := a.additionalInfoParser.CheckAdditionalInfo(extractedData.InputFile)
	if additionalInfo.Error != nil {
		fmt.Fprintf(os.Stderr, "警告: 補足情報の読み込みに失敗しました: %v\n", additionalInfo.Error)
	}

	// 出力の生成
	output := a.generateOutput(records, tracks, additionalInfo)

	// ファイル名の生成と保存
	outputFilename := fileutil.GenerateOutputFilename(extractedData.InputFile)
	outputPath := filepath.Join(a.config.OutputDir, outputFilename)

	if err := fileutil.SaveToFileWithBOM(outputPath, output); err != nil {
		return fmt.Errorf("%w: %w", ErrSaveFile, err)
	}

	a.logger.Printf("データを %s に保存しました\n", outputPath)

	// 標準出力にも表示
	fmt.Println(output)

	return nil
}

// processArchive はアーカイブからデータを抽出します
func (a *App) processArchive(ctx context.Context, archivePath string) (models.ExtractedData, error) {
	// コンテキストのキャンセルチェック
	select {
	case <-ctx.Done():
		return models.ExtractedData{}, ctx.Err()
	default:
	}
	a.logger.Printf("アーカイブファイル %s からデータを読み込みます...\n", archivePath)

	// アーカイブが体験版かどうか判定
	isTrial := fileutil.IsTrialVersion(archivePath)

	// 検索するファイル名
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
	fileData, err := a.extractor.ExtractFiles(ctx, archivePath, a.config.ArchiveType, targetFiles)
	if err != nil {
		return models.ExtractedData{}, fmt.Errorf("アーカイブからのファイル抽出中にエラーが発生しました: %w", err)
	}

	// データの取得
	var thfmt []byte
	var musiccmt string

	if data, ok := fileData[fmtFile]; ok && len(data) > 0 {
		thfmt = data
		if cmtData, ok := fileData[cmtFile]; ok {
			musiccmt = string(cmtData)
		} else {
			return models.ExtractedData{}, fmt.Errorf("%w: %s", ErrFileNotFound, cmtFile)
		}
	} else {
		return models.ExtractedData{}, fmt.Errorf("%w: %s", ErrFileNotFound, fmtFile)
	}

	return models.ExtractedData{
		THFmt:     thfmt,
		MusicCmt:  musiccmt,
		InputFile: archivePath,
	}, nil
}

// processAutoDetect は自動検出またはローカルファイルからデータを取得します
func (a *App) processAutoDetect(ctx context.Context) (models.ExtractedData, error) {
	// コンテキストのキャンセルチェック
	select {
	case <-ctx.Done():
		return models.ExtractedData{}, ctx.Err()
	default:
	}
	// .datファイルの自動検出を試みる
	datFile, err := a.datFileFinder.Find()
	if err != nil {
		return models.ExtractedData{}, err
	}

	if datFile != "" {
		// .datファイルが見つかった場合
		a.logger.Printf("自動検出したアーカイブファイル %s からデータを読み込みます...\n", filepath.Base(datFile))
		return a.processArchive(ctx, datFile)
	}

	// ローカルファイルからの読み込みを試みる
	return a.processLocalFiles(ctx)
}

// processLocalFiles はローカルファイルシステムからファイルを読み込みます
func (a *App) processLocalFiles(ctx context.Context) (models.ExtractedData, error) {
	// コンテキストのキャンセルチェック
	select {
	case <-ctx.Done():
		return models.ExtractedData{}, ctx.Err()
	default:
	}
	// 製品版のファイルをチェック
	if a.fs.FileExists("thbgm.fmt") && a.fs.FileExists("musiccmt.txt") {
		thfmt, err := a.fs.ReadFile("thbgm.fmt")
		if err != nil {
			return models.ExtractedData{}, fmt.Errorf("%w: thbgm.fmt: %w", ErrReadFile, err)
		}

		musiccmtBytes, err := a.fs.ReadFile("musiccmt.txt")
		if err != nil {
			return models.ExtractedData{}, fmt.Errorf("%w: musiccmt.txt: %w", ErrReadFile, err)
		}

		return models.ExtractedData{
			THFmt:     thfmt,
			MusicCmt:  string(musiccmtBytes),
			InputFile: "thbgm",
		}, nil
	}

	// 体験版のファイルをチェック
	if a.fs.FileExists("thbgm_tr.fmt") && a.fs.FileExists("musiccmt_tr.txt") {
		thfmt, err := a.fs.ReadFile("thbgm_tr.fmt")
		if err != nil {
			return models.ExtractedData{}, fmt.Errorf("%w: thbgm_tr.fmt: %w", ErrReadFile, err)
		}

		musiccmtBytes, err := a.fs.ReadFile("musiccmt_tr.txt")
		if err != nil {
			return models.ExtractedData{}, fmt.Errorf("%w: musiccmt_tr.txt: %w", ErrReadFile, err)
		}

		return models.ExtractedData{
			THFmt:     thfmt,
			MusicCmt:  string(musiccmtBytes),
			InputFile: "thbgm_tr",
		}, nil
	}

	return models.ExtractedData{}, ErrNoMusicFiles
}

// generateOutput は出力内容を生成します
func (a *App) generateOutput(records []*models.Record, tracks []*models.Track, additionalInfo models.AdditionalInfo) string {
	var builder strings.Builder

	// 補足情報が存在する場合のみタイトル情報を出力
	if additionalInfo.HasAdditionalInfo {
		if additionalInfo.IsTrialVersion {
			builder.WriteString(fmt.Sprintf("#「%s」体験版曲データ\n", additionalInfo.DisplayTitle))
		} else {
			builder.WriteString(fmt.Sprintf("#「%s」製品版曲データ\n", additionalInfo.DisplayTitle))
		}
		builder.WriteString("#デフォルトのパスと製品名\n")
		builder.WriteString(additionalInfo.TitleInfo + "\n")
	}

	// ヘッダー情報
	builder.WriteString("#曲データ\n")
	builder.WriteString("#開始位置[Bytes]、イントロ部の長さ[Bytes]、ループ部の長さ[Bytes]、曲名\n")
	builder.WriteString("#位置・長さは16進値として記述する\n")

	// トラックデータ
	for _, t := range tracks {
		for _, r := range records {
			if t.FileName == r.FileName {
				builder.WriteString(fmt.Sprintf("%s,%s,%s,%s\n", r.Start, r.Intro, r.Loop, t.Title))
			}
		}
	}

	// トラック情報がない場合のフォールバック
	for i, r := range records {
		if len(tracks) <= i {
			if r.FileName == "th128_08.wav" {
				builder.WriteString(fmt.Sprintf("%s,%s,%s,%s\n", r.Start, r.Intro, r.Loop, "プレイヤーズスコア"))
			} else {
				builder.WriteString(fmt.Sprintf("%s,%s,%s,%s\n", r.Start, r.Intro, r.Loop, r.FileName))
			}
		}
	}

	return builder.String()
}
