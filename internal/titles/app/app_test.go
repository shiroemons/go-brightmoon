package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/shiroemons/go-brightmoon/internal/titles/config"
	"github.com/shiroemons/go-brightmoon/internal/titles/mocks"
	"github.com/shiroemons/go-brightmoon/internal/titles/models"
)

func TestApp_processLocalFiles(t *testing.T) {
	tests := []struct {
		name      string
		files     map[string][]byte
		wantError bool
		wantInput string
	}{
		{
			name: "製品版ファイルが存在する場合",
			files: map[string][]byte{
				"thbgm.fmt":    make([]byte, 52),
				"musiccmt.txt": []byte("test data"),
			},
			wantError: false,
			wantInput: "thbgm",
		},
		{
			name: "体験版ファイルが存在する場合",
			files: map[string][]byte{
				"thbgm_tr.fmt":    make([]byte, 52),
				"musiccmt_tr.txt": []byte("test data"),
			},
			wantError: false,
			wantInput: "thbgm_tr",
		},
		{
			name:      "ファイルが存在しない場合",
			files:     map[string][]byte{},
			wantError: true,
		},
		{
			name: "片方のファイルしか存在しない場合",
			files: map[string][]byte{
				"thbgm.fmt": make([]byte, 52),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックファイルシステムを作成
			fs := mocks.NewMockFileSystem()
			fs.Files = tt.files

			// Appを作成
			cfg := &config.Config{}
			app := NewWithOptions(cfg, Options{
				FileSystem: fs,
			})

			// processLocalFilesを実行
			ctx := context.Background()
			extractedData, err := app.processLocalFiles(ctx)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("processLocalFiles failed: %v", err)
				}
				if extractedData.InputFile != tt.wantInput {
					t.Errorf("Expected InputFile '%s', got '%s'", tt.wantInput, extractedData.InputFile)
				}
			}
		})
	}
}

func TestApp_processLocalFiles_ReadError(t *testing.T) {
	// モックファイルシステムを作成（エラーを返すように設定）
	fs := mocks.NewMockFileSystem()
	fs.Files = map[string][]byte{
		"thbgm.fmt":    make([]byte, 52),
		"musiccmt.txt": []byte("test"),
	}
	fs.Error = errors.New("read error")

	// Appを作成
	cfg := &config.Config{}
	app := NewWithOptions(cfg, Options{
		FileSystem: fs,
	})

	// processLocalFilesを実行
	ctx := context.Background()
	_, err := app.processLocalFiles(ctx)
	if err == nil {
		t.Error("Expected error but got none")
	}
}

func TestApp_processLocalFiles_ContextCancel(t *testing.T) {
	// モックファイルシステムを作成
	fs := mocks.NewMockFileSystem()
	fs.Files = map[string][]byte{
		"thbgm.fmt":    make([]byte, 52),
		"musiccmt.txt": []byte("test data"),
	}

	// Appを作成
	cfg := &config.Config{}
	app := NewWithOptions(cfg, Options{
		FileSystem: fs,
	})

	// キャンセル済みのコンテキストを作成
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 即座にキャンセル

	// processLocalFilesを実行
	_, err := app.processLocalFiles(ctx)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestApp_generateOutput(t *testing.T) {
	cfg := &config.Config{}
	app := New(cfg)

	// テストデータを作成
	records := []*models.Record{
		{
			FileName: "test.wav",
			Start:    "00000000",
			Intro:    "00000100",
			Loop:     "00000200",
			Length:   "00000300",
		},
	}

	tracks := []*models.Track{
		{
			FileName: "test.wav",
			Title:    "Test Track",
		},
	}

	additionalInfo := models.AdditionalInfo{
		HasAdditionalInfo: false,
	}

	// 出力を生成
	output := app.generateOutput(records, tracks, additionalInfo)

	// 出力内容を確認
	expectedContent := []string{
		"#曲データ",
		"00000000,00000100,00000200,Test Track",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(output, expected) {
			t.Errorf("Output does not contain expected string: %s", expected)
		}
	}
}

func TestApp_Run(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.Config
		setupMock     func() *mocks.MockFileSystem
		wantError     bool
		errorContains string
	}{
		{
			name: "ローカルファイル処理が成功",
			config: &config.Config{
				OutputDir: ".",
			},
			setupMock: func() *mocks.MockFileSystem {
				fs := mocks.NewMockFileSystem()
				// thbgm.fmtデータ（52バイトのヘッダー）
				fmtData := make([]byte, 52)
				copy(fmtData[0:], []byte("test.wav\x00"))  // ファイル名
				// musiccmt.txtデータ（Shift-JISを想定）
				cmtData := []byte("@bgm/test\n♪Test Track")
				fs.Files = map[string][]byte{
					"thbgm.fmt":    fmtData,
					"musiccmt.txt": cmtData,
				}
				return fs
			},
			wantError: false,
		},
		{
			name: "ローカルファイルが見つからない",
			config: &config.Config{
				OutputDir: ".",
			},
			setupMock: func() *mocks.MockFileSystem {
				fs := mocks.NewMockFileSystem()
				fs.Files = map[string][]byte{}
				return fs
			},
			wantError:     true,
			errorContains: "thbgm.fmt、musiccmt.txt または thbgm_tr.fmt、musiccmt_tr.txt のファイルがありません",
		},
		{
			name: "出力ディレクトリ作成エラー",
			config: &config.Config{
				OutputDir: "/invalid/path/that/does/not/exist",
			},
			setupMock: func() *mocks.MockFileSystem {
				fs := mocks.NewMockFileSystem()
				// 有効なデータを設定
				fmtData := make([]byte, 52)
				copy(fmtData[0:], []byte("test.wav\x00"))
				cmtData := []byte("@bgm/test\n♪Test Track")
				fs.Files = map[string][]byte{
					"thbgm.fmt":    fmtData,
					"musiccmt.txt": cmtData,
				}
				return fs
			},
			wantError:     true,
			errorContains: "ファイルの保存に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := tt.setupMock()
			app := NewWithOptions(tt.config, Options{
				FileSystem: fs,
			})

			ctx := context.Background()
			err := app.Run(ctx)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error should contain %q, got %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Run failed: %v", err)
				}
			}
		})
	}
}

func TestApp_Run_WithArchive(t *testing.T) {
	// モックエクストラクタを作成
	mockExtractor := &mocks.MockExtractor{
		ExtractedFiles: map[string][]byte{
			"thbgm.fmt":    make([]byte, 52),
			"musiccmt.txt": []byte("@bgm/test\n♪Test Track"),
		},
	}

	// モックファイルシステムを作成
	fs := mocks.NewMockFileSystem()
	fs.Files = map[string][]byte{}

	cfg := &config.Config{
		ArchivePath: "test.dat",
		ArchiveType: 6,
		OutputDir:   ".",
	}

	app := NewWithOptions(cfg, Options{
		FileSystem: fs,
		Extractor:  mockExtractor,
	})

	ctx := context.Background()
	err := app.Run(ctx)
	if err != nil {
		t.Fatalf("Run with archive failed: %v", err)
	}
}

func TestApp_processArchive(t *testing.T) {
	tests := []struct {
		name          string
		archivePath   string
		archiveType   int
		setupMock     func() *mocks.MockExtractor
		wantError     bool
		errorContains string
	}{
		{
			name:        "製品版アーカイブの処理",
			archivePath: "th06.dat",
			archiveType: 6,
			setupMock: func() *mocks.MockExtractor {
				return &mocks.MockExtractor{
					ExtractedFiles: map[string][]byte{
						"thbgm.fmt":    make([]byte, 52),
						"musiccmt.txt": []byte("test"),
					},
				}
			},
			wantError: false,
		},
		{
			name:        "体験版アーカイブの処理",
			archivePath: "th06tr.dat",
			archiveType: 6,
			setupMock: func() *mocks.MockExtractor {
				return &mocks.MockExtractor{
					ExtractedFiles: map[string][]byte{
						"thbgm_tr.fmt":    make([]byte, 52),
						"musiccmt_tr.txt": []byte("test"),
					},
				}
			},
			wantError: false,
		},
		{
			name:        "抽出エラー",
			archivePath: "th06.dat",
			archiveType: 6,
			setupMock: func() *mocks.MockExtractor {
				return &mocks.MockExtractor{
					Error: errors.New("extract failed"),
				}
			},
			wantError:     true,
			errorContains: "アーカイブからのファイル抽出中にエラー",
		},
		{
			name:        "ファイルが見つからない",
			archivePath: "th06.dat",
			archiveType: 6,
			setupMock: func() *mocks.MockExtractor {
				return &mocks.MockExtractor{
					ExtractedFiles: map[string][]byte{},
				}
			},
			wantError:     true,
			errorContains: "ファイルが見つかりません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExtractor := tt.setupMock()
			cfg := &config.Config{
				ArchiveType: tt.archiveType,
			}
			app := NewWithOptions(cfg, Options{
				Extractor: mockExtractor,
			})

			ctx := context.Background()
			_, err := app.processArchive(ctx, tt.archivePath)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error should contain %q, got %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Fatalf("processArchive failed: %v", err)
				}
			}
		})
	}
}

func TestApp_processArchive_ContextCancel(t *testing.T) {
	mockExtractor := &mocks.MockExtractor{
		ExtractedFiles: map[string][]byte{
			"thbgm.fmt":    make([]byte, 52),
			"musiccmt.txt": []byte("test"),
		},
	}

	cfg := &config.Config{
		ArchiveType: 6,
	}
	app := NewWithOptions(cfg, Options{
		Extractor: mockExtractor,
	})

	// キャンセル済みのコンテキストを作成
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := app.processArchive(ctx, "test.dat")
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

func TestApp_processAutoDetect(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func() (*mocks.MockFileSystem, *mocks.MockDatFileFinder, *mocks.MockExtractor)
		wantError bool
	}{
		{
			name: "自動検出でdatファイルを発見",
			setupMock: func() (*mocks.MockFileSystem, *mocks.MockDatFileFinder, *mocks.MockExtractor) {
				fs := mocks.NewMockFileSystem()
				finder := &mocks.MockDatFileFinder{
					FoundFile: "th06.dat",
				}
				extractor := &mocks.MockExtractor{
					ExtractedFiles: map[string][]byte{
						"thbgm.fmt":    make([]byte, 52),
						"musiccmt.txt": []byte("test"),
					},
				}
				return fs, finder, extractor
			},
			wantError: false,
		},
		{
			name: "ローカルファイルで処理",
			setupMock: func() (*mocks.MockFileSystem, *mocks.MockDatFileFinder, *mocks.MockExtractor) {
				fs := mocks.NewMockFileSystem()
				fs.Files = map[string][]byte{
					"thbgm.fmt":    make([]byte, 52),
					"musiccmt.txt": []byte("test"),
				}
				finder := &mocks.MockDatFileFinder{
					FoundFile: "", // datファイルなし
				}
				extractor := &mocks.MockExtractor{}
				return fs, finder, extractor
			},
			wantError: false,
		},
		{
			name: "datファイル検出エラー",
			setupMock: func() (*mocks.MockFileSystem, *mocks.MockDatFileFinder, *mocks.MockExtractor) {
				fs := mocks.NewMockFileSystem()
				finder := &mocks.MockDatFileFinder{
					Error: errors.New("multiple dat files found"),
				}
				extractor := &mocks.MockExtractor{}
				return fs, finder, extractor
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, finder, extractor := tt.setupMock()
			cfg := &config.Config{}
			app := NewWithOptions(cfg, Options{
				FileSystem:    fs,
				DatFileFinder: finder,
				Extractor:     extractor,
			})

			ctx := context.Background()
			_, err := app.processAutoDetect(ctx)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("processAutoDetect failed: %v", err)
				}
			}
		})
	}
}
