package shared

import (
	"math"

	"github.com/dchest/siphash"
)

const (
	SipHashKey0 = uint64(506097522914230528)
	SipHashKey1 = uint64(1084818905618843912)
)

func CalculateHashAndPercent(value string) (float64, float64) {
	voffset := float64(math.MaxInt64 + 1)
	vrange := float64(math.MaxInt64*2 + 1)
	bbytes := []byte(value)
	hashCode := float64(int64(siphash.Hash(SipHashKey0, SipHashKey1, bbytes))) + voffset
	percent := (hashCode / vrange) * 100
	return hashCode, percent
}

func CalculatePercent(value string) int {
	_, calculated := CalculateHashAndPercent(value)
	calculated = math.Round(calculated)
	return int(calculated)
}
