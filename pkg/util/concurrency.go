package util

import (
	"math"
	"runtime"
)

func GetIdealConcurrency() uint {
	return uint(math.Max(float64(runtime.NumCPU()*2), 4))
}
