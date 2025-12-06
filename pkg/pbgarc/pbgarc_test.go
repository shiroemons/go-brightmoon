package pbgarc

import (
	"testing"
)

func TestNewHinanawiArchive(t *testing.T) {
	archive := NewHinanawiArchive()
	if archive == nil {
		t.Error("NewHinanawiArchive() returned nil")
	}
}

func TestNewMarisaArchive(t *testing.T) {
	archive := NewMarisaArchive()
	if archive == nil {
		t.Error("NewMarisaArchive() returned nil")
	}
}

func TestNewYumemiArchive(t *testing.T) {
	archive := NewYumemiArchive()
	if archive == nil {
		t.Error("NewYumemiArchive() returned nil")
	}
}

func TestNewKaguyaArchive(t *testing.T) {
	archive := NewKaguyaArchive()
	if archive == nil {
		t.Error("NewKaguyaArchive() returned nil")
	}
}

func TestNewKanakoArchive(t *testing.T) {
	archive := NewKanakoArchive()
	if archive == nil {
		t.Error("NewKanakoArchive() returned nil")
	}
}

func TestNewSuicaArchive(t *testing.T) {
	archive := NewSuicaArchive()
	if archive == nil {
		t.Error("NewSuicaArchive() returned nil")
	}
}

func TestKaguyaArchive_SetArchiveType(t *testing.T) {
	archive := NewKaguyaArchive()

	// タイプ0
	archive.SetArchiveType(0)
	// パニックしなければOK

	// タイプ1
	archive.SetArchiveType(1)
	// パニックしなければOK
}

func TestKanakoArchive_SetArchiveType(t *testing.T) {
	archive := NewKanakoArchive()

	tests := []struct {
		name     string
		archType int
	}{
		{"MOF", ARCHTYPE_MOF},
		{"SA_OR_UFO", ARCHTYPE_SA_OR_UFO},
		{"TD", ARCHTYPE_TD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive.SetArchiveType(tt.archType)
			got := archive.GetArchiveType()
			if got != tt.archType {
				t.Errorf("GetArchiveType() = %d, want %d", got, tt.archType)
			}
		})
	}
}

func TestGetArchiveTypeOptions(t *testing.T) {
	options := GetArchiveTypeOptions()
	if len(options) == 0 {
		t.Error("GetArchiveTypeOptions() returned empty slice")
	}

	// 3つのオプションがあることを確認
	if len(options) != 3 {
		t.Errorf("GetArchiveTypeOptions() returned %d options, want 3", len(options))
	}

	// 各オプションが空でないことを確認
	for i, opt := range options {
		if opt == "" {
			t.Errorf("GetArchiveTypeOptions()[%d] is empty", i)
		}
	}
}

// TestClose はClose()メソッドのテストです
func TestHinanawiArchive_Close(t *testing.T) {
	archive := NewHinanawiArchive()
	// ファイルが開かれていない状態でCloseしてもエラーにならない
	if err := archive.Close(); err != nil {
		t.Errorf("Close() on unopened archive returned error: %v", err)
	}
}

func TestMarisaArchive_Close(t *testing.T) {
	archive := NewMarisaArchive()
	if err := archive.Close(); err != nil {
		t.Errorf("Close() on unopened archive returned error: %v", err)
	}
}

func TestYumemiArchive_Close(t *testing.T) {
	archive := NewYumemiArchive()
	if err := archive.Close(); err != nil {
		t.Errorf("Close() on unopened archive returned error: %v", err)
	}
}

func TestKaguyaArchive_Close(t *testing.T) {
	archive := NewKaguyaArchive()
	if err := archive.Close(); err != nil {
		t.Errorf("Close() on unopened archive returned error: %v", err)
	}
}

func TestKanakoArchive_Close(t *testing.T) {
	archive := NewKanakoArchive()
	if err := archive.Close(); err != nil {
		t.Errorf("Close() on unopened archive returned error: %v", err)
	}
}

func TestSuicaArchive_Close(t *testing.T) {
	archive := NewSuicaArchive()
	if err := archive.Close(); err != nil {
		t.Errorf("Close() on unopened archive returned error: %v", err)
	}
}
