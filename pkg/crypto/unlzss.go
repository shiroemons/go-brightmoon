package crypto

import (
	"io"
)

const (
	// C++版に合わせた定数
	DictSize = 0x2000 // 8192
)

// UNLZSS はLZSS圧縮されたデータを解凍します
// C++版のロジックをGoで再現
// in: 入力ストリーム
// out: 出力ストリーム
func UNLZSS(in io.Reader, out io.Writer) error {
	dict := make([]byte, DictSize)
	dictPos := 1 // C++版の dictop に相当

	// 辞書はゼロで初期化されている必要はない (C++版のmemsetは古いデータを消すため)

	reader := NewBitReader(in)

	for {
		// フラグを1ビット読み込む
		flag, err := reader.Read(1)
		if err != nil {
			// C++版はEOFなどでループを抜ける条件がないため、エラー時にループを終了する
			// patofs == 0 のチェックで抜けるのが唯一の正常終了パターン
			if err == io.EOF {
				// EOF はエラーではなく、正常終了の可能性がある (patofs==0で検出)
				// ただし、フラグ読み込みでのEOFは通常予期しない
				return io.ErrUnexpectedEOF
			}
			return err
		}

		if flag == 1 {
			// 非圧縮データ (8ビット)
			c, err := reader.Read(8)
			if err != nil {
				return err
			}

			b := byte(c)
			_, err = out.Write([]byte{b})
			if err != nil {
				return err
			}

			dict[dictPos] = b
			dictPos = (dictPos + 1) % DictSize
		} else {
			// 圧縮データ (オフセット13ビット + 長さ4ビット)
			patOfs, err := reader.Read(13)

			// C++版の終了条件: オフセットが0。EOFエラーより優先してチェック。
			if patOfs == 0 {
				// 正常に終端オフセット0を読み込めた (直後にEOFかもしれないが問題ない)
				return nil // 正常終了
			}

			// 終端オフセット0でなかったので、エラーが発生していたかチェック
			if err != nil {
				// 終端ではないのに読み込みエラー(EOF含む)が発生したのは異常
				if err == io.EOF {
					// データが期待される終端オフセット0より前に尽きた
					return io.ErrUnexpectedEOF
				}
				return err // その他の読み取りエラー
			}

			// オフセットが0でなく、エラーもなかった場合、長さを読む
			patLen, err := reader.Read(4)
			if err != nil {
				// 長さを読んでいる途中でEOFになるのは異常
				if err == io.EOF {
					return io.ErrUnexpectedEOF
				}
				return err
			}
			patLen += 3 // 長さは+3する

			for i := 0; i < patLen; i++ {
				c := dict[(patOfs+i)%DictSize]
				_, err = out.Write([]byte{c})
				if err != nil {
					return err
				}
				dict[dictPos] = c
				dictPos = (dictPos + 1) % DictSize
			}
		}
	}
	// 通常はループ内の patofs == 0 で return する
}
