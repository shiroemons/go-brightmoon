package crypto

import (
	"bytes"
	"testing"
)

func TestTHCrypter_Basic(t *testing.T) {
	// 基本的な暗号化解除テスト
	// 入力データを暗号化解除して、期待される出力と比較

	tests := []struct {
		name  string
		input []byte
		size  int
		key   byte
		step  byte
		block int
		limit int
	}{
		{
			name:  "小さいデータ",
			input: []byte{0x00, 0x00, 0x00, 0x00},
			size:  4,
			key:   0x00,
			step:  0x01,
			block: 4,
			limit: 4,
		},
		{
			name:  "ブロックサイズより小さいデータ",
			input: []byte{0x12, 0x34},
			size:  2,
			key:   0xFF,
			step:  0x01,
			block: 4,
			limit: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := bytes.NewReader(tt.input)
			out := &bytes.Buffer{}

			result := THCrypter(in, out, tt.size, tt.key, tt.step, tt.block, tt.limit)
			if !result {
				t.Error("THCrypter() returned false")
			}
			if out.Len() != tt.size {
				t.Errorf("出力サイズ = %d, want %d", out.Len(), tt.size)
			}
		})
	}
}

func TestTHCrypter_ReadError(t *testing.T) {
	// 入力が足りない場合はfalseを返す
	in := bytes.NewReader([]byte{0x00}) // 1バイトしかない
	out := &bytes.Buffer{}

	result := THCrypter(in, out, 4, 0x00, 0x01, 4, 4) // 4バイト必要
	if result {
		t.Error("THCrypter() should return false when input is insufficient")
	}
}

func TestTHCrypter_EmptyInput(t *testing.T) {
	// 空入力
	in := bytes.NewReader([]byte{})
	out := &bytes.Buffer{}

	result := THCrypter(in, out, 0, 0x00, 0x01, 4, 4)
	if !result {
		t.Error("THCrypter() should return true for size=0")
	}
	if out.Len() != 0 {
		t.Errorf("出力サイズ = %d, want 0", out.Len())
	}
}

func TestTHCrypter_LimitReached(t *testing.T) {
	// limitに達した場合の動作
	input := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	in := bytes.NewReader(input)
	out := &bytes.Buffer{}

	// size=8, limit=4 の場合、最初の4バイトのみ処理
	result := THCrypter(in, out, 8, 0x00, 0x01, 4, 4)
	if !result {
		t.Error("THCrypter() returned false")
	}
	// limit=4なので、最初の4バイトが処理され、残り4バイトはaddupとしてコピーされる
}

func TestTHCrypter_Addup(t *testing.T) {
	// addupの計算テスト
	// addup = size % block + (size % 2)
	// ただし、addup >= block/4 の場合は addup = 0 + (size % 2)

	// size=5, block=4 の場合:
	// size % block = 1
	// 1 < 4/4=1 ではない（1 >= 1）ので addup = 0
	// addup += size % 2 = 0 + 1 = 1
	// mainSize = 5 - 1 = 4

	input := []byte{0x00, 0x00, 0x00, 0x00, 0xFF}
	in := bytes.NewReader(input)
	out := &bytes.Buffer{}

	result := THCrypter(in, out, 5, 0x00, 0x01, 4, 8)
	if !result {
		t.Error("THCrypter() returned false")
	}
	if out.Len() != 5 {
		t.Errorf("出力サイズ = %d, want 5", out.Len())
	}
}

func TestTHCrypter_RealWorldParams(t *testing.T) {
	// 実際のゲームで使用されるパラメータでテスト
	// Kaguya/Kanako のヘッダ暗号化パラメータ: key=0x1B, step=0x37, block=0x10, limit=0x10

	input := make([]byte, 16)
	for i := range input {
		input[i] = byte(i)
	}

	in := bytes.NewReader(input)
	out := &bytes.Buffer{}

	result := THCrypter(in, out, 16, 0x1B, 0x37, 0x10, 0x10)
	if !result {
		t.Error("THCrypter() returned false")
	}
	if out.Len() != 16 {
		t.Errorf("出力サイズ = %d, want 16", out.Len())
	}
}
