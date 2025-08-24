package config

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	// フラグをリセット
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// テスト用の引数を設定
	os.Args = []string{"cmd", "-archive", "test.dat", "-t", "1", "-o", "/tmp", "-d"}

	cfg := ParseFlags()

	if cfg.ArchivePath != "test.dat" {
		t.Errorf("Expected ArchivePath 'test.dat', got '%s'", cfg.ArchivePath)
	}
	if cfg.ArchiveType != 1 {
		t.Errorf("Expected ArchiveType 1, got %d", cfg.ArchiveType)
	}
	if cfg.OutputDir != "/tmp" {
		t.Errorf("Expected OutputDir '/tmp', got '%s'", cfg.OutputDir)
	}
	if !cfg.DebugMode {
		t.Error("Expected DebugMode to be true")
	}
}

func TestDebugLogger(t *testing.T) {
	// 出力をキャプチャするためのバッファ
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// デバッグモード有効
	logger := NewDebugLogger(true)
	logger.Printf("test message %d\n", 123)

	w.Close()
	os.Stdout = oldStdout

	// 出力を読み取り
	outputBytes := make([]byte, 1024)
	n, _ := r.Read(outputBytes)
	output := string(outputBytes[:n])

	if !strings.Contains(output, "test message 123") {
		t.Errorf("Expected debug output to contain 'test message 123', got '%s'", output)
	}

	// デバッグモード無効
	buf.Reset()
	logger = NewDebugLogger(false)
	r, w, _ = os.Pipe()
	os.Stdout = w

	logger.Printf("should not appear\n")

	w.Close()
	os.Stdout = oldStdout

	outputBytes = make([]byte, 1024)
	n, _ = r.Read(outputBytes)
	output = string(outputBytes[:n])

	if strings.Contains(output, "should not appear") {
		t.Error("Debug output should not appear when debug mode is disabled")
	}
}
