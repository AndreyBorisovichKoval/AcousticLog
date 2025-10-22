// C:\_Projects_Go\AcousticLog\internal\mathx\audiolevel.go

package mathx

import "math"

func BytesToInt16LE(b []byte) []int16 {
	n := len(b) / 2
	out := make([]int16, n)
	for i := 0; i < n; i++ {
		out[i] = int16(b[2*i]) | int16(b[2*i+1])<<8
	}
	return out
}

func CalcRMSInt16(s []int16) float64 {
	if len(s) == 0 {
		return 0
	}
	var sum float64
	for _, v := range s {
		x := float64(v) / 32768.0
		sum += x * x
	}
	return math.Sqrt(sum / float64(len(s)))
}
