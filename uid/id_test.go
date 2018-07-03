package uid

import (
	"testing"
)

func BenchmarkGenId(b *testing.B) {
	conflict := make(map[string]int)
	InitGenerator(12, 5)
	for i := 0; i < b.N; i++ {
		id, err := GenId()
		if err != nil {
			b.Fatal(err)
		}
		if _, ok := conflict[id]; ok {
			b.Fatal("conflict ", i)
		} else {
			conflict[id] = 1
		}
	}
}
