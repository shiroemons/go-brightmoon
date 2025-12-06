package crypto

import (
	"testing"
)

func TestYumemiCrypt(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		key            byte
		expected       []byte
		expectedNewKey byte
	}{
		{
			name:           "単一バイト",
			input:          []byte{0x00},
			key:            0x10,
			expected:       []byte{0x10}, // 0x00 ^ 0x10 = 0x10
			expectedNewKey: 0x10 + 0x51,
		},
		{
			name:           "複数バイト",
			input:          []byte{0x00, 0x00, 0x00},
			key:            0x00,
			expected:       []byte{0x00, 0x51, 0xA2}, // 0^0, 0^0x51, 0^0xA2
			expectedNewKey: 0xA2 + 0x51,
		},
		{
			name:           "空データ",
			input:          []byte{},
			key:            0xFF,
			expected:       []byte{},
			expectedNewKey: 0xFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, len(tt.input))
			copy(data, tt.input)
			newKey := YumemiCrypt(data, tt.key)

			if len(data) != len(tt.expected) {
				t.Errorf("長さが異なる: got %d, want %d", len(data), len(tt.expected))
				return
			}
			for i := range data {
				if data[i] != tt.expected[i] {
					t.Errorf("data[%d] = 0x%02X, want 0x%02X", i, data[i], tt.expected[i])
				}
			}
			if newKey != tt.expectedNewKey {
				t.Errorf("newKey = 0x%02X, want 0x%02X", newKey, tt.expectedNewKey)
			}
		})
	}
}

func TestYumemiCrypt_RoundTrip(t *testing.T) {
	// YumemiCrypt を2回適用すると元に戻ることを確認
	original := []byte{0x12, 0x34, 0x56, 0x78}
	data := make([]byte, len(original))
	copy(data, original)

	key := byte(0xAB)
	YumemiCrypt(data, key)
	YumemiCrypt(data, key)

	for i := range data {
		if data[i] != original[i] {
			t.Errorf("data[%d] = 0x%02X, want 0x%02X", i, data[i], original[i])
		}
	}
}
