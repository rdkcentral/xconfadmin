package util

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func RandomDouble() float64 {
	return rand.Float64()
}

func RandomPercentage() int {
	return rand.Intn(100)
}
