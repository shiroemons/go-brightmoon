package crypto

const (
	n         = 624
	m         = 397
	matrixA   = 0x9908b0df
	upperMask = 0x80000000
	lowerMask = 0x7fffffff
)

// RNGMT はメルセンヌ・ツイスタ (MT19937) 疑似乱数生成器です。
// C++版の RNG_MT を移植したものです。
type RNGMT struct {
	mt  [n]uint32
	mti int
}

// NewRNGMT は指定されたシードで RNGMT を初期化して返します。
func NewRNGMT(seed uint32) *RNGMT {
	r := &RNGMT{}
	r.init(seed)
	return r
}

// init は指定されたシードで RNGMT を初期化します。
func (r *RNGMT) init(seed uint32) {
	r.mt[0] = seed
	for r.mti = 1; r.mti < n; r.mti++ {
		r.mt[r.mti] = (1812433253*(r.mt[r.mti-1]^(r.mt[r.mti-1]>>30)) + uint32(r.mti))
	}
}

// NextInt32 は次の32ビット符号なし乱数を生成して返します。
func (r *RNGMT) NextInt32() uint32 {
	var y uint32
	mag01 := [2]uint32{0x0, matrixA}

	if r.mti >= n {
		var kk int

		for kk = 0; kk < n-m; kk++ {
			y = (r.mt[kk] & upperMask) | (r.mt[kk+1] & lowerMask)
			r.mt[kk] = r.mt[kk+m] ^ (y >> 1) ^ mag01[y&0x1]
		}
		for ; kk < n-1; kk++ {
			y = (r.mt[kk] & upperMask) | (r.mt[kk+1] & lowerMask)
			r.mt[kk] = r.mt[kk+(m-n)] ^ (y >> 1) ^ mag01[y&0x1]
		}
		y = (r.mt[n-1] & upperMask) | (r.mt[0] & lowerMask)
		r.mt[n-1] = r.mt[m-1] ^ (y >> 1) ^ mag01[y&0x1]

		r.mti = 0
	}

	y = r.mt[r.mti]
	r.mti++

	// Tempering
	y ^= (y >> 11)
	y ^= (y << 7) & 0x9d2c5680
	y ^= (y << 15) & 0xefc60000
	y ^= (y >> 18)

	return y
}
