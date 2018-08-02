package rr

import (
	"math/rand"
)

const (
	ScaleLen = 100
)

func MakeWeightsScale(weights []int) []int {
	scale := make([]int, ScaleLen)
	total := 0
	for _, w := range weights {
		total += w
	}
	if total == 0 {
		panic("no weights")
	}
	cursor := 0
	for i := range weights {
		var weight int
		if i == len(weights)-1 {
			weight = ScaleLen - cursor
		} else {
			weight = weights[i] * ScaleLen / total
		}
		for w := weight; w > 0; w-- {
			if w == 0 {
				break
			}
			scale[cursor] = i
			cursor++
		}
	}
	ChaosInts(scale)
	return scale
}

func ChaosInts(arr []int) []int {
	size := len(arr)
	for i := 0; i < size-1; i++ {
		n := rand.Intn(100) % (size - i - 1)
		arr[n], arr[size-i-1] = arr[size-i-1], arr[n]
	}
	return arr
}
