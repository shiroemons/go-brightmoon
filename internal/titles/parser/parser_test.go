package parser

import (
	"testing"
)

func TestTHBGMParser_ParseTHFmt(t *testing.T) {
	parser := NewTHBGMParser()

	// テスト用の最小限のデータ（52バイト）
	testData := make([]byte, 52)
	// ファイル名 "test.wav" を設定
	copy(testData[0:8], []byte("test.wav"))
	// 開始位置、イントロ、長さを設定（リトルエンディアン）
	testData[16] = 0x10 // start = 0x00000010
	testData[24] = 0x05 // intro = 0x00000005
	testData[28] = 0x0A // length = 0x0000000A

	records, err := parser.ParseTHFmt(testData)
	if err != nil {
		t.Fatalf("ParseTHFmt failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	record := records[0]
	if record.FileName != "test.wav" {
		t.Errorf("Expected filename 'test.wav', got '%s'", record.FileName)
	}
	if record.Start != "00000010" {
		t.Errorf("Expected start '00000010', got '%s'", record.Start)
	}
	if record.Intro != "00000005" {
		t.Errorf("Expected intro '00000005', got '%s'", record.Intro)
	}
	if record.Loop != "00000005" { // length - intro = 10 - 5 = 5
		t.Errorf("Expected loop '00000005', got '%s'", record.Loop)
	}
	if record.Length != "0000000A" {
		t.Errorf("Expected length '0000000A', got '%s'", record.Length)
	}
}

func TestTHBGMParser_ParseMusicCmt(t *testing.T) {
	parser := NewTHBGMParser()

	// テスト用のShift-JISエンコードされたデータ
	// Shift-JISエンコードされた "@bgm/test\n♪Test Track\n"
	testData := []byte{
		0x40, 0x62, 0x67, 0x6d, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x0a, // @bgm/test\n
		0x81, 0xF4, // ♪ in Shift-JIS
		0x54, 0x65, 0x73, 0x74, 0x20, 0x54, 0x72, 0x61, 0x63, 0x6b, 0x0a, // Test Track\n
	}

	tracks, err := parser.ParseMusicCmt(string(testData))
	if err != nil {
		t.Fatalf("ParseMusicCmt failed: %v", err)
	}

	if len(tracks) != 1 {
		t.Fatalf("Expected 1 track, got %d", len(tracks))
	}

	track := tracks[0]
	if track.FileName != "test.wav" {
		t.Errorf("Expected filename 'test.wav', got '%s'", track.FileName)
	}
	if track.Title != "Test Track" {
		t.Errorf("Expected title 'Test Track', got '%s'", track.Title)
	}
}

func TestToHex(t *testing.T) {
	tests := []struct {
		input    uint32
		expected string
	}{
		{0, "00000000"},
		{16, "00000010"},
		{255, "000000FF"},
		{4096, "00001000"},
		{0xDEADBEEF, "DEADBEEF"},
	}

	for _, test := range tests {
		result := toHex(test.input)
		if result != test.expected {
			t.Errorf("toHex(%d) = %s; want %s", test.input, result, test.expected)
		}
	}
}
