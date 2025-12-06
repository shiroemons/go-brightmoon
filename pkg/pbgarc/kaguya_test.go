package pbgarc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKaguyaArchive_OpenNonExistent(t *testing.T) {
	archive := NewKaguyaArchive()
	_, err := archive.Open("/nonexistent/path/to/file.dat")
	if err == nil {
		t.Error("Open() should return error for non-existent file")
	}
}

func TestKaguyaArchive_OpenInvalidMagic(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.dat")

	// 無効なマジックナンバー
	data := make([]byte, 100)
	data[0] = 0x00 // 'ZGBP' (0x5a474250) ではない
	data[1] = 0x00
	data[2] = 0x00
	data[3] = 0x00
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	archive := NewKaguyaArchive()
	_, err := archive.Open(tmpFile)
	if err == nil {
		t.Error("Open() should return error for invalid magic number")
	}
}

func TestKaguyaArchive_EnumBeforeOpen(t *testing.T) {
	archive := NewKaguyaArchive()

	if archive.EnumFirst() {
		t.Error("EnumFirst() should return false before Open()")
	}

	if archive.EnumNext() {
		t.Error("EnumNext() should return false before Open()")
	}
}

func TestKaguyaArchive_GetEntryBeforeOpen(t *testing.T) {
	archive := NewKaguyaArchive()

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

func TestKaguyaEntry_Methods(t *testing.T) {
	entry := &KaguyaEntry{
		Offset:   100,
		OrigSize: 200,
		CompSize: 150,
		Name:     "test.txt",
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
