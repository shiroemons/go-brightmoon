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

// Kaguya アーカイブ形式のマジックナンバーと定数
const (
	// KaguyaMagic は Kaguya アーカイブの識別子 'ZGBP' (リトルエンディアン)
	KaguyaMagic = 0x5a474250

	// ヘッダ値のオフセット補正定数（C++版互換）
	kaguyaFileCountOffset  = 123456
	kaguyaListOffsetOffset = 345678
	kaguyaListSizeOffset   = 567891

	// ヘッダサイズと暗号化パラメータ
	kaguyaHeaderSize      = 12
	kaguyaHeaderKey       = 0x1b
	kaguyaHeaderStep      = 0x37
	kaguyaHeaderBlock     = 0x0c
	kaguyaHeaderLimit     = 0x400
	kaguyaListKey         = 62  // 0x3e
	kaguyaListStep        = 155 // 0x9b
	kaguyaListBlock       = 0x80
	kaguyaListLimit       = 0x400
	kaguyaOrigSizeAdjust  = 4
)

// CryptParam は暗号化パラメータを表します
type CryptParam struct {
	Type  byte
	Key   byte
	Step  byte
	Block int
	Limit int
}

// KaguyaEntry はKaguyaアーカイブ内のエントリを表します
type KaguyaEntry struct {
	Offset   uint32
	Size     uint32
	CompSize uint32 // 圧縮サイズ (計算される)
	OrigSize uint32
	Name     string
	parent   *KaguyaArchive
}

// GetEntryName はエントリ名を取得します
func (e *KaguyaEntry) GetEntryName() string {
	return e.Name
}

// GetOriginalSize は元のサイズを取得します
func (e *KaguyaEntry) GetOriginalSize() uint32 {
	return e.OrigSize
}

// GetCompressedSize は圧縮後のサイズを取得します
func (e *KaguyaEntry) GetCompressedSize() uint32 {
	return e.CompSize
}

// Extract はエントリを抽出します
func (e *KaguyaEntry) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if e.parent == nil {
		return false
	}

	return e.parent.ExtractEntry(e, w, callback, user)
}

// KaguyaArchive はKaguyaアーカイブを表します
type KaguyaArchive struct {
	file     *os.File
	entries  []KaguyaEntry
	curIndex int
	cryprm   []CryptParam
	archType int // 0: 永夜抄, 1: StB (弾幕アマノジャク)
}

// 永夜抄用暗号化パラメータ（Type=4）
var cryprm1 = []CryptParam{
	{0x4d, 0x1b, 0x37, 0x40, 0x2000},
	{0x54, 0x51, 0xe9, 0x40, 0x3000},
	{0x41, 0xc1, 0x51, 0x1400, 0x2000},
	{0x4a, 0x03, 0x19, 0x1400, 0x7800},
	{0x45, 0xab, 0xcd, 0x200, 0x1000},
	{0x57, 0x12, 0x34, 0x400, 0x2800},
	{0x2d, 0x35, 0x97, 0x80, 0x2800},
	{0x2a, 0x99, 0x37, 0x400, 0x1000},
}

// StB用暗号化パラメータ
var cryprm2 = []CryptParam{
	{0x4d, 0x1b, 0x37, 0x40, 0x2800},
	{0x54, 0x51, 0xe9, 0x40, 0x3000},
	{0x41, 0xc1, 0x51, 0x400, 0x400},
	{0x4a, 0x03, 0x19, 0x400, 0x400},
	{0x45, 0xab, 0xcd, 0x200, 0x1000},
	{0x57, 0x12, 0x34, 0x400, 0x400},
	{0x2d, 0x35, 0x97, 0x80, 0x2800},
	{0x2a, 0x99, 0x37, 0x400, 0x1000},
}

// NewKaguyaArchive は新しいKaguyaArchiveを作成します
func NewKaguyaArchive() *KaguyaArchive {
	return &KaguyaArchive{
		entries:  make([]KaguyaEntry, 0),
		curIndex: -1,
		cryprm:   cryprm1, // デフォルトは永夜抄
		archType: 0,
	}
}

// SetArchiveType はアーカイブタイプを設定します
// type=0: 永夜抄用
// type=1: StB用 (弾幕アマノジャク)
func (a *KaguyaArchive) SetArchiveType(archType int) {
	a.archType = archType
	if archType == 0 {
		a.cryprm = cryprm1
	} else {
		a.cryprm = cryprm2
	}
}

// Open はアーカイブファイルを開きます (C++版のロジックに合わせて修正)
func (a *KaguyaArchive) Open(filename string) (bool, error) {
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

	// マジックナンバー 'ZGBP' をチェック
	var magic uint32
	if err := binary.Read(a.file, binary.LittleEndian, &magic); err != nil {
		return false, fmt.Errorf("failed to read magic number: %w", err)
	}
	if magic != KaguyaMagic {
		return false, errors.New("invalid magic number")
	}

	// ヘッダを復号
	headBuf := new(bytes.Buffer)
	if !crypto.THCrypter(a.file, headBuf, kaguyaHeaderSize, kaguyaHeaderKey, kaguyaHeaderStep, kaguyaHeaderBlock, kaguyaHeaderLimit) {
		return false, errors.New("failed to decrypt header")
	}

	// ヘッダ情報を読み込み
	var fileCount, listOffset, listSize uint32
	var errRead error
	errRead = binary.Read(headBuf, binary.LittleEndian, &fileCount)
	if errRead == nil {
		errRead = binary.Read(headBuf, binary.LittleEndian, &listOffset)
	}
	if errRead == nil {
		errRead = binary.Read(headBuf, binary.LittleEndian, &listSize)
	}
	if errRead != nil {
		return false, fmt.Errorf("failed to read header info: %w", errRead)
	}

	// 値を調整 (C++版の定数引き算)
	fileCount -= kaguyaFileCountOffset
	listOffset -= kaguyaListOffsetOffset
	listSize -= kaguyaListSizeOffset // listSize は使われていないが、C++版に合わせて調整

	// listOffset の検証
	if int64(listOffset) >= fileSize {
		return false, fmt.Errorf("invalid list offset %d (filesize %d)", listOffset, fileSize)
	}

	// リストを読み込み (復号 -> 解凍)
	if _, err := a.file.Seek(int64(listOffset), io.SeekStart); err != nil {
		return false, fmt.Errorf("failed to seek to list offset: %w", err)
	}

	// 1. リスト部分を復号
	compListSize := int(fileSize - int64(listOffset))
	cryptedListReader := io.LimitReader(a.file, int64(compListSize))
	compBuf := new(bytes.Buffer)
	if !crypto.THCrypter(cryptedListReader, compBuf, compListSize, kaguyaListKey, kaguyaListStep, kaguyaListBlock, kaguyaListLimit) {
		return false, errors.New("failed to decrypt list data")
	}

	// 2. 復号したリストデータを解凍
	listBuf := new(bytes.Buffer)
	if err := crypto.UNLZSS(compBuf, listBuf); err != nil {
		return false, fmt.Errorf("failed to decompress list data: %w", err)
	}

	// エントリリストを構築
	a.entries = make([]KaguyaEntry, 0, fileCount)
	for i := uint32(0); i < fileCount; i++ {
		var entry KaguyaEntry
		entry.parent = a

		// 名前をヌル終端まで読み込み
		nameBytes := make([]byte, 0, 64) // 適当な初期容量
		for {
			b, err := listBuf.ReadByte()
			if err != nil {
				return false, fmt.Errorf("failed to read entry name byte: %w", err)
			}
			if b == 0 {
				break
			}
			nameBytes = append(nameBytes, b)
		}
		entry.Name = string(nameBytes)

		// オフセット、元サイズ、ダミーを読み込み
		var dummy uint32
		errRead = binary.Read(listBuf, binary.LittleEndian, &entry.Offset)
		if errRead == nil {
			errRead = binary.Read(listBuf, binary.LittleEndian, &entry.OrigSize)
		}
		if errRead == nil {
			errRead = binary.Read(listBuf, binary.LittleEndian, &dummy)
		}
		if errRead != nil {
			return false, fmt.Errorf("failed to read entry metadata for %s: %w", entry.Name, errRead)
		}

		entry.OrigSize -= kaguyaOrigSizeAdjust // C++版の調整

		// オフセット検証
		if int64(entry.Offset) >= fileSize {
			return false, fmt.Errorf("invalid entry offset %d for '%s' (filesize %d)", entry.Offset, entry.Name, fileSize)
		}

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

// EnumFirst は最初のエントリに移動します
func (a *KaguyaArchive) EnumFirst() bool {
	if len(a.entries) == 0 {
		return false
	}
	a.curIndex = 0
	return true
}

// EnumNext は次のエントリに移動します
func (a *KaguyaArchive) EnumNext() bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries)-1 {
		return false
	}
	a.curIndex++
	return true
}

// GetEntryName は現在のエントリ名を取得します
func (a *KaguyaArchive) GetEntryName() string {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return ""
	}
	return a.entries[a.curIndex].Name
}

// GetOriginalSize は元のサイズを取得します
func (a *KaguyaArchive) GetOriginalSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return a.entries[a.curIndex].OrigSize
}

// GetCompressedSize は圧縮後のサイズを取得します
func (a *KaguyaArchive) GetCompressedSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return a.entries[a.curIndex].CompSize
}

// GetEntry は現在のエントリを取得します
func (a *KaguyaArchive) GetEntry() PBGArchiveEntry {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return nil
	}
	// KaguyaEntry が PBGArchiveEntry を満たすようにする
	return &a.entries[a.curIndex]
}

// Extract は現在のエントリを抽出します
func (a *KaguyaArchive) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return false
	}
	return a.ExtractEntry(&a.entries[a.curIndex], w, callback, user)
}

// ExtractEntry は指定されたエントリを抽出します (C++版のロジックに合わせて修正)
func (a *KaguyaArchive) ExtractEntry(entry *KaguyaEntry, w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if callback != nil {
		if !callback(entry.GetEntryName(), user) {
			return false
		}
		if !callback(" extracting...", user) {
			return false
		}
	}

	// ファイルポインタをエントリの開始位置に移動
	if _, err := a.file.Seek(int64(entry.Offset), io.SeekStart); err != nil {
		if callback != nil {
			callback(fmt.Sprintf("failed to seek: %v\r\n", err), user)
		}
		return false
	}

	// 1. データを解凍 (UNLZSS)
	// LimitReader を使用して CompSize 分だけ読み込む
	compressedDataReader := io.LimitReader(a.file, int64(entry.CompSize))
	crypBuf := new(bytes.Buffer)
	if err := crypto.UNLZSS(compressedDataReader, crypBuf); err != nil {
		if callback != nil {
			callback(fmt.Sprintf("failed to decompress: %v\r\n", err), user)
		}
		return false
	}

	// 2. マジックナンバー "edz" + タイプ をチェック
	if crypBuf.Len() < 4 {
		if callback != nil {
			callback("data too short after decompression\r\n", user)
		}
		return false
	}
	magic := crypBuf.Next(4)
	if magic[0] != 'e' || magic[1] != 'd' || magic[2] != 'z' {
		if callback != nil {
			callback("invalid 'edz' magic\r\n", user)
		}
		return false
	}

	// 3. タイプに基づいて暗号化パラメータを検索
	dataType := magic[3]
	var param *CryptParam
	for i := range a.cryprm {
		if a.cryprm[i].Type == dataType {
			param = &a.cryprm[i]
			break
		}
	}
	if param == nil {
		if callback != nil {
			callback(fmt.Sprintf("unknown data type: 0x%x\r\n", dataType), user)
		}
		return false
	}

	// 4. データを復号 (THCrypter)
	// 元のサイズは entry.OrigSize だが、crypBuf の残り全てを復号対象とする
	// (C++版は entry.origsize を渡しているが、内部でサイズ調整されるため、
	//  ここでは解凍後のバッファ全体を渡す方が安全かもしれない)
	decryptedSize := crypBuf.Len() // ヘッダを除いたサイズ
	if !crypto.THCrypter(crypBuf, w, decryptedSize, param.Key, param.Step, param.Block, param.Limit) {
		if callback != nil {
			callback("failed to decrypt data\r\n", user)
		}
		return false
	}

	if callback != nil {
		if !callback("finished.\r\n", user) {
			return false
		}
	}
	return true
}

// ExtractAll はすべてのエントリを抽出します (変更なし)
func (a *KaguyaArchive) ExtractAll(callback func(string, interface{}) bool, user interface{}) bool {
	if !a.EnumFirst() {
		return true // 空のアーカイブは成功とする
	}
	success := true
	for a.EnumNext() {
		if !a.Extract(io.Discard, callback, user) { // 出力先は Discard で良いか？ -> 用途による。通常はファイルに出力
			success = false
			// エラーが発生しても続行するかどうか？ C++版の挙動を確認する必要あり。
			// ここではとりあえず続行する。
		}
	}

	return success
}
