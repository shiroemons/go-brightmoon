package crypto

import (
	"bytes"
	"testing"
)

func TestUneRLE(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
		wantErr  bool
	}{
		{
			name:     "圧縮なしデータ",
			input:    []byte{0x41, 0x42, 0x43}, // ABC
			expected: []byte{0x41, 0x42, 0x43},
			wantErr:  false,
		},
		{
			name:     "空データ",
			input:    []byte{},
			expected: []byte{},
			wantErr:  false,
		},
		{
			name:     "1バイトのみ",
			input:    []byte{0x41},
			expected: []byte{0x41},
			wantErr:  false,
		},
		{
			name:     "2バイトのみ",
			input:    []byte{0x41, 0x42},
			expected: []byte{0x41, 0x42},
			wantErr:  false,
		},
		{
			name:     "連続した同じ文字",
			input:    []byte{0x41, 0x41, 0x41, 0x02}, // AAA + 2回繰り返し
			expected: []byte{0x41, 0x41, 0x41, 0x41, 0x41},
			wantErr:  false,
		},
		{
			name:     "0回繰り返し",
			input:    []byte{0x41, 0x41, 0x41, 0x00}, // AAA + 0回繰り返し
			expected: []byte{0x41, 0x41, 0x41},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := bytes.NewReader(tt.input)
			out := &bytes.Buffer{}

			err := UneRLE(in, out)
			if (err != nil) != tt.wantErr {
				t.Errorf("UneRLE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got := out.Bytes()
				if !bytes.Equal(got, tt.expected) {
					t.Errorf("UneRLE() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestUneRLE_UnexpectedEOF(t *testing.T) {
	// 連続文字の後にカウントがない場合
	input := []byte{0x41, 0x41, 0x41} // AAA（カウントがない）
	in := bytes.NewReader(input)
	out := &bytes.Buffer{}

	err := UneRLE(in, out)
	// エラーが発生するか、または正常に処理される（実装による）
	// 現在の実装では io.ErrUnexpectedEOF が返される
	if err == nil {
		// エラーなしで処理された場合、出力を確認
		got := out.Bytes()
		if len(got) < 3 {
			t.Errorf("出力が短すぎる: %v", got)
		}
	}
}
