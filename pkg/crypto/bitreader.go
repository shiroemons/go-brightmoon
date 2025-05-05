package crypto

import (
	"fmt"
	"io"
)

// BitReader は io.Reader からビット単位でデータを読み込みます。
type BitReader struct {
	reader io.Reader
	buffer byte
	count  uint // 現在のバッファ内のビット数 (0-8)
}

// NewBitReader は新しい BitReader を作成します。
func NewBitReader(r io.Reader) *BitReader {
	return &BitReader{
		reader: r,
		buffer: 0,
		count:  0,
	}
}

// Read は指定されたビット数を読み込み、その値を int として返します。
// エラーが発生した場合でも、それまでに読み込めたビットから構成される値を返します。
// EOFの場合、読み込めた値と io.EOF を返します。
func (br *BitReader) Read(numBits uint) (int, error) {
	if numBits == 0 || numBits > 32 {
		return 0, fmt.Errorf("invalid number of bits to read: %d", numBits)
	}

	value := 0
	var finalError error // ループ内で発生した最後のエラー (主にEOF)

	for i := uint(0); i < numBits; i++ {
		if br.count == 0 {
			buf := make([]byte, 1)
			n, err := br.reader.Read(buf)
			if n == 0 {
				if err == io.EOF {
					finalError = io.EOF // EOFを記録してループを抜ける
					break
				}
				// 0バイト読み込みでEOF以外は予期しないエラー
				return value, fmt.Errorf("read 0 bytes: %w", err)
			}
			if err != nil && err != io.EOF {
				return value, err // EOF以外の読み込みエラーは即時リターン
			}
			// n > 0 の場合 (err が nil または io.EOF でもバイトは読めている)
			br.buffer = buf[0]
			br.count = 8
			// 読み込み中にEOFが発生した場合、それを記録しておく
			if err == io.EOF {
				finalError = io.EOF
				// EOF だがバイトは読めたので続行、ただし次のループで読める保証はない
			}
		}

		// count > 0 の場合、または EOF 前に1バイト読めた場合
		bit := (br.buffer >> 7) & 1
		value = (value << 1) | int(bit)
		br.buffer <<= 1
		br.count--

		// 要求ビット数読む前にEOFになった場合 (ループの最後でチェック)
		if finalError == io.EOF && i < numBits-1 {
			break // 記録されたEOFでループ中断
		}
	}

	return value, finalError // 読み込めた値と、発生したEOF(なければnil)を返す
}
