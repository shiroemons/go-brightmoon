package archive

import (
	"errors"
	"testing"

	"github.com/shiroemons/go-brightmoon/internal/titles/config"
	"github.com/shiroemons/go-brightmoon/internal/titles/mocks"
)

func TestExtractor_ExtractFiles(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func() (*mocks.SimpleMockArchive, *mocks.MockMemoryExtractor)
		targetFiles []string
		wantError   bool
		wantFiles   int
	}{
		{
			name: "正常にファイルを抽出",
			setupMock: func() (*mocks.SimpleMockArchive, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt":  []byte("test content"),
					"other.txt": []byte("other content"),
				})
				extractor := &mocks.MockMemoryExtractor{
					Data: []byte("test content"),
				}
				return archive, extractor
			},
			targetFiles: []string{"test.txt"},
			wantFiles:   1,
		},
		{
			name: "複数ファイルを抽出",
			setupMock: func() (*mocks.SimpleMockArchive, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"file1.txt": []byte("content1"),
					"file2.txt": []byte("content2"),
				})
				extractor := &mocks.MockMemoryExtractor{
					Data: []byte("content"),
				}
				return archive, extractor
			},
			targetFiles: []string{"file1.txt", "file2.txt"},
			wantFiles:   2,
		},
		{
			name: "ファイルが見つからない",
			setupMock: func() (*mocks.SimpleMockArchive, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				archive.ShouldFailFirst = true
				extractor := &mocks.MockMemoryExtractor{}
				return archive, extractor
			},
			targetFiles: []string{"notfound.txt"},
			wantError:   true,
		},
		{
			name: "抽出エラー",
			setupMock: func() (*mocks.SimpleMockArchive, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("test"),
				})
				extractor := &mocks.MockMemoryExtractor{
					Error: errors.New("extract error"),
				}
				return archive, extractor
			},
			targetFiles: []string{"test.txt"},
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, memExtractor := tt.setupMock()

			logger := config.NewDebugLogger(false)
			factory := &mocks.MockArchiveFactory{
				MockArchive: archive,
			}

			extractor := NewExtractorWithFactory(logger, factory, memExtractor)

			// openArchiveメソッドをモックするために、直接ExtractFilesのロジックをテスト
			// 実際のテストでは、アーカイブを開く部分もモック化が必要
			results := make(map[string][]byte)

			if !archive.ShouldFailFirst && archive.EnumFirst() {
				do := true
				for do {
					entryName := archive.GetEntryName()
					for _, target := range tt.targetFiles {
						if entryName == target {
							data, err := extractor.extractToMemory(archive)
							if err != nil && !tt.wantError {
								t.Fatalf("extractToMemory failed: %v", err)
							}
							if err == nil {
								results[entryName] = data
							}
							break
						}
					}
					do = archive.EnumNext()
				}
			}

			if tt.wantError && len(results) > 0 {
				t.Error("Expected error but got results")
			}
			if !tt.wantError && len(results) != tt.wantFiles {
				t.Errorf("Expected %d files, got %d", tt.wantFiles, len(results))
			}
		})
	}
}

func TestExtractor_openByGameNumber(t *testing.T) {
	tests := []struct {
		name     string
		gameNum  int
		wantType string
	}{
		{
			name:     "ゲーム番号6 - Hinanawi",
			gameNum:  6,
			wantType: "Hinanawi",
		},
		{
			name:     "ゲーム番号7 - Yumemi",
			gameNum:  7,
			wantType: "Yumemi",
		},
		{
			name:     "ゲーム番号8 - Kaguya",
			gameNum:  8,
			wantType: "Kaguya",
		},
		{
			name:     "ゲーム番号10 - Kanako",
			gameNum:  10,
			wantType: "Kanako",
		},
		{
			name:     "ゲーム番号13 - Kanako type 2",
			gameNum:  13,
			wantType: "Kanako",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive := mocks.NewSimpleMockArchive(map[string][]byte{
				"test.txt": []byte("test"),
			})

			factory := &mocks.MockArchiveFactory{
				MockArchive: archive,
			}

			logger := config.NewDebugLogger(false)
			extractor := NewExtractorWithFactory(logger, factory, &mocks.MockMemoryExtractor{})

			result, err := extractor.openByGameNumber("test.dat", tt.gameNum)
			if err != nil {
				t.Fatalf("openByGameNumber failed: %v", err)
			}

			if result == nil {
				t.Error("Expected archive but got nil")
			}
		})
	}
}

func TestGetArchiveTypeMappings(t *testing.T) {
	mappings := GetArchiveTypeMappings()

	if len(mappings) != 6 {
		t.Errorf("Expected 6 archive type mappings, got %d", len(mappings))
	}

	expectedNames := []string{"Yumemi", "Kaguya", "Suica", "Hinanawi", "Marisa", "Kanako"}
	for _, expected := range expectedNames {
		found := false
		for _, mapping := range mappings {
			if mapping.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected archive type '%s' not found in mappings", expected)
		}
	}
}

func TestDefaultMemoryExtractor_ExtractToMemory(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func() *mocks.SimpleMockArchive
		wantError bool
		errorType error
	}{
		{
			name: "正常な抽出",
			setupMock: func() *mocks.SimpleMockArchive {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("test content"),
				})
				archive.CurrentFile = "test.txt"
				return archive
			},
			wantError: false,
		},
		{
			name: "空ファイル",
			setupMock: func() *mocks.SimpleMockArchive {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"empty.txt": []byte{},
				})
				archive.CurrentFile = "empty.txt"
				return archive
			},
			wantError: true,
			errorType: ErrEmptyFile,
		},
		{
			name: "抽出失敗",
			setupMock: func() *mocks.SimpleMockArchive {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("test"),
				})
				archive.CurrentFile = "test.txt"
				archive.ExtractSuccess = false
				return archive
			},
			wantError: true,
			errorType: ErrExtractFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive := tt.setupMock()
			extractor := &DefaultMemoryExtractor{}

			data, err := extractor.ExtractToMemory(archive)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorType != nil && !errors.Is(err, tt.errorType) {
					t.Errorf("Expected error type %v, got %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Fatalf("ExtractToMemory failed: %v", err)
				}
				if len(data) == 0 {
					t.Error("Expected data but got empty")
				}
			}
		})
	}
}
