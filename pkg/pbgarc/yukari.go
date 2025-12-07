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

// YukariMagic は Yukari アーカイブの識別子 'PBG4' (リトルエンディアン)
const YukariMagic = 0x34474250 // "PBG4" in little-endian

// YukariEntry はYukari(PBG4)アーカイブ内のエントリを表します
type YukariEntry struct {
	Offset uint32 // ファイルデータの開始位置
	Size   uint32 // 展開後サイズ
	ZSize  uint32 // 圧縮サイズ（次のoffset - 現在のoffset）
	Extra  uint32 // 追加情報
	Name   string // ファイル名
	parent *YukariArchive
}

// GetEntryName はエントリ名を取得します
func (e *YukariEntry) GetEntryName() string {
	return e.Name
}

// GetOriginalSize は元のサイズを取得します
func (e *YukariEntry) GetOriginalSize() uint32 {
	return e.Size
}

// GetCompressedSize は圧縮後のサイズを取得します
func (e *YukariEntry) GetCompressedSize() uint32 {
	return e.ZSize
}

// Extract はエントリを抽出します
func (e *YukariEntry) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if e.parent == nil {
		return false
	}

	return e.parent.ExtractEntry(e, w, callback, user)
}

// YukariArchive はYukari(PBG4)アーカイブを表します
type YukariArchive struct {
	file     *os.File
	entries  []YukariEntry
	curIndex int
}

// NewYukariArchive は新しいYukariArchiveを作成します
func NewYukariArchive() *YukariArchive {
	return &YukariArchive{
		entries:  make([]YukariEntry, 0),
		curIndex: -1,
	}
}

// Close はアーカイブファイルを閉じます
func (a *YukariArchive) Close() error {
	if a.file != nil {
		err := a.file.Close()
		a.file = nil
		return err
	}
	return nil
}

// Open はアーカイブファイルを開きます (PBG4形式)
func (a *YukariArchive) Open(filename string) (bool, error) {
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
			a.file = nil
		}
	}()

	// ファイルサイズを取得
	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}
	fileSize := fileInfo.Size()

	// ヘッダを読み込み (16バイト: magic(4) + count(4) + offset(4) + size(4))
	header := make([]byte, 16)
	if _, err := io.ReadFull(a.file, header); err != nil {
		return false, fmt.Errorf("failed to read header: %w", err)
	}

	// マジックナンバーを確認
	var magic uint32
	buf := bytes.NewReader(header)
	if err := binary.Read(buf, binary.LittleEndian, &magic); err != nil {
		return false, err
	}
	if magic != YukariMagic {
		return false, errors.New("invalid magic number: not PBG4")
	}

	// ヘッダ情報を読み込み
	var entryCount uint32
	var listOffset uint32
	var listSize uint32

	if err := binary.Read(buf, binary.LittleEndian, &entryCount); err != nil {
		return false, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &listOffset); err != nil {
		return false, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &listSize); err != nil {
		return false, err
	}

	// ヘッダ情報の検証
	if int64(listOffset) > fileSize {
		return false, fmt.Errorf("invalid list offset %d > filesize %d", listOffset, fileSize)
	}

	// エントリリストの位置に移動
	if _, err := a.file.Seek(int64(listOffset), io.SeekStart); err != nil {
		return false, fmt.Errorf("failed to seek to entry list: %w", err)
	}

	// 圧縮されたエントリリストを読み込み (ファイル末尾まで)
	compressedSize := fileSize - int64(listOffset)
	compressedData := make([]byte, compressedSize)
	if _, err := io.ReadFull(a.file, compressedData); err != nil {
		return false, fmt.Errorf("failed to read compressed entry list: %w", err)
	}

	// LZSS展開
	compressedReader := bytes.NewReader(compressedData)
	var decompressedBuf bytes.Buffer
	if err := crypto.UNLZSS(compressedReader, &decompressedBuf); err != nil {
		return false, fmt.Errorf("failed to decompress entry list: %w", err)
	}

	// エントリリストをパース
	listData := decompressedBuf.Bytes()
	listReader := bytes.NewReader(listData)

	a.entries = make([]YukariEntry, 0, entryCount)

	for i := uint32(0); i < entryCount; i++ {
		var entry YukariEntry

		// ファイル名を読み込み (null終端文字列)
		name, err := readNullTerminatedString(listReader)
		if err != nil {
			return false, fmt.Errorf("failed to read entry name %d: %w", i, err)
		}
		entry.Name = name

		// offset, size, extra を読み込み
		if err := binary.Read(listReader, binary.LittleEndian, &entry.Offset); err != nil {
			return false, fmt.Errorf("failed to read offset for entry %d: %w", i, err)
		}
		if err := binary.Read(listReader, binary.LittleEndian, &entry.Size); err != nil {
			return false, fmt.Errorf("failed to read size for entry %d: %w", i, err)
		}
		if err := binary.Read(listReader, binary.LittleEndian, &entry.Extra); err != nil {
			return false, fmt.Errorf("failed to read extra for entry %d: %w", i, err)
		}

		entry.parent = a
		a.entries = append(a.entries, entry)
	}

	// 各エントリの圧縮サイズを計算 (次のエントリのoffset - 現在のoffset)
	for i := range a.entries {
		if i < len(a.entries)-1 {
			a.entries[i].ZSize = a.entries[i+1].Offset - a.entries[i].Offset
		} else {
			// 最後のエントリはリストの開始位置まで
			a.entries[i].ZSize = listOffset - a.entries[i].Offset
		}
	}

	if len(a.entries) == 0 && entryCount > 0 {
		return false, errors.New("no valid entries found")
	}

	success = true
	return true, nil
}

// readNullTerminatedString はnull終端文字列を読み込みます
func readNullTerminatedString(r io.Reader) (string, error) {
	var result []byte
	buf := make([]byte, 1)

	for {
		n, err := r.Read(buf)
		if err != nil {
			return string(result), err
		}
		if n == 0 {
			return string(result), io.EOF
		}
		if buf[0] == 0 {
			break
		}
		result = append(result, buf[0])
	}

	return string(result), nil
}

// EnumFirst は最初のエントリに移動します
func (a *YukariArchive) EnumFirst() bool {
	if len(a.entries) == 0 {
		return false
	}
	a.curIndex = 0
	return true
}

// EnumNext は次のエントリに移動します
func (a *YukariArchive) EnumNext() bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries)-1 {
		return false
	}
	a.curIndex++
	return true
}

// GetEntryName は現在のエントリ名を取得します
func (a *YukariArchive) GetEntryName() string {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return ""
	}
	return a.entries[a.curIndex].Name
}

// GetOriginalSize は元のサイズを取得します
func (a *YukariArchive) GetOriginalSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return a.entries[a.curIndex].Size
}

// GetCompressedSize は圧縮後のサイズを取得します
func (a *YukariArchive) GetCompressedSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return a.entries[a.curIndex].ZSize
}

// GetEntry は現在のエントリを取得します
func (a *YukariArchive) GetEntry() PBGArchiveEntry {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return nil
	}
	return &a.entries[a.curIndex]
}

// Extract は現在のエントリを抽出します
func (a *YukariArchive) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return false
	}
	return a.ExtractEntry(&a.entries[a.curIndex], w, callback, user)
}

// ExtractEntry は指定されたエントリを抽出します
func (a *YukariArchive) ExtractEntry(entry *YukariEntry, w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
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
	compressedData := make([]byte, entry.ZSize)
	if _, err := io.ReadFull(a.file, compressedData); err != nil {
		return false
	}

	// LZSS展開
	compressedReader := bytes.NewReader(compressedData)
	if err := crypto.UNLZSS(compressedReader, w); err != nil {
		return false
	}

	if callback != nil {
		if !callback("finished.\r\n", user) {
			return false
		}
	}

	return true
}

// ExtractAll は全てのエントリを抽出します
func (a *YukariArchive) ExtractAll(callback func(string, interface{}) bool, user interface{}) bool {
	if !a.EnumFirst() {
		return true // Empty archive is success
	}

	for {
		if !a.Extract(io.Discard, callback, user) {
			return false
		}
		if !a.EnumNext() {
			break
		}
	}

	return true
}
