package parser

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// toShiftJIS はUTF-8文字列をShift-JISに変換するヘルパー関数
func toShiftJIS(str string) ([]byte, error) {
	var buf bytes.Buffer
	w := transform.NewWriter(&buf, japanese.ShiftJIS.NewEncoder())
	_, err := io.WriteString(w, str)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestAdditionalInfoParser_CheckAdditionalInfo(t *testing.T) {
	tests := []struct {
		name             string
		setupFiles       func(dir string) error
		wantHasInfo      bool
		wantTrialVersion bool
	}{
		{
			name: "製品版の補足情報あり",
			setupFiles: func(dir string) error {
				// readme.txt作成（Shift-JISエンコーディング）
				// 注意: 波ダッシュ「〜」は「～」(U+FF5E)に変更
				readmeContent, err := toShiftJIS("タイトル\n○東方紅魔郷　～ the Embodiment of Scarlet Devil.\n")
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(dir, "readme.txt"), readmeContent, 0644); err != nil {
					return err
				}
				// thbgm.dat作成
				return os.WriteFile(filepath.Join(dir, "thbgm.dat"), []byte("dummy"), 0644)
			},
			wantHasInfo:      true,
			wantTrialVersion: false,
		},
		{
			name: "体験版の補足情報あり",
			setupFiles: func(dir string) error {
				// readme.txt作成（「体験版」を含む）  
				readmeContent, err := toShiftJIS("タイトル\n○東方紅魔郷 体験版\n")
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(dir, "readme.txt"), readmeContent, 0644); err != nil {
					return err
				}
				// thbgm_tr.dat作成
				return os.WriteFile(filepath.Join(dir, "thbgm_tr.dat"), []byte("dummy"), 0644)
			},
			wantHasInfo:      true,
			wantTrialVersion: true,
		},
		{
			name: "readme.txtがない",
			setupFiles: func(dir string) error {
				// thbgm.dat作成
				return os.WriteFile(filepath.Join(dir, "thbgm.dat"), []byte("dummy"), 0644)
			},
			wantHasInfo: false,
		},
		{
			name: "thbgm.datがない",
			setupFiles: func(dir string) error {
				// readme.txt作成
				readmeContent, err := toShiftJIS("タイトル\n○東方紅魔郷\n")
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "readme.txt"), readmeContent, 0644)
			},
			wantHasInfo: false,
		},
		{
			name: "2行目が○で始まらない",
			setupFiles: func(dir string) error {
				// readme.txt作成（○なし）
				readmeContent, err := toShiftJIS("タイトル\n東方紅魔郷\n")
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(dir, "readme.txt"), readmeContent, 0644); err != nil {
					return err
				}
				// thbgm.dat作成
				return os.WriteFile(filepath.Join(dir, "thbgm.dat"), []byte("dummy"), 0644)
			},
			wantHasInfo: false,
		},
		{
			name: "readme.txtが1行しかない",
			setupFiles: func(dir string) error {
				// readme.txt作成（1行のみ）
				readmeContent, err := toShiftJIS("タイトル")
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(dir, "readme.txt"), readmeContent, 0644); err != nil {
					return err
				}
				// thbgm.dat作成
				return os.WriteFile(filepath.Join(dir, "thbgm.dat"), []byte("dummy"), 0644)
			},
			wantHasInfo: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tmpDir, err := os.MkdirTemp("", "test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// テスト用のファイルをセットアップ
			if err := tt.setupFiles(tmpDir); err != nil {
				t.Fatalf("Failed to setup files: %v", err)
			}

			// テスト対象のパーサーを作成
			parser := NewAdditionalInfoParser()

			// ダミーのアーカイブパスを使用（同じディレクトリ内のファイルを検索）
			archivePath := filepath.Join(tmpDir, "dummy.dat")

			// CheckAdditionalInfoを実行
			result := parser.CheckAdditionalInfo(archivePath)

			// 結果を検証
			if result.HasAdditionalInfo != tt.wantHasInfo {
				t.Errorf("HasAdditionalInfo = %v, want %v", result.HasAdditionalInfo, tt.wantHasInfo)
			}

			if result.HasAdditionalInfo && result.IsTrialVersion != tt.wantTrialVersion {
				t.Errorf("IsTrialVersion = %v, want %v", result.IsTrialVersion, tt.wantTrialVersion)
			}

			if result.Error != nil {
				t.Logf("Got error: %v", result.Error)
			}
		})
	}
}

func TestAdditionalInfoParser_CheckAdditionalInfo_Error(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// 読み込めないreadme.txtを作成（権限なし）
	readmePath := filepath.Join(tmpDir, "readme.txt")
	readmeContent, err := toShiftJIS("test")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(readmePath, readmeContent, 0000); err != nil {
		t.Fatal(err)
	}

	// thbgm.dat作成
	if err := os.WriteFile(filepath.Join(tmpDir, "thbgm.dat"), []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewAdditionalInfoParser()
	result := parser.CheckAdditionalInfo(filepath.Join(tmpDir, "dummy.dat"))

	// エラーが返されることを確認
	if result.Error == nil {
		t.Error("Expected error but got none")
	}

	// 補足情報がないことを確認
	if result.HasAdditionalInfo {
		t.Error("Expected HasAdditionalInfo to be false")
	}
}

// TestAdditionalInfoParser_DisplayTitle は表示タイトルの処理をテスト
func TestAdditionalInfoParser_DisplayTitle(t *testing.T) {
	tests := []struct {
		name             string
		title            string
		wantDisplayTitle string
		wantIsTrialVer   bool
	}{
		{
			name:             "製品版タイトル",
			title:            "東方紅魔郷",
			wantDisplayTitle: "東方紅魔郷",
			wantIsTrialVer:   false,
		},
		{
			name:             "体験版タイトル",
			title:            "東方紅魔郷 体験版",
			wantDisplayTitle: "東方紅魔郷",
			wantIsTrialVer:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 一時ディレクトリを作成
			tmpDir, err := os.MkdirTemp("", "test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// readme.txt作成
			readmeContent, err := toShiftJIS("タイトル\n○" + tt.title + "\n")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), readmeContent, 0644); err != nil {
				t.Fatal(err)
			}

			// thbgm.dat作成
			if err := os.WriteFile(filepath.Join(tmpDir, "thbgm.dat"), []byte("dummy"), 0644); err != nil {
				t.Fatal(err)
			}

			parser := NewAdditionalInfoParser()
			result := parser.CheckAdditionalInfo(filepath.Join(tmpDir, "dummy.dat"))

			if result.DisplayTitle != tt.wantDisplayTitle {
				t.Errorf("DisplayTitle = %v, want %v", result.DisplayTitle, tt.wantDisplayTitle)
			}

			if result.IsTrialVersion != tt.wantIsTrialVer {
				t.Errorf("IsTrialVersion = %v, want %v", result.IsTrialVersion, tt.wantIsTrialVer)
			}
		})
	}
}

// TestAdditionalInfoParser_TitleInfo は補足情報のタイトル情報をテスト
func TestAdditionalInfoParser_TitleInfo(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// readme.txt作成
	readmeContent, err := toShiftJIS("タイトル\n○東方紅魔郷\n")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), readmeContent, 0644); err != nil {
		t.Fatal(err)
	}

	// thbgm.dat作成
	thbgmPath := filepath.Join(tmpDir, "thbgm.dat")
	if err := os.WriteFile(thbgmPath, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewAdditionalInfoParser()
	result := parser.CheckAdditionalInfo(filepath.Join(tmpDir, "dummy.dat"))

	expectedTitleInfo := "@" + thbgmPath + ",東方紅魔郷"
	if result.TitleInfo != expectedTitleInfo {
		t.Errorf("TitleInfo = %v, want %v", result.TitleInfo, expectedTitleInfo)
	}
}
