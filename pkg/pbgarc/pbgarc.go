package pbgarc

import "io"

// PBGArchive はアーカイブファイルの基本インターフェース
type PBGArchive interface {
	// Open はアーカイブファイルを開きます
	Open(filename string) (bool, error)

	// EnumFirst は最初のエントリに移動します
	EnumFirst() bool

	// EnumNext は次のエントリに移動します
	EnumNext() bool

	// GetEntryName は現在のエントリ名を取得します
	GetEntryName() string

	// GetOriginalSize は元のサイズを取得します
	GetOriginalSize() uint32

	// GetCompressedSize は圧縮後のサイズを取得します
	GetCompressedSize() uint32

	// GetEntry は現在のエントリを取得します
	GetEntry() PBGArchiveEntry

	// Extract は現在のエントリを抽出します
	Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool

	// ExtractAll は全てのエントリを抽出します
	ExtractAll(callback func(string, interface{}) bool, user interface{}) bool
}

// PBGArchiveEntry はアーカイブ内のエントリを表すインターフェース
type PBGArchiveEntry interface {
	// GetEntryName はエントリ名を取得します
	GetEntryName() string

	// GetOriginalSize は元のサイズを取得します
	GetOriginalSize() uint32

	// GetCompressedSize は圧縮後のサイズを取得します
	GetCompressedSize() uint32

	// Extract はエントリを抽出します
	Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool
}
