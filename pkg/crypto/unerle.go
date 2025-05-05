package crypto

import (
	"io"
)

// UneRLE は RLE (Run-Length Encoding) 圧縮されたデータを展開します。
// C++版の unerle ロジックを再現します。
func UneRLE(in io.Reader, out io.Writer) error {
	br := make([]byte, 1)
	var prev, pprev byte
	firstByte := true
	secondByte := true

	for {
		n, err := in.Read(br)
		if n == 0 {
			if err == io.EOF {
				break // 正常終了
			}
			return err // その他のエラー
		}
		if err != nil && err != io.EOF {
			return err
		}

		current := br[0]

		if firstByte {
			if _, err := out.Write([]byte{current}); err != nil {
				return err
			}
			pprev = current
			firstByte = false
		} else if secondByte {
			if _, err := out.Write([]byte{current}); err != nil {
				return err
			}
			prev = current
			secondByte = false
		} else {
			if _, err := out.Write([]byte{current}); err != nil {
				return err
			}
			if prev == pprev {
				// 同じ文字が続いたので、次のバイトはカウント
				n, err := in.Read(br)
				if n == 0 || (err != nil && err != io.EOF) {
					// カウントが読めないのはエラー (EOF含む)
					if err == io.EOF {
						return io.ErrUnexpectedEOF
					}
					return err
				}
				count := int(br[0])
				for i := 0; i < count; i++ {
					if _, err := out.Write([]byte{prev}); err != nil {
						return err
					}
				}
				// pprev, prev は次のループのために更新されるので、ここでは何もしない
				// prev は次のループの current になる
			}
			// prev と pprev を更新
			pprev = prev
			prev = current
		}

		if err == io.EOF {
			break
		}
	}
	return nil
}
