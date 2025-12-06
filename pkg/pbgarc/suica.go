package pbgarc

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// SuicaEntry はSuicaアーカイブ内のエントリを表します
type SuicaEntry struct {
	Offset uint32
	Size   uint32
	Name   string
	parent *SuicaArchive
}

// GetEntryName はエントリ名を取得します
func (e *SuicaEntry) GetEntryName() string {
	return e.Name
}

// GetOriginalSize は元のサイズを取得します
func (e *SuicaEntry) GetOriginalSize() uint32 {
	return e.Size
}

// GetCompressedSize は圧縮後のサイズを取得します
func (e *SuicaEntry) GetCompressedSize() uint32 {
	return e.Size // Suicaは圧縮されていない
}

// Extract はエントリを抽出します
func (e *SuicaEntry) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if e.parent == nil {
		return false
	}

	return e.parent.ExtractEntry(e, w, callback, user)
}

// SuicaArchive はSuicaアーカイブを表します
type SuicaArchive struct {
	file     *os.File
	entries  []SuicaEntry
	curIndex int
}

// NewSuicaArchive は新しいSuicaArchiveを作成します
func NewSuicaArchive() *SuicaArchive {
	return &SuicaArchive{
		entries:  make([]SuicaEntry, 0),
		curIndex: -1,
	}
}

// Close はアーカイブファイルを閉じます
func (a *SuicaArchive) Close() error {
	if a.file != nil {
		err := a.file.Close()
		a.file = nil
		return err
	}
	return nil
}

// Open はアーカイブファイルを開きます
func (a *SuicaArchive) Open(filename string) (bool, error) {
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

	// エントリ数を読み込み
	var entryCount uint16
	if err := binary.Read(file, binary.LittleEndian, &entryCount); err != nil {
		return false, err
	}

	// エントリリストを読み込み
	ok, err := a.open(file, uint32(entryCount), uint32(fileSize))
	if !ok {
		return false, err
	}

	success = true
	return true, nil
}

// エントリリストを読み込みます
func (a *SuicaArchive) open(file *os.File, listCount, fileSize uint32) (bool, error) {
	// リストサイズを計算
	listSize := listCount * 0x6C

	// エントリ数の検証
	if listCount == 0 || listSize+2 > fileSize {
		return false, errors.New("invalid entry count or list size")
	}

	// エントリリストを読み込み
	listBuf := make([]byte, listSize)
	if _, err := io.ReadFull(file, listBuf); err != nil {
		return false, err
	}

	// 暗号化解除
	k, t := byte(0x64), byte(0x64)
	for i := uint32(0); i < listSize; i++ {
		listBuf[i] ^= k
		k += t
		t += 0x4D
	}

	// エントリリストを解析
	a.entries = make([]SuicaEntry, 0, listCount)
	for i := uint32(0); i < listCount; i++ {
		p := i * 0x6C

		// 名前を取得
		nameLen := 0
		for j := uint32(0); j < 0x64; j++ {
			if listBuf[p+j] == 0 {
				break
			}
			nameLen++
		}

		if nameLen == 0 {
			return false, errors.New("invalid entry name")
		}

		// エントリを作成
		entry := SuicaEntry{
			Offset: binary.LittleEndian.Uint32(listBuf[p+0x68 : p+0x6C]),
			Size:   binary.LittleEndian.Uint32(listBuf[p+0x64 : p+0x68]),
			Name:   string(listBuf[p : p+uint32(nameLen)]),
			parent: a,
		}

		// オフセットとサイズの検証
		if entry.Offset < listSize+2 || entry.Offset > fileSize {
			return false, errors.New("invalid entry offset")
		}

		if entry.Size > fileSize-entry.Offset {
			return false, errors.New("invalid entry size")
		}

		a.entries = append(a.entries, entry)
	}

	return true, nil
}

// EnumFirst は最初のエントリに移動します
func (a *SuicaArchive) EnumFirst() bool {
	if len(a.entries) == 0 {
		return false
	}
	a.curIndex = 0
	return true
}

// EnumNext は次のエントリに移動します
func (a *SuicaArchive) EnumNext() bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries)-1 {
		return false
	}
	a.curIndex++
	return true
}

// GetEntryName は現在のエントリ名を取得します
func (a *SuicaArchive) GetEntryName() string {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return ""
	}
	return a.entries[a.curIndex].Name
}

// GetOriginalSize は元のサイズを取得します
func (a *SuicaArchive) GetOriginalSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return a.entries[a.curIndex].Size
}

// GetCompressedSize は圧縮後のサイズを取得します
func (a *SuicaArchive) GetCompressedSize() uint32 {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return 0
	}
	return a.entries[a.curIndex].Size
}

// GetEntry は現在のエントリを取得します
func (a *SuicaArchive) GetEntry() PBGArchiveEntry {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return nil
	}
	return &a.entries[a.curIndex]
}

// Extract は現在のエントリを抽出します
func (a *SuicaArchive) Extract(w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
	if a.curIndex < 0 || a.curIndex >= len(a.entries) {
		return false
	}
	return a.ExtractEntry(&a.entries[a.curIndex], w, callback, user)
}

// ExtractEntry は指定されたエントリを抽出します
func (a *SuicaArchive) ExtractEntry(entry *SuicaEntry, w io.Writer, callback func(string, interface{}) bool, user interface{}) bool {
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

	// バッファサイズ
	bufSize := uint32(1024)
	if bufSize > entry.Size {
		bufSize = entry.Size
	}

	// データを読み込み
	buffer := make([]byte, bufSize)
	remaining := entry.Size

	for remaining > 0 {
		readSize := bufSize
		if readSize > remaining {
			readSize = remaining
		}

		if _, err := io.ReadFull(a.file, buffer[:readSize]); err != nil {
			return false
		}

		if _, err := w.Write(buffer[:readSize]); err != nil {
			return false
		}

		remaining -= readSize
	}

	if callback != nil {
		if !callback("finished.\r\n", user) {
			return false
		}
	}

	return true
}

// ExtractAll は全てのエントリを抽出します
func (a *SuicaArchive) ExtractAll(callback func(string, interface{}) bool, user interface{}) bool {
	// 未実装
	return false
}
