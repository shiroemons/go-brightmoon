package pbgarc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestYukariArchive_OpenNonExistent(t *testing.T) {
	archive := NewYukariArchive()
	_, err := archive.Open("/nonexistent/path/to/file.dat")
	if err == nil {
		t.Error("Open() should return error for non-existent file")
	}
}

func TestYukariArchive_OpenInvalidMagic(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.dat")

	// 無効なマジックナンバー（"PBG4" ではない）
	data := make([]byte, 100)
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	archive := NewYukariArchive()
	_, err := archive.Open(tmpFile)
	if err == nil {
		t.Error("Open() should return error for invalid magic number")
	}
}

func TestYukariArchive_OpenTooSmall(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "small.dat")

	// ヘッダサイズ未満のファイル
	data := []byte("PBG4")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	archive := NewYukariArchive()
	_, err := archive.Open(tmpFile)
	if err == nil {
		t.Error("Open() should return error for too small file")
	}
}

func TestYukariArchive_EnumBeforeOpen(t *testing.T) {
	archive := NewYukariArchive()

	if archive.EnumFirst() {
		t.Error("EnumFirst() should return false before Open()")
	}

	if archive.EnumNext() {
		t.Error("EnumNext() should return false before Open()")
	}
}

func TestYukariArchive_GetEntryBeforeOpen(t *testing.T) {
	archive := NewYukariArchive()

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

func TestYukariEntry_Methods(t *testing.T) {
	entry := &YukariEntry{
		Offset: 100,
		Size:   200,
		ZSize:  150,
		Extra:  0,
		Name:   "test.txt",
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

func TestYukariArchive_Close(t *testing.T) {
	archive := NewYukariArchive()

	// Close before Open should not panic
	err := archive.Close()
	if err != nil {
		t.Errorf("Close() on unopened archive returned error: %v", err)
	}
}

func TestYukariArchive_MagicConstant(t *testing.T) {
	// YukariMagic should be "PBG4" in little-endian
	expected := uint32(0x34474250) // 'P'=0x50, 'B'=0x42, 'G'=0x47, '4'=0x34
	if YukariMagic != expected {
		t.Errorf("YukariMagic = 0x%x, want 0x%x", YukariMagic, expected)
	}
}

func TestYukariEntry_ExtractWithoutParent(t *testing.T) {
	entry := &YukariEntry{
		Offset: 100,
		Size:   200,
		ZSize:  150,
		Name:   "test.txt",
		parent: nil, // parent is nil
	}

	// Extract should return false when parent is nil
	result := entry.Extract(nil, nil, nil)
	if result {
		t.Error("Extract() should return false when parent is nil")
	}
}

func TestReadNullTerminatedString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
		wantErr  bool
	}{
		{
			name:     "simple string",
			input:    []byte("hello\x00"),
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    []byte("\x00"),
			expected: "",
			wantErr:  false,
		},
		{
			name:     "string with special chars",
			input:    []byte("test.txt\x00"),
			expected: "test.txt",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &byteReader{data: tt.input, pos: 0}
			result, err := readNullTerminatedString(reader)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.wantErr && result != tt.expected {
				t.Errorf("readNullTerminatedString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// byteReader is a simple io.Reader implementation for testing
type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, nil
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
