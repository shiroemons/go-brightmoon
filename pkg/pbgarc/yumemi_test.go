package pbgarc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestYumemiArchive_OpenNonExistent(t *testing.T) {
	archive := NewYumemiArchive()
	_, err := archive.Open("/nonexistent/path/to/file.dat")
	if err == nil {
		t.Error("Open() should return error for non-existent file")
	}
}

func TestYumemiArchive_OpenTooSmall(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "small.dat")

	// 16バイト未満のファイルを作成
	if err := os.WriteFile(tmpFile, []byte{0x00, 0x01, 0x02}, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	archive := NewYumemiArchive()
	_, err := archive.Open(tmpFile)
	if err == nil {
		t.Error("Open() should return error for file smaller than header size")
	}
}

func TestYumemiArchive_OpenInvalidHeader(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.dat")

	// 無効なヘッダー
	data := make([]byte, 32)
	// entrySize=0 (無効)
	data[0] = 0x00
	data[1] = 0x00
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	archive := NewYumemiArchive()
	_, err := archive.Open(tmpFile)
	if err == nil {
		t.Error("Open() should return error for invalid header")
	}
}

func TestYumemiArchive_EnumBeforeOpen(t *testing.T) {
	archive := NewYumemiArchive()

	if archive.EnumFirst() {
		t.Error("EnumFirst() should return false before Open()")
	}

	if archive.EnumNext() {
		t.Error("EnumNext() should return false before Open()")
	}
}

func TestYumemiArchive_GetEntryBeforeOpen(t *testing.T) {
	archive := NewYumemiArchive()

	if name := archive.GetEntryName(); name != "" {
		t.Errorf("GetEntryName() = %q, want empty string", name)
	}

	if size := archive.GetOriginalSize(); size != 0 {
		t.Errorf("GetOriginalSize() = %d, want 0", size)
	}

	if size := archive.GetCompressedSize(); size != 0 {
		t.Errorf("GetCompressedSize() = %d, want 0", size)
	}

	if entry := archive.GetEntry(); entry != nil {
		t.Error("GetEntry() should return nil before Open()")
	}
}

func TestYumemiEntry_Methods(t *testing.T) {
	entry := &YumemiEntry{
		Offset:   100,
		OrigSize: 200,
		CompSize: 150,
		Name:     "test.txt",
		Key:      0xAB,
	}

	if name := entry.GetEntryName(); name != "test.txt" {
		t.Errorf("GetEntryName() = %q, want %q", name, "test.txt")
	}

	if size := entry.GetOriginalSize(); size != 200 {
		t.Errorf("GetOriginalSize() = %d, want %d", size, 200)
	}

	if size := entry.GetCompressedSize(); size != 150 {
		t.Errorf("GetCompressedSize() = %d, want %d", size, 150)
	}
}
