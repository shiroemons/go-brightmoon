package pbgarc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/shiroemons/go-brightmoon/pkg/crypto"
)

// YumemiEntry はYumemiアーカイブ内のエントリを表します (C++版に合わせる)
type YumemiEntry struct {
	Offset   uint32         // 4 bytes
	CompSize uint16         // 2 bytes
	OrigSize uint16         // 2 bytes
	Key      uint8          // 1 byte
	Name     string         // 13 bytes (raw)
	parent   *YumemiArchive // Extractで使用するため必要

	// C++版のエントリ構造は Magic(2), Key(1), Name(13), CompSize(2), OrigSize(2), Offset(4), Padding(8) = 32 bytes
	// Go版では読み込み時にこれを考慮する必要がある
}

// GetEntryName はエントリ名を取得します (内部のNameはそのまま返す)
func (e *YumemiEntry) GetEntryName() string {
	return e.Name
}

// GetOriginalSize は元のサイズを取得します
func (e *YumemiEntry) GetOriginalSize() uint32 {
	return uint32(e.OrigSize)
}

// GetCompressedSize は圧縮後のサイズを取得します
func (e *YumemiEntry) GetCompressedSize() uint32 {
	return uint32(e.CompSize)
}

// Extract はエントリを抽出します
func (e *YumemiEntry) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if e.parent == nil {
		return false
	}

	return e.parent.ExtractEntry(e, w, callback, user)
}

// YumemiArchive はYumemiアーカイブを表します
type YumemiArchive struct {
	file     *os.File
	entries  []YumemiEntry
	curIndex int
}

// NewYumemiArchive は新しいYumemiArchiveを作成します
func NewYumemiArchive() *YumemiArchive {
	return &YumemiArchive{
		entries:  make([]YumemiEntry, 0),
		curIndex: -1,
	}
}

// isfchr は C++ 版の ValidateName で使われる文字チェック関数
func isfchr(c byte) bool {
	return c >= ' ' && c != '+' && c != ',' && c != ';' && c != '=' && c != '[' && c != ']' && c != '.'
}

// ValidateName は C++ 版のロジックを再現します (8.3形式チェック)
func validateName(nameBytes []byte) (string, bool) {
	var i, j int
	nameLen := len(nameBytes)

	// Find base name length (up to 8 chars)
	for i = 0; i < 8 && i < nameLen && isfchr(nameBytes[i]); i++ {
	}

	// Check for extension
	if i < nameLen && nameBytes[i] == '.' {
		// Find extension length (up to 3 chars)
		for j = 1; j < 4 && i+j < nameLen && isfchr(nameBytes[i+j]); j++ {
		}
	} else {
		// If no dot, check if remaining chars are null
		for k := i; k < nameLen; k++ {
			if nameBytes[k] != 0 {
				return "", false // Invalid char after base name without dot
			}
		}
	}

	// Check if the first character after valid name/ext is null terminator
	if i+j >= nameLen || nameBytes[i+j] != 0 {
		return "", false
	}

	// Check constraints: base name must exist (i>0), extension cannot be empty (j!=1 if dot exists)
	if i == 0 || (j == 1 && nameBytes[i] == '.') {
		return "", false
	}

	// Extract the valid name part (excluding null terminator)
	validName := string(nameBytes[:i+j])
	return validName, true
}

// Open はアーカイブファイルを開きます (C++版のロジックに合わせて修正)
func (a *YumemiArchive) Open(filename string) (bool, error) {
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

	// ヘッダを読み込み (C++版に合わせる)
	header := make([]byte, 16)
	if _, err := io.ReadFull(a.file, header); err != nil {
		a.file.Close()
		return false, fmt.Errorf("failed to read header: %w", err)
	}

	var entrySize uint16 // リスト全体のサイズ (バイト)
	var entryNum uint16  // エントリ数
	var entryKey byte    // リスト復号キー
	buf := bytes.NewReader(header)
	if err := binary.Read(buf, binary.LittleEndian, &entrySize); err != nil { // 2 bytes
		a.file.Close()
		return false, err
	}
	if _, err := buf.Seek(2, io.SeekCurrent); err != nil { // skip 2 bytes padding
		a.file.Close()
		return false, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &entryNum); err != nil { // 2 bytes
		a.file.Close()
		return false, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &entryKey); err != nil { // 1 byte
		a.file.Close()
		return false, err
	}
	// 残り9バイトのパディングは無視

	// ヘッダ情報の検証 (C++版に合わせる)
	if int64(entrySize) > fileSize {
		a.file.Close()
		return false, fmt.Errorf("invalid entry list size %d > filesize %d", entrySize, fileSize)
	}
	if (entrySize&0x1F) != 0 || int(entrySize)/32 < int(entryNum) {
		a.file.Close()
		return false, fmt.Errorf("invalid entry size/num relation: size=%d, num=%d", entrySize, entryNum)
	}

	// リストデータを読み込み (ヘッダの16バイトは読み飛ばし済み)
	// entrySize はリスト全体のサイズだが、ヘッダ分を引く必要があるか？ C++版を見ると limit_input_filter(entrysize) を使っているので、
	// ヘッダを含めたサイズ entrySize を読み込み対象とするのが正しい。
	// ただしファイルポインタはヘッダの次にあるので、読み込むサイズは entrySize - 16
	listDataSize := int(entrySize) - 16
	if listDataSize < 0 {
		a.file.Close()
		return false, fmt.Errorf("invalid list data size: %d", listDataSize)
	}
	listData := make([]byte, listDataSize)
	if _, err := io.ReadFull(a.file, listData); err != nil {
		a.file.Close()
		return false, fmt.Errorf("failed to read list data: %w", err)
	}

	// リストデータを復号 (YumemiCrypt)
	_ = crypto.YumemiCrypt(listData, entryKey) // 更新後のキーは不要

	// エントリリストを構築 (C++版 DeserializeList 相当)
	listReader := bytes.NewReader(listData)
	a.entries = make([]YumemiEntry, 0, entryNum)
	entryBuf := make([]byte, 32) // 各エントリは32バイト

	for i := uint16(0); i < entryNum; i++ {
		// 32バイト読み込み
		n, err := io.ReadFull(listReader, entryBuf)
		if err != nil {
			if err == io.EOF && n == 0 && i > 0 { // 正常に読み込めるエントリが一つでもあればOK?
				// C++版は magic == 0 でループを抜ける。GoではEOFを正常終了とみなす。
				break
			}
			a.file.Close()
			return false, fmt.Errorf("failed to read entry %d: %w", i, err)
		}
		if n != 32 {
			a.file.Close()
			return false, fmt.Errorf("short read for entry %d, expected 32, got %d", i, n)
		}

		entryReader := bytes.NewReader(entryBuf)
		var magic uint16
		var entry YumemiEntry
		var nameBytes [13]byte
		var padding [8]byte

		if err := binary.Read(entryReader, binary.LittleEndian, &magic); err != nil { // 2 bytes
			continue // or return error?
		}
		if magic == 0 { // C++版の終了条件
			break
		}
		if err := binary.Read(entryReader, binary.LittleEndian, &entry.Key); err != nil { // 1 byte
			continue
		}
		if _, err := entryReader.Read(nameBytes[:]); err != nil { // 13 bytes
			continue
		}
		if err := binary.Read(entryReader, binary.LittleEndian, &entry.CompSize); err != nil { // 2 bytes
			continue
		}
		if err := binary.Read(entryReader, binary.LittleEndian, &entry.OrigSize); err != nil { // 2 bytes
			continue
		}
		if err := binary.Read(entryReader, binary.LittleEndian, &entry.Offset); err != nil { // 4 bytes
			continue
		}
		if _, err := entryReader.Read(padding[:]); err != nil { // 8 bytes
			continue
		}

		// マジックナンバー検証
		if magic != 0x9595 && magic != 0xF388 {
			a.file.Close()
			return false, fmt.Errorf("invalid magic 0x%x for entry %d", magic, i)
		}

		// 名前検証 (C++版ロジック)
		validName, ok := validateName(nameBytes[:])
		if !ok {
			a.file.Close()
			// C++版は false を返すだけだが、デバッグのためファイル名も出す
			rawName := strings.TrimRight(string(nameBytes[:]), "\x00")
			return false, fmt.Errorf("invalid name for entry %d: raw='%s'", i, rawName)
		}
		entry.Name = validName

		// オフセットとサイズの検証 (C++版に合わせる)
		if int64(entry.Offset) >= fileSize {
			a.file.Close()
			return false, fmt.Errorf("invalid offset %d >= filesize %d for entry %d", entry.Offset, fileSize, i)
		}
		if int64(fileSize)-int64(entry.Offset) < int64(entry.CompSize) {
			a.file.Close()
			return false, fmt.Errorf("invalid size: offset %d + compsize %d > filesize %d for entry %d", entry.Offset, entry.CompSize, fileSize, i)
		}

		entry.parent = a
		a.entries = append(a.entries, entry)
	}

	// 実際に読み込めたエントリ数が entryNum より少ない場合がある (magic==0 で抜けた場合)
	if len(a.entries) == 0 && entryNum > 0 {
		// 有効なエントリが一つもなかった場合
		a.file.Close()
		return false, errors.New("no valid entries found")
	}

	return true, nil
}

// EnumFirst は最初のエントリに移動します
func (a *YumemiArchive) EnumFirst() bool {
	if len(a.entries) == 0 {
		return false
	}
	a.curIndex = 0
	return true
}

// EnumNext は次のエントリに移動します
func (a *YumemiArchive) EnumNext() bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries)-1 {
		return false
	}
	a.curIndex++
	return true
}

// GetEntryName は現在のエントリ名を取得します
func (a *YumemiArchive) GetEntryName() string {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return ""
	}
	return a.entries[a.curIndex].Name
}

// GetOriginalSize は元のサイズを取得します
func (a *YumemiArchive) GetOriginalSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return uint32(a.entries[a.curIndex].OrigSize)
}

// GetCompressedSize は圧縮後のサイズを取得します
func (a *YumemiArchive) GetCompressedSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return uint32(a.entries[a.curIndex].CompSize)
}

// GetEntry は現在のエントリを取得します
func (a *YumemiArchive) GetEntry() PBGArchiveEntry {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return nil
	}
	return &a.entries[a.curIndex]
}

// Extract は現在のエントリを抽出します
func (a *YumemiArchive) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return false
	}
	return a.ExtractEntry(&a.entries[a.curIndex], w, callback, user)
}

// ExtractEntry は指定されたエントリを抽出します
func (a *YumemiArchive) ExtractEntry(entry *YumemiEntry, w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
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

	// データを読み込み
	data := make([]byte, entry.CompSize)
	if _, err := io.ReadFull(a.file, data); err != nil {
		return false
	}

	// 暗号化解除
	yumemiDecrypt(data, entry.Key)

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

// ExtractAll は全てのエントリを抽出します
func (a *YumemiArchive) ExtractAll(callback func(string, interface{}) bool, user interface{}) bool {
	// 未実装
	return false
}

// 暗号化解除関数
func yumemiDecrypt(data []byte, key uint8) {
	for i := range data {
		data[i] ^= key
	}
}
