package app

import (
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
			extractedData, err := app.processLocalFiles()

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
	_, err := app.processLocalFiles()
	if err == nil {
		t.Error("Expected error but got none")
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
