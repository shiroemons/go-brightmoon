package crypto

// XOR はデータストリームの各バイトを指定されたキーで XOR します。
func XOR(data []byte, key byte) {
	for i := range data {
		data[i] ^= key
	}
}
