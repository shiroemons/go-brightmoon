package parser

import "errors"

var (
	// ErrCharacterEncoding は文字コード変換エラー
	ErrCharacterEncoding = errors.New("文字コード変換エラー")

	// ErrScanError はスキャンエラー
	ErrScanError = errors.New("スキャンエラー")

	// ErrReadmeRead はreadme.txtの読み込みに失敗した場合のエラー
	ErrReadmeRead = errors.New("readme.txtの読み込みに失敗しました")

	// ErrReadmeEncodingConversion はreadme.txtの文字コード変換に失敗した場合のエラー
	ErrReadmeEncodingConversion = errors.New("readme.txtの文字コード変換に失敗しました")
)
