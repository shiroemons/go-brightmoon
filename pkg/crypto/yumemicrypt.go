package crypto

// YumemiCrypt はデータストリームの各バイトをキーで XOR し、キーを更新します。
func YumemiCrypt(data []byte, key byte) byte {
	currentKey := key
	for i := range data {
		data[i] ^= currentKey
		currentKey += 0x51
	}
	return currentKey // 更新されたキーを返す (必要なら)
}
