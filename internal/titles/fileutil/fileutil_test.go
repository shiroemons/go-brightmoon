package fileutil

import (
	"os"
	"path/filepath"
	"strings"
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

func TestFromShiftJIS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "ASCII text",
			input:    "Hello World",
			expected: "Hello World",
			hasError: false,
		},
		{
			name: "Japanese text in Shift_JIS",
			// "こんにちは" in Shift_JIS (as string)
			input:    string([]byte{0x82, 0xb1, 0x82, 0xf1, 0x82, 0xc9, 0x82, 0xbf, 0x82, 0xcd}),
			expected: "こんにちは",
			hasError: false,
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "",
			hasError: false,
		},
		{
			name: "Invalid Shift_JIS sequence",
			// 不正なShift_JISシーケンス
			input:    string([]byte{0xFF, 0xFF, 0xFF}),
			expected: string([]rune{0xFFFD, 0xFFFD, 0xFFFD}), // Unicode replacement character
			hasError: false,                                  // transformは置換文字を使用するのでエラーにはならない
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FromShiftJIS(tt.input)
			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("FromShiftJIS() = %s; want %s", result, tt.expected)
				}
			}
		})
	}
}

func TestNewDatFileFinder(t *testing.T) {
	finder := NewDatFileFinder()
	if finder == nil {
		t.Fatal("NewDatFileFinder() returned nil")
	}

	// DatFileFinder型であることを確認
	// (NewDatFileFinderは*DatFileFinderを返すので、nilチェックで十分)
}

func TestDatFileFinder_Find(t *testing.T) {
	t.Run("カレントディレクトリに1つのdatファイル", func(t *testing.T) {
		// 一時ディレクトリを作成
		tmpDir, err := os.MkdirTemp("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// 元のディレクトリを保存して、テスト後に復元
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(originalDir)

		// テスト用ディレクトリに移動
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// テスト用の.datファイルを作成
		testFile := "th06.dat"
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		finder := NewDatFileFinder()
		result, err := finder.Find()

		if err != nil {
			t.Fatalf("Find() failed: %v", err)
		}

		if !strings.Contains(result, testFile) {
			t.Errorf("Find() did not find %s, got %s", testFile, result)
		}
	})

	t.Run("カレントディレクトリに複数のdatファイル", func(t *testing.T) {
		// 一時ディレクトリを作成
		tmpDir, err := os.MkdirTemp("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// 元のディレクトリを保存して、テスト後に復元
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(originalDir)

		// テスト用ディレクトリに移動
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// 複数の.datファイルを作成
		if err := os.WriteFile("th06.dat", []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("th07.dat", []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		finder := NewDatFileFinder()
		_, err = finder.Find()

		if err == nil {
			t.Error("Expected error for multiple dat files, but got none")
		}
		if !strings.Contains(err.Error(), "複数の.datファイル") {
			t.Errorf("Error message should contain '複数の.datファイル', got %v", err)
		}
	})

	t.Run("datファイルが見つからない", func(t *testing.T) {
		// 一時ディレクトリを作成
		tmpDir, err := os.MkdirTemp("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// 元のディレクトリを保存
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(originalDir)

		// テスト用ディレクトリに移動
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		finder := NewDatFileFinder()
		result, err := finder.Find()

		if err != nil {
			t.Errorf("Find() should not return error for no files, got %v", err)
		}

		if result != "" {
			t.Errorf("Find() should return empty string for no files, got %s", result)
		}
	})

	t.Run("thbgm.datは除外される", func(t *testing.T) {
		// 一時ディレクトリを作成
		tmpDir, err := os.MkdirTemp("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// 元のディレクトリを保存
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(originalDir)

		// テスト用ディレクトリに移動
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// thbgm.datファイルのみ作成
		if err := os.WriteFile("thbgm.dat", []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		finder := NewDatFileFinder()
		result, err := finder.Find()

		if err != nil {
			t.Errorf("Find() should not return error, got %v", err)
		}

		if result != "" {
			t.Errorf("Find() should ignore thbgm.dat and return empty, got %s", result)
		}
	})
}

func TestDatFileFinder_findInDir(t *testing.T) {
	t.Run("ディレクトリ内にサブディレクトリがある場合", func(t *testing.T) {
		// 一時ディレクトリを作成
		tmpDir, err := os.MkdirTemp("", "test")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		// 元のディレクトリを保存
		originalDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(originalDir)

		// テスト用ディレクトリに移動
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}

		// サブディレクトリとファイルを作成
		if err := os.Mkdir("subdir", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("th06.dat", []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("subdir/th07.dat", []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		finder := NewDatFileFinder()
		result, err := finder.Find()

		if err != nil {
			t.Fatalf("Find() failed: %v", err)
		}

		// カレントディレクトリのファイルのみが見つかるはず
		if !strings.Contains(result, "th06.dat") {
			t.Errorf("Find() should find th06.dat in current dir, got %s", result)
		}
		if strings.Contains(result, "th07.dat") {
			t.Errorf("Find() should not find th07.dat in subdir, got %s", result)
		}
	})
}

func TestSaveToFileWithBOM_ErrorCases(t *testing.T) {
	// 無効なパスでのテスト
	err := SaveToFileWithBOM("/nonexistent/dir/file.txt", "test")
	if err == nil {
		t.Error("SaveToFileWithBOM should return error for invalid path")
	}

	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// ディレクトリ作成のパーミッションエラーをシミュレート
	// （実際にはOSによってはエラーにならない場合がある）
	restrictedDir := filepath.Join(tmpDir, "restricted")
	if err := os.Mkdir(restrictedDir, 0000); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(restrictedDir, "subdir", "file.txt")
	err = SaveToFileWithBOM(testFile, "test")
	// このテストはOS/環境によって結果が異なる可能性がある
	if err == nil {
		// パーミッションエラーが発生しない環境の場合はスキップ
		t.Skip("Permission error test skipped on this environment")
	}
}
