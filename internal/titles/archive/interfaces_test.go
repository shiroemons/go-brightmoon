package archive

import (
	"testing"
)

func TestDefaultArchiveFactory(t *testing.T) {
	factory := &DefaultArchiveFactory{}

	tests := []struct {
		name    string
		newFunc func() interface{}
	}{
		{"NewYumemiArchive", func() interface{} { return factory.NewYumemiArchive() }},
		{"NewKaguyaArchive", func() interface{} { return factory.NewKaguyaArchive() }},
		{"NewSuicaArchive", func() interface{} { return factory.NewSuicaArchive() }},
		{"NewHinanawiArchive", func() interface{} { return factory.NewHinanawiArchive() }},
		{"NewMarisaArchive", func() interface{} { return factory.NewMarisaArchive() }},
		{"NewKanakoArchive", func() interface{} { return factory.NewKanakoArchive() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive := tt.newFunc()
			if archive == nil {
				t.Errorf("%s() returned nil", tt.name)
			}
		})
	}
}

func TestMemoryWriter_Write(t *testing.T) {
	buf := make([]byte, 0)
	writer := &memoryWriter{buf: &buf}

	// 最初の書き込み
	n, err := writer.Write([]byte("hello"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 5 {
		t.Errorf("Write() returned %d, want 5", n)
	}
	if string(*writer.buf) != "hello" {
		t.Errorf("buf = %q, want %q", string(*writer.buf), "hello")
	}

	// 追加の書き込み
	n, err = writer.Write([]byte(" world"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 6 {
		t.Errorf("Write() returned %d, want 6", n)
	}
	if string(*writer.buf) != "hello world" {
		t.Errorf("buf = %q, want %q", string(*writer.buf), "hello world")
	}
}
