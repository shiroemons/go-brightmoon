package pbgarc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/shiroemons/go-brightmoon/pkg/crypto"
)

// Kanako アーカイブ形式のマジックナンバーと定数
const (
	// KanakoMagic は Kanako アーカイブの識別子 'THA1' (リトルエンディアン)
	KanakoMagic = 0x31414854

	// ヘッダ値のオフセット補正定数（C++版互換）
	kanakoListSizeOffset     = 123456789
	kanakoListCompSizeOffset = 987654321
	kanakoFileCountOffset    = 135792468

	// ヘッダサイズと暗号化パラメータ
	kanakoHeaderSize  = 0x10
	kanakoHeaderKey   = 0x1B
	kanakoHeaderStep  = 0x37
	kanakoHeaderBlock = 0x10
	kanakoHeaderLimit = 0x10
	kanakoListKey     = 0x3e
	kanakoListStep    = 0x9b
	kanakoListBlock   = 0x80
)

// KanakoEntry はKanakoアーカイブ内のエントリを表します
type KanakoEntry struct {
	Offset   uint32
	CompSize uint32
	OrigSize uint32
	Name     string
	parent   *KanakoArchive
}

// GetEntryName はエントリ名を取得します
func (e *KanakoEntry) GetEntryName() string {
	return e.Name
}

// GetOriginalSize は元のサイズを取得します
func (e *KanakoEntry) GetOriginalSize() uint32 {
	return e.OrigSize
}

// GetCompressedSize は圧縮後のサイズを取得します
func (e *KanakoEntry) GetCompressedSize() uint32 {
	return e.CompSize
}

// Extract はエントリを抽出します
func (e *KanakoEntry) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if e.parent == nil {
		return false
	}

	return e.parent.ExtractEntry(e, w, callback, user)
}

// KanakoCryptParam は暗号化パラメータを表します
type KanakoCryptParam struct {
	Key   byte
	Step  byte
	Block int
	Limit int
}

// KanakoArchive はKanakoアーカイブを表します
type KanakoArchive struct {
	file     *os.File
	entries  []KanakoEntry
	curIndex int
	cryprm   []KanakoCryptParam
	archType int
}

// 風神録用暗号化パラメータ
var kanakoCryprm1 = []KanakoCryptParam{
	{0x1b, 0x37, 0x40, 0x2800},
	{0x51, 0xe9, 0x40, 0x3000},
	{0xc1, 0x51, 0x80, 0x3200},
	{0x03, 0x19, 0x400, 0x7800},
	{0xab, 0xcd, 0x200, 0x2800},
	{0x12, 0x34, 0x80, 0x3200},
	{0x35, 0x97, 0x80, 0x2800},
	{0x99, 0x37, 0x400, 0x2000},
}

// 星蓮船/ダブルスポイラー/妖精大戦争用暗号化パラメータ
var kanakoCryprm2 = []KanakoCryptParam{
	{0x1b, 0x73, 0x40, 0x3800},
	{0x51, 0x9e, 0x40, 0x4000},
	{0xc1, 0x15, 0x400, 0x2c00},
	{0x03, 0x91, 0x80, 0x6400},
	{0xab, 0xdc, 0x80, 0x6e00},
	{0x12, 0x43, 0x200, 0x3c00},
	{0x35, 0x79, 0x400, 0x3c00},
	{0x99, 0x7d, 0x80, 0x2800},
}

// 神霊廟用暗号化パラメータ
var kanakoCryprm3 = []KanakoCryptParam{
	{0x1b, 0x73, 0x0100, 0x3800},
	{0x12, 0x43, 0x0200, 0x3e00},
	{0x35, 0x79, 0x0400, 0x3c00},
	{0x03, 0x91, 0x0080, 0x6400},
	{0xab, 0xdc, 0x0080, 0x6e00},
	{0x51, 0x9e, 0x0100, 0x4000},
	{0xc1, 0x15, 0x0400, 0x2c00},
	{0x99, 0x7d, 0x0080, 0x4400},
}

// NewKanakoArchive は新しいKanakoArchiveを作成します
func NewKanakoArchive() *KanakoArchive {
	return &KanakoArchive{
		entries:  make([]KanakoEntry, 0),
		curIndex: -1,
		cryprm:   kanakoCryprm1, // デフォルトは風神録
		archType: 0,
	}
}

// Close はアーカイブファイルを閉じます
func (a *KanakoArchive) Close() error {
	if a.file != nil {
		err := a.file.Close()
		a.file = nil
		return err
	}
	return nil
}

// ArchiveType定数
const (
	ARCHTYPE_MOF       = 0 // TH10 風神録/TH11 地霊殿
	ARCHTYPE_SA_OR_UFO = 1 // TH12 星蓮船/TH12.5 ダブルスポイラー/TH12.8 妖精大戦争
	ARCHTYPE_TD        = 2 // TH13 神霊廟以降の全作品
)

// GetArchiveTypeOptions はアーカイブタイプの選択肢を取得します
func GetArchiveTypeOptions() []string {
	return []string{
		"TH10 Mountain of Faith / TH11 Subterranean Animism",
		"TH12 Undefined Fantastic Object / TH12.5 Double Spoiler / TH12.8 Fairy Wars",
		"TH13 Ten Desires and later games",
	}
}

// SetArchiveType はアーカイブタイプを設定します
func (a *KanakoArchive) SetArchiveType(archType int) {
	a.archType = archType
	switch archType {
	case ARCHTYPE_MOF:
		a.cryprm = kanakoCryprm1
	case ARCHTYPE_SA_OR_UFO:
		a.cryprm = kanakoCryprm2
	case ARCHTYPE_TD:
		a.cryprm = kanakoCryprm3
	default:
		a.cryprm = kanakoCryprm1
	}
}

// GetArchiveType は現在のアーカイブタイプを取得します
func (a *KanakoArchive) GetArchiveType() int {
	return a.archType
}

// Open はアーカイブファイルを開きます
func (a *KanakoArchive) Open(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	a.file = file

	// エラー時にクリーンアップするためのフラグ
	success := false
	defer func() {
		if !success {
			a.file.Close()
		}
	}()

	// ファイルサイズを取得
	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}
	fileSize := fileInfo.Size()

	// ヘッダーを読み込み
	headerBuf := &bytes.Buffer{}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return false, err
	}

	// ヘッダー暗号化解除（固定キー）
	headerReader := io.LimitReader(file, kanakoHeaderSize)
	if !crypto.THCrypter(headerReader, headerBuf, kanakoHeaderSize, kanakoHeaderKey, kanakoHeaderStep, kanakoHeaderBlock, kanakoHeaderLimit) {
		return false, errors.New("header decryption failed")
	}

	// マジックナンバー 'THA1' を確認
	var magic uint32
	if err := binary.Read(headerBuf, binary.LittleEndian, &magic); err != nil {
		return false, err
	}

	if magic != KanakoMagic {
		return false, errors.New("invalid magic number")
	}

	// リスト情報を読み込み
	var listSize, listCompSize, fileCount uint32
	if err := binary.Read(headerBuf, binary.LittleEndian, &listSize); err != nil {
		return false, err
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &listCompSize); err != nil {
		return false, err
	}
	if err := binary.Read(headerBuf, binary.LittleEndian, &fileCount); err != nil {
		return false, err
	}

	// 暗号化された値を復元
	listSize -= kanakoListSizeOffset
	listCompSize -= kanakoListCompSizeOffset
	fileCount -= kanakoFileCountOffset

	// リストのサイズチェック
	if listCompSize > uint32(fileSize) {
		return false, errors.New("invalid list size")
	}

	// リストオフセットを計算
	listOffset := uint32(fileSize) - listCompSize

	// リスト暗号化解除
	if _, err := file.Seek(int64(listOffset), io.SeekStart); err != nil {
		return false, err
	}

	compBuf := &bytes.Buffer{}
	listReader := io.LimitReader(file, int64(listCompSize))
	if !crypto.THCrypter(listReader, compBuf, int(listCompSize), kanakoListKey, kanakoListStep, kanakoListBlock, int(listCompSize)) {
		return false, errors.New("list decryption failed")
	}

	// LZSS解凍
	listBuf := &bytes.Buffer{}
	if err := crypto.UNLZSS(bytes.NewReader(compBuf.Bytes()), listBuf); err != nil {
		return false, fmt.Errorf("list decompression failed: %v", err)
	}

	// エントリ情報を読み込み
	a.entries = make([]KanakoEntry, 0, fileCount)
	for i := uint32(0); i < fileCount; i++ {
		var entry KanakoEntry

		// 名前を読み込み (4バイトずつ、0終端まで)
		var nameParts []byte
		for {
			buff := make([]byte, 4)
			if _, err := io.ReadFull(listBuf, buff); err != nil {
				return false, err
			}

			endIdx := bytes.IndexByte(buff, 0)
			if endIdx >= 0 {
				nameParts = append(nameParts, buff[:endIdx]...)
				break
			}
			nameParts = append(nameParts, buff...)
		}
		entry.Name = string(nameParts)

		// オフセットとサイズを読み込み
		if err := binary.Read(listBuf, binary.LittleEndian, &entry.Offset); err != nil {
			return false, err
		}
		if err := binary.Read(listBuf, binary.LittleEndian, &entry.OrigSize); err != nil {
			return false, err
		}

		// 4バイトの0パディングをスキップ
		padding := make([]byte, 4)
		if _, err := io.ReadFull(listBuf, padding); err != nil {
			return false, err
		}

		entry.parent = a
		a.entries = append(a.entries, entry)
	}

	// 圧縮サイズを計算
	for i := 0; i < len(a.entries)-1; i++ {
		a.entries[i].CompSize = a.entries[i+1].Offset - a.entries[i].Offset
	}
	if len(a.entries) > 0 {
		a.entries[len(a.entries)-1].CompSize = listOffset - a.entries[len(a.entries)-1].Offset
	}

	success = true
	return true, nil
}

// getCryptParamIndex は暗号化パラメータのインデックスを取得します
func (a *KanakoArchive) getCryptParamIndex(entryName string) int {
	index := byte(0)
	for i := 0; i < len(entryName); i++ {
		index += entryName[i]
	}
	return int(index & 7)
}

// EnumFirst は最初のエントリに移動します
func (a *KanakoArchive) EnumFirst() bool {
	if len(a.entries) == 0 {
		return false
	}
	a.curIndex = 0
	return true
}

// EnumNext は次のエントリに移動します
func (a *KanakoArchive) EnumNext() bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries)-1 {
		return false
	}
	a.curIndex++
	return true
}

// GetEntryName は現在のエントリ名を取得します
func (a *KanakoArchive) GetEntryName() string {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return ""
	}
	return a.entries[a.curIndex].Name
}

// GetOriginalSize は元のサイズを取得します
func (a *KanakoArchive) GetOriginalSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return a.entries[a.curIndex].OrigSize
}

// GetCompressedSize は圧縮後のサイズを取得します
func (a *KanakoArchive) GetCompressedSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return a.entries[a.curIndex].CompSize
}

// GetEntry は現在のエントリを取得します
func (a *KanakoArchive) GetEntry() PBGArchiveEntry {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return nil
	}
	return &a.entries[a.curIndex]
}

// Extract は現在のエントリを抽出します
func (a *KanakoArchive) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return false
	}
	return a.ExtractEntry(&a.entries[a.curIndex], w, callback, user)
}

// ExtractEntry は指定されたエントリを抽出します
func (a *KanakoArchive) ExtractEntry(entry *KanakoEntry, w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if callback != nil {
		if !callback(entry.GetEntryName(), user) {
			return false
		}
		if !callback(" extracting...", user) {
			return false
		}
	}

	// ファイルポインタを移動
	if _, err := a.file.Seek(int64(entry.Offset), io.SeekStart); err != nil {
		return false
	}

	// 圧縮データを読み込み
	compressedData := make([]byte, entry.CompSize)
	if _, err := io.ReadFull(a.file, compressedData); err != nil {
		if callback != nil {
			callback("データ読込エラー!\r\n", user)
		}
		return false
	}

	// 暗号化インデックスを取得
	cryIdx := a.getCryptParamIndex(entry.GetEntryName())

	// 暗号化解除バッファ
	compBuf := &bytes.Buffer{}

	// 暗号化解除
	if !crypto.THCrypter(
		bytes.NewReader(compressedData),
		compBuf,
		int(entry.CompSize),
		a.cryprm[cryIdx].Key,
		a.cryprm[cryIdx].Step,
		a.cryprm[cryIdx].Block,
		a.cryprm[cryIdx].Limit,
	) {
		if callback != nil {
			callback("暗号化解除に失敗しました。\r\n", user)
		}
		return false
	}

	// 解凍
	if entry.CompSize == entry.OrigSize {
		// 圧縮なしの場合は直接書き込み
		if _, err := w.Write(compBuf.Bytes()); err != nil {
			if callback != nil {
				callback("書き込みエラー!\r\n", user)
			}
			return false
		}
	} else {
		// LZSS解凍
		if err := crypto.UNLZSS(compBuf, w); err != nil {
			if callback != nil {
				callback("解凍エラー!\r\n", user)
			}
			return false
		}
	}

	if callback != nil {
		if !callback("finished.\r\n", user) {
			return false
		}
	}

	return true
}

// ExtractAll は全てのエントリを抽出します
func (a *KanakoArchive) ExtractAll(callback func(string, interface{}) bool, user interface{}) bool {
	success := true
	for i := range a.entries {
		if !a.ExtractEntry(&a.entries[i], nil, callback, user) {
			success = false
			break
		}
	}
	return success
}
