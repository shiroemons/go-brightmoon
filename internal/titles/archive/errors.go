package archive

import "errors"

var (
	// ErrEmptyFile はファイルサイズが0の場合のエラー
	ErrEmptyFile = errors.New("ファイルサイズが0です")

	// ErrExtractFailed はファイルの展開に失敗した場合のエラー
	ErrExtractFailed = errors.New("ファイルの展開に失敗しました")

	// ErrNoFilesFound はアーカイブ内にファイルが見つからない場合のエラー
	ErrNoFilesFound = errors.New("アーカイブ内にファイルが見つかりません")

	// ErrUnsupportedArchiveType はサポートされていないアーカイブタイプの場合のエラー
	ErrUnsupportedArchiveType = errors.New("サポートされていないアーカイブ形式です")

	// ErrInvalidArchiveType は不明または不正なアーカイブタイプのエラー
	ErrInvalidArchiveType = errors.New("指定されたアーカイブタイプが不明または不正です")

	// ErrArchiveOpenFailed はアーカイブを開けない場合のエラー
	ErrArchiveOpenFailed = errors.New("アーカイブを開けませんでした")

	// ErrArchiveEmpty はアーカイブが空または無効の場合のエラー
	ErrArchiveEmpty = errors.New("アーカイブが無効か空のようです")

	// ErrFileExtraction はファイル抽出中のエラー
	ErrFileExtraction = errors.New("アーカイブからのファイル抽出中にエラーが発生しました")
)
