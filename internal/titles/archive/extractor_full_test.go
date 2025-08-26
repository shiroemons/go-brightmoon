package archive

import (
	"context"
	"testing"

	"github.com/shiroemons/go-brightmoon/internal/titles/config"
	"github.com/shiroemons/go-brightmoon/internal/titles/mocks"
)

func TestExtractor_ExtractFiles_Full(t *testing.T) {
	tests := []struct {
		name        string
		archivePath string
		archiveType int
		targetFiles []string
		setupMock   func() *mocks.MockArchiveFactory
		wantError   bool
		wantFiles   int
	}{
		{
			name:        "正常な抽出",
			archivePath: "test.dat",
			archiveType: 6,
			targetFiles: []string{"test.txt"},
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("test content"),
				})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
			wantFiles: 1,
		},
		{
			name:        "自動検出モード",
			archivePath: "test.dat",
			archiveType: -1,
			targetFiles: []string{"test.txt"},
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("test content"),
				})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
			wantFiles: 1,
		},
		{
			name:        "アーカイブオープンエラー",
			archivePath: "test.dat",
			archiveType: 6,
			targetFiles: []string{"test.txt"},
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				archive.ShouldFailOpen = true
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: true,
			wantFiles: 0,
		},
		{
			name:        "ファイルが見つからない",
			archivePath: "test.dat",
			archiveType: 6,
			targetFiles: []string{"missing.txt"},
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{
					"test.txt": []byte("test content"),
				})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: true,
			wantFiles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := config.NewDebugLogger(false)
			factory := tt.setupMock()
			extractor := NewExtractorWithFactory(logger, factory, &mocks.MockMemoryExtractor{})

			ctx := context.Background()
			results, err := extractor.ExtractFiles(ctx, tt.archivePath, tt.archiveType, tt.targetFiles)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("ExtractFiles failed: %v", err)
				}
				if len(results) != tt.wantFiles {
					t.Errorf("Expected %d files, got %d", tt.wantFiles, len(results))
				}
			}
		})
	}
}

func TestExtractor_ExtractFiles_ContextCancel(t *testing.T) {
	logger := config.NewDebugLogger(false)
	archive := mocks.NewSimpleMockArchive(map[string][]byte{
		"test.txt": []byte("test content"),
	})
	factory := &mocks.MockArchiveFactory{
		MockArchive: archive,
	}
	extractor := NewExtractorWithFactory(logger, factory, &mocks.MockMemoryExtractor{})

	// キャンセル済みのコンテキスト
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := extractor.ExtractFiles(ctx, "test.dat", 6, []string{"test.txt"})
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestNewExtractor(t *testing.T) {
	logger := config.NewDebugLogger(false)
	extractor := NewExtractor(logger)
	if extractor == nil {
		t.Fatal("NewExtractor returned nil")
	}
}

func TestExtractor_openArchive(t *testing.T) {
	tests := []struct {
		name        string
		archivePath string
		archiveType int
		setupMock   func() *mocks.MockArchiveFactory
		wantError   bool
	}{
		{
			name:        "特定タイプでオープン",
			archivePath: "test.dat",
			archiveType: 6,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
				}
			},
			wantError: false,
		},
		{
			name:        "自動検出でオープン",
			archivePath: "test.dat",
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
			name:        "オープンエラー",
			archivePath: "test.dat",
			archiveType: 6,
			setupMock: func() *mocks.MockArchiveFactory {
				archive := mocks.NewSimpleMockArchive(map[string][]byte{})
				archive.ShouldFailOpen = true
				return &mocks.MockArchiveFactory{
					MockArchive: archive,
					Error:       ErrArchiveOpenFailed,
				}
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := config.NewDebugLogger(false)
			factory := tt.setupMock()
			extractor := NewExtractorWithFactory(logger, factory, &mocks.MockMemoryExtractor{})

			_, err := extractor.openArchive(tt.archivePath, tt.archiveType)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("openArchive failed: %v", err)
				}
			}
		})
	}
}