package crypto

import (
	"bytes"
	"io"
	"testing"
)

func TestBitReader_Read(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		numBits  uint
		expected int
		wantErr  bool
	}{
		{
			name:     "1ビット読み込み（MSB=1）",
			data:     []byte{0x80}, // 10000000
			numBits:  1,
			expected: 1,
			wantErr:  false,
		},
		{
			name:     "1ビット読み込み（MSB=0）",
			data:     []byte{0x7F}, // 01111111
			numBits:  1,
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "8ビット読み込み",
			data:     []byte{0xAB},
			numBits:  8,
			expected: 0xAB,
			wantErr:  false,
		},
		{
			name:     "4ビット読み込み",
			data:     []byte{0xF0}, // 11110000
			numBits:  4,
			expected: 0x0F, // 上位4ビット = 1111
			wantErr:  false,
		},
		{
			name:     "13ビット読み込み",
			data:     []byte{0xFF, 0x80}, // 1111111110000000
			numBits:  13,
			expected: 0x1FF0, // 最上位13ビット = 1111111110000
			wantErr:  false,
		},
		{
			name:     "空データで読み込み",
			data:     []byte{},
			numBits:  1,
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewBitReader(bytes.NewReader(tt.data))
			got, err := reader.Read(tt.numBits)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("Read() = 0x%X, want 0x%X", got, tt.expected)
			}
		})
	}
}

func TestBitReader_InvalidNumBits(t *testing.T) {
	reader := NewBitReader(bytes.NewReader([]byte{0xFF}))

	// 0ビット
	_, err := reader.Read(0)
	if err == nil {
		t.Error("Read(0) should return error")
	}

	// 33ビット以上
	reader2 := NewBitReader(bytes.NewReader([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF}))
	_, err = reader2.Read(33)
	if err == nil {
		t.Error("Read(33) should return error")
	}
}

func TestBitReader_SequentialReads(t *testing.T) {
	// 0b10110011 = 0xB3
	data := []byte{0xB3}
	reader := NewBitReader(bytes.NewReader(data))

	// 1ビット読み込み: 1
	val, err := reader.Read(1)
	if err != nil {
		t.Fatalf("Read(1) error: %v", err)
	}
	if val != 1 {
		t.Errorf("Read(1) = %d, want 1", val)
	}

	// 3ビット読み込み: 011 = 3
	val, err = reader.Read(3)
	if err != nil {
		t.Fatalf("Read(3) error: %v", err)
	}
	if val != 0b011 {
		t.Errorf("Read(3) = %d, want 3", val)
	}

	// 4ビット読み込み: 0011 = 3
	val, err = reader.Read(4)
	if err != nil {
		t.Fatalf("Read(4) error: %v", err)
	}
	if val != 0b0011 {
		t.Errorf("Read(4) = %d, want 3", val)
	}
}

func TestBitReader_CrossByteBoundary(t *testing.T) {
	// 2バイトにまたがって読み込み
	data := []byte{0xFF, 0x00} // 11111111 00000000
	reader := NewBitReader(bytes.NewReader(data))

	// 最初の12ビット: 111111110000 = 0xFF0
	val, err := reader.Read(12)
	if err != nil {
		t.Fatalf("Read(12) error: %v", err)
	}
	if val != 0xFF0 {
		t.Errorf("Read(12) = 0x%X, want 0xFF0", val)
	}
}

func TestBitReader_EOF(t *testing.T) {
	data := []byte{0xAB}
	reader := NewBitReader(bytes.NewReader(data))

	// 8ビット読み込み
	_, err := reader.Read(8)
	if err != nil {
		t.Fatalf("Read(8) error: %v", err)
	}

	// 次の読み込みでEOF
	_, err = reader.Read(1)
	if err != io.EOF {
		t.Errorf("Read(1) after EOF = %v, want io.EOF", err)
	}
}
