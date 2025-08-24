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
)
