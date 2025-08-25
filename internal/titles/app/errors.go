package app

import "errors"

var (
	// ErrParseTHFmt はTHFmtの解析に失敗した場合のエラー
	ErrParseTHFmt = errors.New("THFmtの解析に失敗しました")

	// ErrParseMusicCmt はMusicCmtの解析に失敗した場合のエラー
	ErrParseMusicCmt = errors.New("MusicCmtの解析に失敗しました")

	// ErrSaveFile はファイルの保存に失敗した場合のエラー
	ErrSaveFile = errors.New("ファイルの保存に失敗しました")

	// ErrFileNotFound は必要なファイルが見つからない場合のエラー
	ErrFileNotFound = errors.New("必要なファイルが見つかりませんでした")

	// ErrReadFile はファイルの読み込みに失敗した場合のエラー
	ErrReadFile = errors.New("ファイルの読み込みに失敗しました")

	// ErrNoMusicFiles は音楽ファイルが見つからない場合のエラー
	ErrNoMusicFiles = errors.New("thbgm.fmt、musiccmt.txt または thbgm_tr.fmt、musiccmt_tr.txt のファイルがありません")
)
