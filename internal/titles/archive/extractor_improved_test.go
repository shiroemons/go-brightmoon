package archive

import (
	"context"
	"errors"
	"testing"

	"github.com/shiroemons/go-brightmoon/internal/titles/config"
	"github.com/shiroemons/go-brightmoon/internal/titles/mocks"
)

func TestExtractor_ExtractFiles_Improved(t *testing.T) {
	tests := []struct {
		name        string
		archivePath string
		archiveType int
		targetFiles []string
		setupMock   func() (*mocks.MockArchiveFactory, *mocks.MockMemoryExtractor)
		wantFiles   int
		wantError   bool
	}{
		{
			name:        "ゲーム番号から自動判別 - th06",
			archivePath: "th06.dat",
			archiveType: -1,
			targetFiles: []string{"test.txt"},
			setupMock: func() (*mocks.MockArchiveFactory, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("content"),
				})
				factory := &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
				extractor := &mocks.MockMemoryExtractor{
					Data: []byte("content"),
				}
				return factory, extractor
			},
			wantFiles: 1,
			wantError: false,
		},
		{
			name:        "複数ファイルの抽出",
			archivePath: "th07.dat",
			archiveType: -1,
			targetFiles: []string{"file1.txt", "file2.txt"},
			setupMock: func() (*mocks.MockArchiveFactory, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"file1.txt": []byte("content1"),
					"file2.txt": []byte("content2"),
				})
				factory := &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
				extractor := &mocks.MockMemoryExtractor{
					Data: []byte("content"),
				}
				return factory, extractor
			},
			wantFiles: 2,
			wantError: false,
		},
		{
			name:        "コンテキストキャンセル",
			archivePath: "th06.dat",
			archiveType: -1,
			targetFiles: []string{"test.txt"},
			setupMock: func() (*mocks.MockArchiveFactory, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("content"),
				})
				factory := &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
				extractor := &mocks.MockMemoryExtractor{
					Data: []byte("content"),
				}
				return factory, extractor
			},
			wantFiles: 0,
			wantError: true,
		},
		{
			name:        "ファイルが見つからない場合",
			archivePath: "th06.dat",
			archiveType: -1,
			targetFiles: []string{"notfound.txt"},
			setupMock: func() (*mocks.MockArchiveFactory, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("content"),
				})
				factory := &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
				extractor := &mocks.MockMemoryExtractor{
					Data: []byte("content"),
				}
				return factory, extractor
			},
			wantFiles: 0,
			wantError: false,
		},
		{
			name:        "アーカイブオープンエラー",
			archivePath: "th06.dat",
			archiveType: -1,
			targetFiles: []string{"test.txt"},
			setupMock: func() (*mocks.MockArchiveFactory, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				archive.ShouldFailOpen = true
				factory := &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
				extractor := &mocks.MockMemoryExtractor{}
				return factory, extractor
			},
			wantFiles: 0,
			wantError: true,
		},
		{
			name:        "EnumFirst失敗",
			archivePath: "th06.dat",
			archiveType: -1,
			targetFiles: []string{"test.txt"},
			setupMock: func() (*mocks.MockArchiveFactory, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				archive.ShouldFailFirst = true
				factory := &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
				extractor := &mocks.MockMemoryExtractor{}
				return factory, extractor
			},
			wantFiles: 0,
			wantError: true,
		},
		{
			name:        "メモリ抽出エラー",
			archivePath: "th06.dat",
			archiveType: -1,
			targetFiles: []string{"test.txt"},
			setupMock: func() (*mocks.MockArchiveFactory, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("content"),
				})
				factory := &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
				extractor := &mocks.MockMemoryExtractor{
					Error: errors.New("extraction failed"),
				}
				return factory, extractor
			},
			wantFiles: 0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, memExtractor := tt.setupMock()
			logger := config.NewDebugLogger(false)
			extractor := NewExtractorWithFactory(logger, factory, memExtractor)

			ctx := context.Background()
			if tt.name == "コンテキストキャンセル" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			results, err := extractor.ExtractFiles(ctx, tt.archivePath, tt.archiveType, tt.targetFiles)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil && tt.name != "ファイルが見つからない場合" {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if len(results) != tt.wantFiles {
				t.Errorf("Expected %d files, got %d", tt.wantFiles, len(results))
			}
		})
	}
}

func TestExtractor_openArchive_Improved(t *testing.T) {
	tests := []struct {
		name        string
		archivePath string
		archiveType int
		setupMock   func() *mocks.MockArchiveFactory
		wantError   bool
	}{
		{
			name:        "th06.datから自動判別",
			archivePath: "th06.dat",
			archiveType: -1,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "th07.datから自動判別",
			archivePath: "th07.dat",
			archiveType: -1,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "th08.datから自動判別",
			archivePath: "th08.dat",
			archiveType: -1,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "th10.datから自動判別",
			archivePath: "th10.dat",
			archiveType: -1,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "th13.datから自動判別",
			archivePath: "th13.dat",
			archiveType: -1,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "不明なファイル名",
			archivePath: "unknown.dat",
			archiveType: -1,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				// すべてのアーカイブでオープンに失敗
				archive.ShouldFailOpen = true
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := tt.setupMock()
			logger := config.NewDebugLogger(false)
			extractor := NewExtractorWithFactory(logger, factory, &mocks.MockMemoryExtractor{})

			_, err := extractor.openArchive(tt.archivePath, tt.archiveType)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("openArchive failed: %v", err)
				}
			}
		})
	}
}

func TestExtractor_openByGameNumber_Improved(t *testing.T) {
	tests := []struct {
		name        string
		archivePath string
		gameNum     int
		setupMock   func() *mocks.MockArchiveFactory
		wantError   bool
	}{
		{
			name:        "ゲーム番号6 - Hinanawi",
			archivePath: "th06.dat",
			gameNum:     6,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "ゲーム番号7 - Yumemi",
			archivePath: "th07.dat",
			gameNum:     7,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "ゲーム番号8 - Kaguya タイプ0",
			archivePath: "th08.dat",
			gameNum:     8,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "ゲーム番号9 - Kaguya タイプ1",
			archivePath: "th09.dat",
			gameNum:     9,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "ゲーム番号10 - Kanako タイプ0",
			archivePath: "th10.dat",
			gameNum:     10,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "ゲーム番号12 - Kanako タイプ1",
			archivePath: "th12.dat",
			gameNum:     12,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "ゲーム番号13 - Kanako タイプ2",
			archivePath: "th13.dat",
			gameNum:     13,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "不明なゲーム番号",
			archivePath: "th99.dat",
			gameNum:     99,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				// すべてのアーカイブでオープンに失敗するように設定
				archive.ShouldFailOpen = true
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: true,
		},
		{
			name:        "オープン失敗後のフォールバック",
			archivePath: "th06.dat",
			gameNum:     6,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				archive.ShouldFailOpen = true
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := tt.setupMock()
			logger := config.NewDebugLogger(false)
			extractor := NewExtractorWithFactory(logger, factory, &mocks.MockMemoryExtractor{})

			_, err := extractor.openByGameNumber(tt.archivePath, tt.gameNum)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("openByGameNumber failed: %v", err)
				}
			}
		})
	}
}

func TestExtractor_extractToMemory(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func() (*mocks.SimpleMockArchive, *mocks.MockMemoryExtractor)
		wantError bool
	}{
		{
			name: "正常な抽出",
			setupMock: func() (*mocks.SimpleMockArchive, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("test content"),
				})
				archive.CurrentFile = "test.txt"
				extractor := &mocks.MockMemoryExtractor{
					Data: []byte("test content"),
				}
				return archive, extractor
			},
			wantError: false,
		},
		{
			name: "抽出エラー",
			setupMock: func() (*mocks.SimpleMockArchive, *mocks.MockMemoryExtractor) {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("test content"),
				})
				archive.CurrentFile = "test.txt"
				extractor := &mocks.MockMemoryExtractor{
					Error: errors.New("extraction error"),
				}
				return archive, extractor
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, memExtractor := tt.setupMock()
			logger := config.NewDebugLogger(false)
			extractor := NewExtractorWithFactory(logger, &mocks.MockArchiveFactory{}, memExtractor)

			data, err := extractor.extractToMemory(archive)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("extractToMemory failed: %v", err)
				}
				if len(data) == 0 {
					t.Error("Expected data but got empty")
				}
			}
		})
	}
}