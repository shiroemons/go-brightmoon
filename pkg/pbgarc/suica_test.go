package pbgarc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSuicaArchive_OpenNonExistent(t *testing.T) {
	archive := NewSuicaArchive()
	_, err := archive.Open("/nonexistent/path/to/file.dat")
	if err == nil {
		t.Error("Open() should return error for non-existent file")
	}
}

func TestSuicaArchive_OpenTooSmall(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "small.dat")

	// 2バイト未満のファイルを作成
	if err := os.WriteFile(tmpFile, []byte{0x00}, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	archive := NewSuicaArchive()
	_, err := archive.Open(tmpFile)
	if err == nil {
		t.Error("Open() should return error for file smaller than 2 bytes")
	}
}

func TestSuicaArchive_OpenInvalidEntryCount(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.dat")

	// entryCount=0 (無効)
	data := []byte{0x00, 0x00}
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	archive := NewSuicaArchive()
	_, err := archive.Open(tmpFile)
	if err == nil {
		t.Error("Open() should return error for zero entry count")
	}
}

func TestSuicaArchive_EnumBeforeOpen(t *testing.T) {
	archive := NewSuicaArchive()

	if archive.EnumFirst() {
		t.Error("EnumFirst() should return false before Open()")
	}

	if archive.EnumNext() {
		t.Error("EnumNext() should return false before Open()")
	}
}

func TestSuicaArchive_GetEntryBeforeOpen(t *testing.T) {
	archive := NewSuicaArchive()

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

func TestSuicaEntry_Methods(t *testing.T) {
	entry := &SuicaEntry{
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

	if size := entry.GetCompressedSize(); size != 200 {
		t.Errorf("GetCompressedSize() = %d, want %d", size, 200)
	}
}
