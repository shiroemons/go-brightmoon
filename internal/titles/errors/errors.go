// Package errors はカスタムエラータイプを提供します
package errors

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrFileNotFound はファイルが見つからない場合のエラー
	ErrFileNotFound = errors.New("ファイルが見つかりません")

	// ErrInvalidArchive はアーカイブが無効な場合のエラー
	ErrInvalidArchive = errors.New("無効なアーカイブファイルです")

	// ErrNoDataFound はデータが見つからない場合のエラー
	ErrNoDataFound = errors.New("必要なデータが見つかりません")

	// ErrParseFailure は解析に失敗した場合のエラー
	ErrParseFailure = errors.New("データの解析に失敗しました")
)

// ArchiveError はアーカイブ関連のエラー
type ArchiveError struct {
	Op   string // 実行していた操作
	Path string // ファイルパス
	Err  error  // 元のエラー
}

// Error はエラーメッセージを返します
func (e *ArchiveError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

// Unwrap は元のエラーを返します
func (e *ArchiveError) Unwrap() error {
	return e.Err
}

// NewArchiveError は新しいArchiveErrorを作成します
func NewArchiveError(op, path string, err error) *ArchiveError {
	return &ArchiveError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}

// ParseError は解析関連のエラー
type ParseError struct {
	File string // ファイル名
	Err  error  // 元のエラー
}

// Error はエラーメッセージを返します
func (e *ParseError) Error() string {
	return fmt.Sprintf("%sの解析エラー: %v", e.File, e.Err)
}

// Unwrap は元のエラーを返します
func (e *ParseError) Unwrap() error {
	return e.Err
}

// NewParseError は新しいParseErrorを作成します
func NewParseError(file string, err error) *ParseError {
	return &ParseError{
		File: file,
		Err:  err,
	}
}
