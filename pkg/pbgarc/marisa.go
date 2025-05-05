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

// MarisaEntry はMarisaアーカイブ内のエントリを表します (C++版に合わせる)
type MarisaEntry struct {
	Offset uint32 // 4 bytes: データオフセット
	Size   uint32 // 4 bytes: データサイズ
	Name   string // 可変長: ファイル名
	parent *MarisaArchive
	// CompSize は Marisa 形式には存在しない
}

// GetEntryName はエントリ名を取得します
func (e *MarisaEntry) GetEntryName() string {
	return e.Name
}

// GetOriginalSize は元のサイズを取得します
func (e *MarisaEntry) GetOriginalSize() uint32 {
	return e.Size
}

// GetCompressedSize は圧縮後のサイズを取得します (Marisa形式は非圧縮)
func (e *MarisaEntry) GetCompressedSize() uint32 {
	return e.Size
}

// Extract はエントリを抽出します
func (e *MarisaEntry) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if e.parent == nil {
		return false
	}

	return e.parent.ExtractEntry(e, w, callback, user)
}

// MarisaArchive はMarisaアーカイブを表します
type MarisaArchive struct {
	file     *os.File
	entries  []MarisaEntry
	curIndex int
}

// NewMarisaArchive は新しいMarisaArchiveを作成します
func NewMarisaArchive() *MarisaArchive {
	return &MarisaArchive{
		entries:  make([]MarisaEntry, 0),
		curIndex: -1,
	}
}

// Open はアーカイブファイルを開きます (C++版のロジックに合わせて修正)
func (a *MarisaArchive) Open(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	a.file = file

	// ファイルサイズを取得
	fileInfo, err := file.Stat()
	if err != nil {
		a.file.Close()
		return false, err
	}
	fileSize := fileInfo.Size()
	if fileSize < 6 {
		a.file.Close()
		return false, errors.New("file size too small")
	}

	// ヘッダ読み込み (list_count, list_size)
	header := make([]byte, 6)
	if _, err := io.ReadFull(a.file, header); err != nil {
		a.file.Close()
		return false, fmt.Errorf("failed to read header: %w", err)
	}
	var listCount uint16
	var listSize uint32
	headerReader := bytes.NewReader(header)
	if err := binary.Read(headerReader, binary.LittleEndian, &listCount); err != nil {
		a.file.Close()
		return false, err
	}
	if err := binary.Read(headerReader, binary.LittleEndian, &listSize); err != nil {
		a.file.Close()
		return false, err
	}

	if listCount == 0 || listSize == 0 {
		a.file.Close()
		return false, errors.New("invalid list count or size in header")
	}
	if fileSize < 6+int64(listSize) {
		a.file.Close()
		return false, fmt.Errorf("file size %d is smaller than header+list_size %d", fileSize, 6+listSize)
	}

	// リストデータを読み込み
	listBuf := make([]byte, listSize)
	if _, err := io.ReadFull(a.file, listBuf); err != nil {
		a.file.Close()
		return false, fmt.Errorf("failed to read list data: %w", err)
	}

	// リストデータを復号 (MT -> Simple XOR の順で試行)
	listDataMT := make([]byte, listSize)
	copy(listDataMT, listBuf)
	mt := crypto.NewRNGMT(listSize + 6)
	for i := uint32(0); i < listSize; i++ {
		listDataMT[i] ^= byte(mt.NextInt32() & 0xFF)
	}

	a.entries = make([]MarisaEntry, 0, listCount) // スライスを初期化
	ok, err := a.deserializeList(listDataMT, uint32(listCount), listSize, uint32(fileSize))
	if !ok {
		// MT復号が失敗した場合、Simple XOR で試す
		listDataSimpleXOR := make([]byte, listSize)
		copy(listDataSimpleXOR, listBuf) // 元の暗号化データからコピーし直す
		var k byte = 0xC5
		var t byte = 0x89
		for i := uint32(0); i < listSize; i++ {
			listDataSimpleXOR[i] ^= k
			k += t
			t += 0x49
		}
		a.entries = make([]MarisaEntry, 0, listCount) // スライスを再初期化
		ok, err = a.deserializeList(listDataSimpleXOR, uint32(listCount), listSize, uint32(fileSize))
		if !ok {
			a.file.Close()
			return false, fmt.Errorf("failed to deserialize list after both decryptions: %w", err)
		}
	}

	return true, nil
}

// deserializeList は復号されたリストデータを解析します (C++版に合わせて修正)
func (a *MarisaArchive) deserializeList(listBuf []byte, listCount, listSize, fileSize uint32) (bool, error) {
	listReader := bytes.NewReader(listBuf)
	readOffset := uint32(0)

	for i := uint32(0); i < listCount; i++ {
		var entry MarisaEntry
		var nameLen byte

		// 最小エントリサイズ (offset 4 + size 4 + name_len 1 = 9) を確認
		if readOffset+9 > listSize {
			return false, fmt.Errorf("list buffer overflow at entry %d (header read)", i)
		}

		// オフセット(4), サイズ(4) を読み込み
		if err := binary.Read(listReader, binary.LittleEndian, &entry.Offset); err != nil {
			return false, fmt.Errorf("failed to read offset for entry %d: %w", i, err)
		}
		if err := binary.Read(listReader, binary.LittleEndian, &entry.Size); err != nil {
			return false, fmt.Errorf("failed to read size for entry %d: %w", i, err)
		}

		// 名前の長さ(1) を読み込み
		if err := binary.Read(listReader, binary.LittleEndian, &nameLen); err != nil {
			return false, fmt.Errorf("failed to read name_len for entry %d: %w", i, err)
		}

		// 名前データに必要なサイズを確認
		if readOffset+9+uint32(nameLen) > listSize {
			return false, fmt.Errorf("list buffer overflow at entry %d (name read, len=%d)", i, nameLen)
		}

		// 名前を読み込み
		nameBytes := make([]byte, nameLen)
		if _, err := io.ReadFull(listReader, nameBytes); err != nil {
			return false, fmt.Errorf("failed to read name for entry %d: %w", i, err)
		}
		entry.Name = string(nameBytes)

		// エントリデータの検証 (C++版に合わせる)
		if entry.Offset < listSize+6 || entry.Offset > fileSize {
			return false, fmt.Errorf("invalid entry offset %d for entry %d ('%s') (listSize=%d, fileSize=%d)", entry.Offset, i, entry.Name, listSize, fileSize)
		}
		if uint64(entry.Offset)+uint64(entry.Size) > uint64(fileSize) { // uint32 overflow check
			return false, fmt.Errorf("invalid entry size: offset %d + size %d > filesize %d for entry %d ('%s')", entry.Offset, entry.Size, fileSize, i, entry.Name)
		}
		if entry.Name == "" {
			return false, fmt.Errorf("empty name for entry %d", i)
		}

		entry.parent = a
		a.entries = append(a.entries, entry)

		readOffset += 9 + uint32(nameLen)
	}

	// リスト全体のサイズが一致するか確認 (オプション)
	if readOffset != listSize {
		// C++版ではチェックしていないが、念のため警告など出す？
		// return false, fmt.Errorf("list size mismatch: expected %d, read %d", listSize, readOffset)
	}

	return true, nil
}

// EnumFirst は最初のエントリに移動します
func (a *MarisaArchive) EnumFirst() bool {
	if len(a.entries) == 0 {
		return false
	}
	a.curIndex = 0
	return true
}

// EnumNext は次のエントリに移動します
func (a *MarisaArchive) EnumNext() bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries)-1 {
		return false
	}
	a.curIndex++
	return true
}

// GetEntryName は現在のエントリ名を取得します
func (a *MarisaArchive) GetEntryName() string {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return ""
	}
	return a.entries[a.curIndex].Name
}

// GetOriginalSize は元のサイズを取得します
func (a *MarisaArchive) GetOriginalSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return a.entries[a.curIndex].Size
}

// GetCompressedSize は圧縮後のサイズを取得します
func (a *MarisaArchive) GetCompressedSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return a.entries[a.curIndex].Size
}

// GetEntry は現在のエントリを取得します
func (a *MarisaArchive) GetEntry() PBGArchiveEntry {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return nil
	}
	return &a.entries[a.curIndex]
}

// Extract は現在のエントリを抽出します
func (a *MarisaArchive) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return false
	}
	return a.ExtractEntry(&a.entries[a.curIndex], w, callback, user)
}

// ExtractEntry は指定されたエントリを抽出します (C++版のロジックに合わせて修正)
func (a *MarisaArchive) ExtractEntry(entry *MarisaEntry, w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if callback != nil {
		if !callback(entry.GetEntryName(), user) {
			return false
		}
		if !callback(" extracting...", user) {
			return false
		}
	}

	// データを読み込み
	data := make([]byte, entry.Size)
	if _, err := a.file.Seek(int64(entry.Offset), io.SeekStart); err != nil {
		// C++版は fail() のみチェック
		return false
	}
	if _, err := io.ReadFull(a.file, data); err != nil {
		// C++版は eof() || fail() チェック
		return false
	}

	// XOR復号 (C++版のロジック)
	key := byte((entry.Offset >> 1) | 0x23)
	for i := range data {
		data[i] ^= key
	}

	// 書き込み
	if _, err := w.Write(data); err != nil {
		return false
	}

	if callback != nil {
		if !callback("finished.\r\n", user) {
			return false
		}
	}
	return true
}

// ExtractAll はすべてのエントリを抽出します
func (a *MarisaArchive) ExtractAll(callback func(string, interface{}) bool, user interface{}) bool {
	if !a.EnumFirst() {
		return true // Empty archive is success
	}
	success := true
	for {
		// TODO: Implement proper extraction target, not io.Discard
		if !a.Extract(io.Discard, callback, user) {
			success = false
			// Continue on error?
		}
		if !a.EnumNext() {
			break // ループ終了
		}
	}

	return success
}

// --- 元のGo版にあったヘルパー (不要) ---
/*
func marisaDecompress(compressed []byte, originalSize uint32) ([]byte, error) {
	// ... (古いコード)
}
*/
