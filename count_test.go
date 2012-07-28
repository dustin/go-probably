package probably

import (
	"testing"
)

func TestNewSketch(t *testing.T) {
	s := NewSketch(100, 8)

	if s.String() != "{Sketch 100x8}" {
		t.Fatalf("Didn't String() properly: %v", s)
	}
}

func TestCounting(t *testing.T) {
	s := NewSketch(8, 3)

	hello := "hello"
	there := "there"
	world := "world"

	s.Increment(hello)
	s.Increment(hello)
	s.Increment(there)

	exp := []struct {
		s string
		v uint64
	}{
		{hello, 2},
		{there, 1},
		{world, 0},
	}

	for _, e := range exp {

		if s.Count(e.s) != e.v {
			t.Fatalf("Expected %v for %v, got %v", e.v, e.s, s.Count(e.s))
		}
	}
}

func TestMerging(t *testing.T) {
	s := NewSketch(8, 3)

	hello := "hello"
	there := "there"
	world := "world"

	s1 := NewSketch(8, 3)
	s2 := NewSketch(8, 3)

	s1.Increment(hello)
	s2.Increment(hello)
	s1.Increment(there)

	s.Merge(s1)
	s.Merge(s2)

	exp := []struct {
		s string
		v uint64
	}{
		{hello, 2},
		{there, 1},
		{world, 0},
	}

	for _, e := range exp {
		if s.Count(e.s) != e.v {
			t.Fatalf("Expected %v for %v, got %v", e.v, e.s, s.Count(e.s))
		}
	}
}

func BenchmarkHashNStringDepth64(b *testing.B) {
	s := "this is a test string to hash"

	for i := 0; i < b.N; i++ {
		hashn(s, 64, 64)
	}
}
