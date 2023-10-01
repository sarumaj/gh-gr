package util

import (
	"math"
	"runtime"
)

func GetIdealConcurrency() uint {
	con := float64(runtime.NumCPU() * 2)
	con = math.Max(con, 4)
	return uint(con)
}
