package pbgarc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKanakoArchive_OpenNonExistent(t *testing.T) {
	archive := NewKanakoArchive()
	_, err := archive.Open("/nonexistent/path/to/file.dat")
	if err == nil {
		t.Error("Open() should return error for non-existent file")
	}
}

func TestKanakoArchive_OpenInvalidMagic(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.dat")

	// 無効なマジックナンバー（THCrypterで復号後に "THA1" にならない）
	data := make([]byte, 100)
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	archive := NewKanakoArchive()
	_, err := archive.Open(tmpFile)
	if err == nil {
		t.Error("Open() should return error for invalid magic number")
	}
}

func TestKanakoArchive_EnumBeforeOpen(t *testing.T) {
	archive := NewKanakoArchive()

	if archive.EnumFirst() {
		t.Error("EnumFirst() should return false before Open()")
	}

	if archive.EnumNext() {
		t.Error("EnumNext() should return false before Open()")
	}
}

func TestKanakoArchive_GetEntryBeforeOpen(t *testing.T) {
	archive := NewKanakoArchive()

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

func TestKanakoEntry_Methods(t *testing.T) {
	entry := &KanakoEntry{
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

func TestKanakoArchive_ArchiveTypeConstants(t *testing.T) {
	// 定数が正しく定義されていることを確認
	if ARCHTYPE_MOF != 0 {
		t.Errorf("ARCHTYPE_MOF = %d, want 0", ARCHTYPE_MOF)
	}
	if ARCHTYPE_SA_OR_UFO != 1 {
		t.Errorf("ARCHTYPE_SA_OR_UFO = %d, want 1", ARCHTYPE_SA_OR_UFO)
	}
	if ARCHTYPE_TD != 2 {
		t.Errorf("ARCHTYPE_TD = %d, want 2", ARCHTYPE_TD)
	}
}
