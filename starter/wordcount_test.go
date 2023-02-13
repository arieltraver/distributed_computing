package main

import (
	"testing"
)

func BenchmarkSingleThreaded(b *testing.B) {
    for i := 0; i < b.N; i++ {
        single_threaded([]string{"input/big.txt", "input/book2.txt"})
    }
}

func BenchmarkMultiThreaded(b *testing.B) {
    for i := 0; i < b.N; i++ {
        multi_threaded([]string{"input/big.txt", "input/book2.txt"})
    }
}
