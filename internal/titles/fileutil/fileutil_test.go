package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	// 一時ファイルを作成
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	// 存在するファイルのテスト
	if !FileExists(tmpfile.Name()) {
		t.Errorf("FileExists returned false for existing file")
	}

	// 存在しないファイルのテスト
	if FileExists("/nonexistent/file/path") {
		t.Errorf("FileExists returned true for non-existing file")
	}
}

func TestExtractGameNumber(t *testing.T) {
	tests := []struct {
		filename string
		expected int
	}{
		{"th06.dat", 6},
		{"th08.dat", 8},
		{"th09.dat", 9},
		{"th10.dat", 10},
		{"th128.dat", 128},
		{"th06tr.dat", 6},
		{"th20tr.dat", 20},
		{"thbgm.dat", -1},
		{"notth.dat", -1},
		{"test.dat", -1},
	}

	for _, test := range tests {
		result := ExtractGameNumber(test.filename)
		if result != test.expected {
			t.Errorf("ExtractGameNumber(%s) = %d; want %d", test.filename, result, test.expected)
		}
	}
}

func TestIsTrialVersion(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"th06.dat", false},
		{"th06tr.dat", true},
		{"th20tr.dat", true},
		{"th20.dat", false},
		{"thbgm_tr.fmt", true},
		{"thbgm.fmt", false},
	}

	for _, test := range tests {
		result := IsTrialVersion(test.filename)
		if result != test.expected {
			t.Errorf("IsTrialVersion(%s) = %v; want %v", test.filename, result, test.expected)
		}
	}
}

func TestGenerateOutputFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"th06.dat", "titles_th06.txt"},
		{"th08.dat", "titles_th08.txt"},
		{"/path/to/th10.dat", "titles_th10.txt"},
		{"thbgm", "titles_thbgm.txt"},
		{"thbgm_tr", "titles_thbgm_tr.txt"},
	}

	for _, test := range tests {
		result := GenerateOutputFilename(test.input)
		if result != test.expected {
			t.Errorf("GenerateOutputFilename(%s) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestSaveToFileWithBOM(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// テストファイルのパス
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "テストコンテンツ"

	// ファイルを保存
	err = SaveToFileWithBOM(testFile, content)
	if err != nil {
		t.Fatalf("SaveToFileWithBOM failed: %v", err)
	}

	// ファイルを読み込み
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// BOMをチェック
	if len(data) < 3 || data[0] != 0xEF || data[1] != 0xBB || data[2] != 0xBF {
		t.Error("File does not start with UTF-8 BOM")
	}

	// コンテンツをチェック
	actualContent := string(data[3:])
	if actualContent != content {
		t.Errorf("Content mismatch: got %s, want %s", actualContent, content)
	}
}

func TestDatFilePattern(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"th06.dat", true},
		{"th08.dat", true},
		{"th128.dat", true},
		{"th06tr.dat", true},
		{"th20tr.dat", true},
		{"thbgm.dat", false}, // "thbgm" doesn't have digits after "th"
		{"th.dat", false},    // no digits
		{"06.dat", false},    // doesn't start with "th"
		{"th06.txt", false},  // wrong extension
	}

	for _, test := range tests {
		result := DatFilePattern.MatchString(test.filename)
		if result != test.expected {
			t.Errorf("DatFilePattern.MatchString(%s) = %v; want %v", test.filename, result, test.expected)
		}
	}
}
