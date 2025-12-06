// Package crypto は東方Projectのアーカイブファイルで使用される暗号化・圧縮アルゴリズムを提供します。
//
// 主な機能:
//   - THCrypter: 東方Project特有のXORベース暗号化の解除
//   - UNLZSS: LZSS圧縮データの解凍
//   - UNERLE: RLE圧縮データの解凍
//   - XOR: 単純なXOR暗号化
//   - RNGMT: メルセンヌ・ツイスタ疑似乱数生成器
package crypto

import (
	"io"
)

// THCrypter は東方Project特有の暗号化を解除する関数です
// C++版のロジックをGoで再現 (再確認・修正)
// in: 入力ストリーム
// out: 出力ストリーム
// size: 元のサイズ
// key: 暗号化キー
// step: 暗号化ステップ
// block: ブロックサイズ
// limit: 制限サイズ
func THCrypter(in io.Reader, out io.Writer, size int, key byte, step byte, block int, limit int) bool {
	inBuf := make([]byte, block)
	outBuf := make([]byte, block)

	// addup の計算 (C++版と同じ)
	addup := size % block
	if addup >= block/4 {
		addup = 0
	}
	addup += size % 2

	// 実際に処理するメイン部分のサイズ
	mainSize := size - addup

	remainingSize := mainSize
	remainingLimit := limit
	currentKey := key // key はブロック間で引き継がれる

	for remainingSize > 0 && remainingLimit > 0 {
		// このブロックで処理するサイズを決定
		processBlockSize := block // 基本ブロックサイズ
		if remainingSize < processBlockSize {
			processBlockSize = remainingSize
		}
		if remainingLimit < processBlockSize {
			processBlockSize = remainingLimit
		}

		// データ読み込み
		n, err := io.ReadFull(in, inBuf[:processBlockSize])
		if err != nil || n != processBlockSize {
			return false // 読み込み失敗
		}

		// C++版の暗号化解除ロジック
		pin := 0 // inBuf の読み取りインデックス
		// processBlockSize 分だけ処理する
		for j := 0; j < 2; j++ {
			pout := processBlockSize - j - 1 // outBuf の書き込みインデックス
			// 内側ループの回数も processBlockSize に基づく
			for i := 0; i < (processBlockSize-j+1)/2; i++ {
				if pout >= 0 && pout < len(outBuf) && pin < len(inBuf) {
					// グローバルに更新されるキーで XOR
					outBuf[pout] = inBuf[pin] ^ currentKey
				}
				pin++
				pout -= 2
				currentKey += step // キーを更新
			}
		}

		// 処理したブロックを書き込み
		if _, err := out.Write(outBuf[:processBlockSize]); err != nil {
			return false // 書き込み失敗
		}

		remainingLimit -= processBlockSize
		remainingSize -= processBlockSize
	}

	// 残りの addup バイトをそのままコピー
	if addup > 0 {
		restBuf := make([]byte, addup)
		n, err := io.ReadFull(in, restBuf)
		if err != nil || n != addup { // 厳密にチェック
			return false
		}
		if _, err := out.Write(restBuf); err != nil {
			return false
		}
	}

	return true
}
