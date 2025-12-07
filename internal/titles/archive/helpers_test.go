package archive

import (
	"testing"

	"github.com/shiroemons/go-brightmoon/internal/titles/config"
	"github.com/shiroemons/go-brightmoon/pkg/pbgarc"
)

// テスト用のヘルパー関数

func TestChooseOldFormat(t *testing.T) {
	logger := config.NewDebugLogger(false)
	extractor := NewExtractor(logger)

	tests := []struct {
		name       string
		candidates []archiveCandidate
		gameNum    int
		wantName   string
		wantNil    bool
	}{
		{
			name: "th06でHinanawi選択",
			candidates: []archiveCandidate{
				{name: "Hinanawi", archive: &pbgarc.HinanawiArchive{}},
				{name: "Yumemi", archive: &pbgarc.YumemiArchive{}},
			},
			gameNum:  6,
			wantName: "Hinanawi",
		},
		{
			name: "th07でYukari選択",
			candidates: []archiveCandidate{
				{name: "Hinanawi", archive: &pbgarc.HinanawiArchive{}},
				{name: "Yukari", archive: &pbgarc.YukariArchive{}},
			},
			gameNum:  7,
			wantName: "Yukari",
		},
		{
			name:       "候補なし",
			candidates: []archiveCandidate{},
			gameNum:    6,
			wantName:   "",
			wantNil:    true,
		},
		{
			name: "マッチなし",
			candidates: []archiveCandidate{
				{name: "Kanako", archive: &pbgarc.KanakoArchive{}},
			},
			gameNum:  6,
			wantName: "",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, name := extractor.chooseOldFormat(tt.candidates, tt.gameNum)

			if name != tt.wantName {
				t.Errorf("Expected name %q, got %q", tt.wantName, name)
			}

			if tt.wantNil && archive != nil {
				t.Error("Expected nil archive but got non-nil")
			}
			if !tt.wantNil && archive == nil {
				t.Error("Expected archive but got nil")
			}
		})
	}
}

func TestChooseKaguya(t *testing.T) {
	logger := config.NewDebugLogger(false)
	extractor := NewExtractor(logger)

	tests := []struct {
		name       string
		candidates []archiveCandidate
		gameNum    int
		wantName   string
		wantType   int
	}{
		{
			name: "th08でタイプ0",
			candidates: []archiveCandidate{
				{name: "Kaguya", archive: pbgarc.NewKaguyaArchive()},
			},
			gameNum:  8,
			wantName: "Kaguya",
			wantType: 0,
		},
		{
			name: "th09でタイプ2",
			candidates: []archiveCandidate{
				{name: "Kaguya", archive: pbgarc.NewKaguyaArchive()},
			},
			gameNum:  9,
			wantName: "Kaguya",
			wantType: 2,
		},
		{
			name:       "Kaguya候補なし",
			candidates: []archiveCandidate{},
			gameNum:    8,
			wantName:   "",
			wantType:   -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, name, archiveType := extractor.chooseKaguya(tt.candidates, tt.gameNum)

			if name != tt.wantName {
				t.Errorf("Expected name %q, got %q", tt.wantName, name)
			}
			if archiveType != tt.wantType {
				t.Errorf("Expected type %d, got %d", tt.wantType, archiveType)
			}
			if tt.wantName == "" && archive != nil {
				t.Error("Expected nil archive but got non-nil")
			}
			if tt.wantName != "" && archive == nil {
				t.Error("Expected archive but got nil")
			}
		})
	}
}

func TestChooseKanako(t *testing.T) {
	logger := config.NewDebugLogger(false)
	extractor := NewExtractor(logger)

	tests := []struct {
		name       string
		candidates []archiveCandidate
		gameNum    int
		wantName   string
		wantType   int
	}{
		{
			name: "th10でタイプ0",
			candidates: []archiveCandidate{
				{name: "Kanako", archive: pbgarc.NewKanakoArchive()},
			},
			gameNum:  10,
			wantName: "Kanako",
			wantType: 0,
		},
		{
			name: "th12でタイプ1",
			candidates: []archiveCandidate{
				{name: "Kanako", archive: pbgarc.NewKanakoArchive()},
			},
			gameNum:  12,
			wantName: "Kanako",
			wantType: 1,
		},
		{
			name: "th13でタイプ2",
			candidates: []archiveCandidate{
				{name: "Kanako", archive: pbgarc.NewKanakoArchive()},
			},
			gameNum:  13,
			wantName: "Kanako",
			wantType: 2,
		},
		{
			name: "th95でタイプ0",
			candidates: []archiveCandidate{
				{name: "Kanako", archive: pbgarc.NewKanakoArchive()},
			},
			gameNum:  95,
			wantName: "Kanako",
			wantType: 0,
		},
		{
			name:       "Kanako候補なし",
			candidates: []archiveCandidate{},
			gameNum:    10,
			wantName:   "",
			wantType:   -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, name, archiveType := extractor.chooseKanako(tt.candidates, tt.gameNum)

			if name != tt.wantName {
				t.Errorf("Expected name %q, got %q", tt.wantName, name)
			}
			if archiveType != tt.wantType {
				t.Errorf("Expected type %d, got %d", tt.wantType, archiveType)
			}
			if tt.wantName == "" && archive != nil {
				t.Error("Expected nil archive but got non-nil")
			}
			if tt.wantName != "" && archive == nil {
				t.Error("Expected archive but got nil")
			}
		})
	}
}

func TestGetKanakoSubType(t *testing.T) {
	logger := config.NewDebugLogger(false)
	extractor := NewExtractor(logger)

	tests := []struct {
		name     string
		gameNum  int
		wantType int
	}{
		{"th10", 10, 0},
		{"th11", 11, 0},
		{"th95", 95, 0},
		{"th12", 12, 1},
		{"th13", 13, 2},
		{"th14", 14, 2},
		{"th15", 15, 2},
		{"th16", 16, 2},
		{"th17", 17, 2},
		{"th18", 18, 2},
		{"th19", 19, 2},
		{"不明な番号", 99, 2},
		{"古いゲーム", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.getKanakoSubType(tt.gameNum)
			if result != tt.wantType {
				t.Errorf("getKanakoSubType(%d) = %d, want %d", tt.gameNum, result, tt.wantType)
			}
		})
	}
}

func TestChooseFromCandidates(t *testing.T) {
	logger := config.NewDebugLogger(false)
	extractor := NewExtractor(logger)

	tests := []struct {
		name       string
		candidates []archiveCandidate
		gameNum    int
		wantName   string
		wantType   int
	}{
		{
			name: "th06でHinanawi選択",
			candidates: []archiveCandidate{
				{name: "Hinanawi", archive: &pbgarc.HinanawiArchive{}},
			},
			gameNum:  6,
			wantName: "Hinanawi",
			wantType: -1,
		},
		{
			name: "th07でYukari選択",
			candidates: []archiveCandidate{
				{name: "Yukari", archive: &pbgarc.YukariArchive{}},
			},
			gameNum:  7,
			wantName: "Yukari",
			wantType: -1,
		},
		{
			name: "th08でKaguya選択",
			candidates: []archiveCandidate{
				{name: "Kaguya", archive: pbgarc.NewKaguyaArchive()},
			},
			gameNum:  8,
			wantName: "Kaguya",
			wantType: 0,
		},
		{
			name: "th10でKanako選択",
			candidates: []archiveCandidate{
				{name: "Kanako", archive: pbgarc.NewKanakoArchive()},
			},
			gameNum:  10,
			wantName: "Kanako",
			wantType: 0,
		},
		{
			name: "複数候補から正しく選択",
			candidates: []archiveCandidate{
				{name: "Hinanawi", archive: &pbgarc.HinanawiArchive{}},
				{name: "Yukari", archive: &pbgarc.YukariArchive{}},
			},
			gameNum:  7,
			wantName: "Yukari",
			wantType: -1,
		},
		{
			name: "該当なし",
			candidates: []archiveCandidate{
				{name: "Kanako", archive: pbgarc.NewKanakoArchive()},
			},
			gameNum:  6,
			wantName: "",
			wantType: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, name, archiveType := extractor.chooseFromCandidates(tt.candidates, tt.gameNum)

			if name != tt.wantName {
				t.Errorf("Expected name %q, got %q", tt.wantName, name)
			}
			if archiveType != tt.wantType {
				t.Errorf("Expected type %d, got %d", tt.wantType, archiveType)
			}
			if tt.wantName == "" && archive != nil {
				t.Error("Expected nil archive but got non-nil")
			}
			if tt.wantName != "" && archive == nil {
				t.Error("Expected archive but got nil")
			}
		})
	}
}
