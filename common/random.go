package common

type Random struct {
	mag01 []uint32
	mt    []uint32
	mti   int
}

func NewRandom(seed int) *Random {
	r := &Random{}
	r.initGenRand(uint32(seed))
	return r
}

func (r *Random) Next() int {
	return int(r.genRandInt31())
}

func (r *Random) NextRange(minValue, maxValue int) int {
	if minValue == maxValue {
		return minValue
	}

	if minValue > maxValue {
		minValue, maxValue = maxValue, minValue
	}

	rang := maxValue - minValue

	return minValue + r.Next()%rang
}

const (
	N         = 624
	M         = 397
	MatrixA   = 0x9908b0df
	UpperMask = 0x80000000
	LowerMask = 0x7fffffff
)

func (r *Random) initGenRand(s uint32) {
	r.mt = make([]uint32, N)
	r.mag01 = []uint32{0x0, MatrixA}
	r.mt[0] = s & 0xffffffff
	for r.mti = 1; r.mti < N; r.mti++ {
		r.mt[r.mti] = 1812433253*(r.mt[r.mti-1]^(r.mt[r.mti-1]>>30)) + uint32(r.mti)
		r.mt[r.mti] &= 0xffffffff
	}
}

func (r *Random) genRandInt32() uint32 {
	var y uint32
	if r.mti >= N {
		var kk int
		if r.mti == N+1 {
			r.initGenRand(5489)
		}
		for kk = 0; kk < N-M; kk++ {
			y = (r.mt[kk] & UpperMask) | (r.mt[kk+1] & LowerMask)
			r.mt[kk] = r.mt[kk+M] ^ (y >> 1) ^ r.mag01[y&0x1]
		}
		for ; kk < N-1; kk++ {
			y = (r.mt[kk] & UpperMask) | (r.mt[kk+1] & LowerMask)
			r.mt[kk] = r.mt[kk+(M-N)] ^ (y >> 1) ^ r.mag01[y&0x1]
		}
		y = (r.mt[N-1] & UpperMask) | (r.mt[0] & LowerMask)
		r.mt[N-1] = r.mt[M-1] ^ (y >> 1) ^ r.mag01[y&0x1]
		r.mti = 0
	}
	y = r.mt[r.mti]
	r.mti++
	y ^= y >> 11
	y ^= (y << 7) & 0x9d2c5680
	y ^= (y << 15) & 0xefc60000
	y ^= y >> 18
	return y
}

func (r *Random) genRandInt31() int32 {
	return int32(r.genRandInt32() >> 1)
}
