package pbgarc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHinanawiArchive_OpenNonExistent(t *testing.T) {
	archive := NewHinanawiArchive()
	_, err := archive.Open("/nonexistent/path/to/file.dat")
	if err == nil {
		t.Error("Open() should return error for non-existent file")
	}
}

func TestHinanawiArchive_OpenTooSmall(t *testing.T) {
	// 一時ファイルを作成
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "small.dat")

	// 6バイト未満のファイルを作成
	if err := os.WriteFile(tmpFile, []byte{0x00, 0x01, 0x02}, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	archive := NewHinanawiArchive()
	_, err := archive.Open(tmpFile)
	if err == nil {
		t.Error("Open() should return error for file smaller than 6 bytes")
	}
}

func TestHinanawiArchive_OpenInvalidHeader(t *testing.T) {
	// 一時ファイルを作成
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.dat")

	// 無効なヘッダー（listCount=0, listSize=0）
	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	archive := NewHinanawiArchive()
	_, err := archive.Open(tmpFile)
	if err == nil {
		t.Error("Open() should return error for invalid header")
	}
}

func TestHinanawiArchive_EnumBeforeOpen(t *testing.T) {
	archive := NewHinanawiArchive()

	// Open前のEnumFirst
	if archive.EnumFirst() {
		t.Error("EnumFirst() should return false before Open()")
	}

	// Open前のEnumNext
	if archive.EnumNext() {
		t.Error("EnumNext() should return false before Open()")
	}
}

func TestHinanawiArchive_GetEntryBeforeOpen(t *testing.T) {
	archive := NewHinanawiArchive()

	// Open前のGetEntryName
	if name := archive.GetEntryName(); name != "" {
		t.Errorf("GetEntryName() = %q, want empty string", name)
	}

	// Open前のGetOriginalSize
	if size := archive.GetOriginalSize(); size != 0 {
		t.Errorf("GetOriginalSize() = %d, want 0", size)
	}

	// Open前のGetCompressedSize
	if size := archive.GetCompressedSize(); size != 0 {
		t.Errorf("GetCompressedSize() = %d, want 0", size)
	}

	// Open前のGetEntry
	if entry := archive.GetEntry(); entry != nil {
		t.Error("GetEntry() should return nil before Open()")
	}
}

func TestHinanawiEntry_Methods(t *testing.T) {
	entry := &HinanawiEntry{
		Offset: 100,
		Size:   200,
		Name:   "test.txt",
	}

	if name := entry.GetEntryName(); name != "test.txt" {
		t.Errorf("GetEntryName() = %q, want %q", name, "test.txt")
	}

	if size := entry.GetOriginalSize(); size != 200 {
		t.Errorf("GetOriginalSize() = %d, want %d", size, 200)
	}

	// HinanawiEntry はサイズフィールドが1つのみ（圧縮なし）
	if size := entry.GetCompressedSize(); size != 200 {
		t.Errorf("GetCompressedSize() = %d, want %d", size, 200)
	}
}
