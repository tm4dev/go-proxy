package utils

import (
	"testing"
)

func BenchmarkRandomInt(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			RandomInt(256)
		}
	})
}
