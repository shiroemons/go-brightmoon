package fileutil

import "errors"

var (
	// ErrCreateDirectory は出力先ディレクトリの作成に失敗した場合のエラー
	ErrCreateDirectory = errors.New("出力先ディレクトリの作成に失敗しました")

	// ErrCreateFile はファイルの作成に失敗した場合のエラー
	ErrCreateFile = errors.New("ファイルの作成に失敗しました")

	// ErrWriteBOM はBOMの書き込みに失敗した場合のエラー
	ErrWriteBOM = errors.New("BOMの書き込みに失敗しました")

	// ErrWriteContent は内容の書き込みに失敗した場合のエラー
	ErrWriteContent = errors.New("内容の書き込みに失敗しました")

	// ErrGetCurrentDirectory はカレントディレクトリを取得できない場合のエラー
	ErrGetCurrentDirectory = errors.New("カレントディレクトリを取得できませんでした")

	// ErrGetExecutablePath は実行ファイルのパスを取得できない場合のエラー
	ErrGetExecutablePath = errors.New("実行ファイルのパスを取得できませんでした")

	// ErrReadDirectory はディレクトリ内のファイル一覧を取得できない場合のエラー
	ErrReadDirectory = errors.New("ディレクトリ内のファイル一覧を取得できませんでした")

	// ErrMultipleDatFiles は複数の.datファイルが見つかった場合のエラー
	ErrMultipleDatFiles = errors.New("複数の.datファイルが見つかりました。-archive フラグで使用するファイルを指定してください")
)
