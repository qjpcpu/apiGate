package rr

import (
	"fmt"
	"testing"
)

func TestScale(t *testing.T) {
	stats := func(raw, arr []int) {
		m := make(map[int]int)
		for _, v := range arr {
			m[v] += 1
		}
		fmt.Println("total length:", len(arr))
		for k, v := range m {
			fmt.Printf("%v:%v\n", raw[k], v)
		}
		fmt.Println(arr)
	}
	weights := []int{0, 30, 30}
	scale := MakeWeightsScale(weights)
	stats(weights, scale)
}
