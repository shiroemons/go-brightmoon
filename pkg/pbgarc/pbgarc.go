// Package pbgarc は東方Projectのアーカイブファイル（.datファイル）を読み込むためのパッケージです。
//
// サポートするアーカイブ形式:
//   - Hinanawi: 東方紅魔郷 (TH06)
//   - Yumemi: 東方妖々夢 (TH07)
//   - Kaguya: 東方永夜抄 (TH08)、弾幕アマノジャク (TH14.3)
//   - Marisa: 東方文花帖 (TH09.5)
//   - Kanako: 東方風神録 (TH10) 以降の作品
//   - Suica: 東方風神録の別形式
//
// 基本的な使い方:
//
//	archive := pbgarc.NewKanakoArchive()
//	archive.SetArchiveType(pbgarc.ARCHTYPE_TD)
//	if ok, err := archive.Open("thbgm.dat"); ok {
//	    defer archive.Close()
//	    for archive.EnumFirst(); ; {
//	        name := archive.GetEntryName()
//	        // エントリを処理...
//	        if !archive.EnumNext() {
//	            break
//	        }
//	    }
//	}
package pbgarc

import "io"

// PBGArchive はアーカイブファイルの基本インターフェース
type PBGArchive interface {
	// Open はアーカイブファイルを開きます
	Open(filename string) (bool, error)

	// Close はアーカイブファイルを閉じます
	Close() error

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

	// Extract は現在のエントリを抽出します。
	// callback は進捗報告用のコールバック関数で、falseを返すと処理を中断します。
	// user はコールバックに渡されるユーザーデータです。
	// コールバックがnilの場合は進捗報告なしで抽出します。
	Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool

	// ExtractAll は全てのエントリを抽出します。
	// callback は進捗報告用のコールバック関数で、falseを返すと処理を中断します。
	// user はコールバックに渡されるユーザーデータです。
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
