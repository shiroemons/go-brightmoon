package crypto

import (
	"bytes"
	"io"
	"testing"
)

func TestUNLZSS_TerminatorOnly(t *testing.T) {
	// 終端オフセット0のみのデータ
	// flag=0 (1ビット) + offset=0 (13ビット) = 14ビット = 0b00000000000000
	// バイト表現: 0x00, 0x00
	input := []byte{0x00, 0x00}
	in := bytes.NewReader(input)
	out := &bytes.Buffer{}

	err := UNLZSS(in, out)
	if err != nil {
		t.Errorf("UNLZSS() error = %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("出力が空であるべき: got %d bytes", out.Len())
	}
}

func TestUNLZSS_SingleLiteral(t *testing.T) {
	// 1バイトのリテラル + 終端
	// flag=1 (1ビット) + literal=0x41 (8ビット) = 9ビット
	// flag=0 (1ビット) + offset=0 (13ビット) = 14ビット
	// 合計23ビット
	// バイト: 10100000 1_0000000 00000000
	//         0xA0      0x80      0x00
	// 実際: 1 01000001 0 0000000000000
	// = 1010 0000 | 1000 0000 | 0000 0...
	input := []byte{0xA0, 0x80, 0x00}
	in := bytes.NewReader(input)
	out := &bytes.Buffer{}

	err := UNLZSS(in, out)
	if err != nil {
		t.Errorf("UNLZSS() error = %v", err)
	}
	expected := []byte{0x41}
	if !bytes.Equal(out.Bytes(), expected) {
		t.Errorf("UNLZSS() = %v, want %v", out.Bytes(), expected)
	}
}

func TestUNLZSS_EmptyInput(t *testing.T) {
	input := []byte{}
	in := bytes.NewReader(input)
	out := &bytes.Buffer{}

	err := UNLZSS(in, out)
	if err != io.ErrUnexpectedEOF {
		t.Errorf("UNLZSS() error = %v, want io.ErrUnexpectedEOF", err)
	}
}

func TestUNLZSS_MultipleLiterals(t *testing.T) {
	// 複数のリテラル: 'A', 'B' + 終端
	// 1 01000001 | 1 01000010 | 0 0000000000000
	// = 10100000 | 11010000 | 10_00000 | 00000000
	// = 0xA0, 0xD0, 0x80, 0x00
	input := []byte{0xA0, 0xD0, 0x80, 0x00}
	in := bytes.NewReader(input)
	out := &bytes.Buffer{}

	err := UNLZSS(in, out)
	if err != nil {
		t.Errorf("UNLZSS() error = %v", err)
	}
	expected := []byte{0x41, 0x42}
	if !bytes.Equal(out.Bytes(), expected) {
		t.Errorf("UNLZSS() = %v, want %v", out.Bytes(), expected)
	}
}
