package config

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
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

func TestParseFlagsWithUsage(t *testing.T) {
	// フラグをリセット
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// 出力をキャプチャ
	var buf bytes.Buffer
	flag.CommandLine.SetOutput(&buf)

	// ParseFlagsを呼び出してUsage関数を設定
	os.Args = []string{"cmd"}
	_ = ParseFlags()

	// Usage関数を直接呼び出す
	flag.Usage()

	// Usage関数の出力を確認
	output := buf.String()
	expectedStrings := []string{
		"Usage of",
		"--archive string",
		"path to .dat archive file",
		"--debug",
		"enable debug output",
		"--version",
		"show version information",
		"-a string",
		"-d\tenable debug output (shorthand)",
		"-v\tshow version information (shorthand)",
		"-o string",
		"output directory for the generated files",
		"-t int",
		"archive type",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Usage output does not contain expected string: %s\nGot output:\n%s", expected, output)
		}
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

func TestParseFlagsWithShorthandOptions(t *testing.T) {
	// フラグをリセット
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// ショートハンドオプションをテスト
	os.Args = []string{"cmd", "-a", "archive.dat", "-d", "-v"}

	cfg := ParseFlags()

	if cfg.ArchivePath != "archive.dat" {
		t.Errorf("Expected ArchivePath 'archive.dat' with -a flag, got '%s'", cfg.ArchivePath)
	}
	if !cfg.DebugMode {
		t.Error("Expected DebugMode to be true with -d flag")
	}
	if !cfg.ShowVersion {
		t.Error("Expected ShowVersion to be true with -v flag")
	}
}

func TestParseFlagsWithLongOptions(t *testing.T) {
	// フラグをリセット
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// ロングオプションをテスト
	os.Args = []string{"cmd", "--archive", "test.dat", "--debug", "--version"}

	cfg := ParseFlags()

	if cfg.ArchivePath != "test.dat" {
		t.Errorf("Expected ArchivePath 'test.dat' with --archive flag, got '%s'", cfg.ArchivePath)
	}
	if !cfg.DebugMode {
		t.Error("Expected DebugMode to be true with --debug flag")
	}
	if !cfg.ShowVersion {
		t.Error("Expected ShowVersion to be true with --version flag")
	}
}

func TestHandleVersion(t *testing.T) {
	// バージョン表示が無効の場合、関数は何もしない
	HandleVersion(false) // この場合、os.Exitは呼ばれない

	// os.Exitが呼ばれることをテストするため、サブプロセスで実行
	if os.Getenv("BE_CRASHER") == "1" {
		HandleVersion(true)
		return
	}

	// サブプロセスでテストを実行
	cmd := exec.Command(os.Args[0], "-test.run=TestHandleVersion")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	// os.Exit(0)が呼ばれるので、エラーはnilではない
	if err == nil {
		t.Error("Expected HandleVersion(true) to call os.Exit")
	}
}
