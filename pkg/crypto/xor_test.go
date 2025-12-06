package crypto

import (
	"testing"
)

func TestXOR(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		key      byte
		expected []byte
	}{
		{
			name:     "単一バイト",
			input:    []byte{0x00},
			key:      0xFF,
			expected: []byte{0xFF},
		},
		{
			name:     "複数バイト",
			input:    []byte{0x00, 0xFF, 0xAA, 0x55},
			key:      0xFF,
			expected: []byte{0xFF, 0x00, 0x55, 0xAA},
		},
		{
			name:     "キー0x00",
			input:    []byte{0x12, 0x34, 0x56},
			key:      0x00,
			expected: []byte{0x12, 0x34, 0x56},
		},
		{
			name:     "空データ",
			input:    []byte{},
			key:      0xFF,
			expected: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, len(tt.input))
			copy(data, tt.input)
			XOR(data, tt.key)
			if len(data) != len(tt.expected) {
				t.Errorf("長さが異なる: got %d, want %d", len(data), len(tt.expected))
				return
			}
			for i := range data {
				if data[i] != tt.expected[i] {
					t.Errorf("data[%d] = 0x%02X, want 0x%02X", i, data[i], tt.expected[i])
				}
			}
		})
	}
}

func TestXOR_RoundTrip(t *testing.T) {
	// XOR を2回適用すると元に戻ることを確認
	original := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	data := make([]byte, len(original))
	copy(data, original)

	XOR(data, 0xAB)
	XOR(data, 0xAB)

	for i := range data {
		if data[i] != original[i] {
			t.Errorf("data[%d] = 0x%02X, want 0x%02X", i, data[i], original[i])
		}
	}
}
