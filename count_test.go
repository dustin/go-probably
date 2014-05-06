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

	if s.Increment(hello) != 1 {
		t.Fatalf("Expected increment to set to 1")
	}
	if s.Increment(hello) != 2 {
		t.Fatalf("Expected increment to set to 2")
	}
	s.Increment(there)

	exp := []struct {
		s string
		v uint32
	}{
		{hello, 2},
		{there, 1},
		{world, 0},
	}

	for _, e := range exp {
		if s.Count(e.s) != e.v {
			t.Errorf("Expected %v for %v, got %v", e.v, e.s, s.Count(e.s))
		}
	}

	if s.Del(hello, 1) != 1 {
		t.Fatalf("Expected increment to set to 2")
	}
	exp = []struct {
		s string
		v uint32
	}{
		{hello, 1},
		{there, 1},
		{world, 0},
	}

	for _, e := range exp {
		if s.Count(e.s) != e.v {
			t.Errorf("Expected %v for %v, got %v", e.v, e.s, s.Count(e.s))
		}
	}
}

func TestMeanCounting(t *testing.T) {
	s := NewSketch(8, 3)

	hello := "hello"
	there := "there"
	world := "world"

	if s.ConservativeIncrement(hello) != 1 {
		t.Fatalf("Expected increment to set to 1")
	}
	if s.ConservativeIncrement(hello) != 2 {
		t.Fatalf("Expected increment to set to 2")
	}
	s.ConservativeIncrement(there)

	exp := []struct {
		s string
		v uint32
	}{
		{hello, 1},
		{there, 0},
		{world, 0},
	}

	for _, e := range exp {
		if s.CountMeanMin(e.s) != e.v {
			t.Errorf("Expected %v for %v, got %v", e.v, e.s, s.CountMeanMin(e.s))
		}
	}

	if s.Del(hello, 1) != 1 {
		t.Fatalf("Expected increment to set to 2")
	}
	exp = []struct {
		s string
		v uint32
	}{
		{hello, 0},
		{there, 0},
		{world, 0},
	}

	for _, e := range exp {
		if s.CountMeanMin(e.s) != e.v {
			t.Errorf("Expected %v for %v, got %v", e.v, e.s, s.CountMeanMin(e.s))
		}
	}
}

func TestConservativeCounting(t *testing.T) {
	s := NewSketch(8, 3)

	hello := "hello"
	there := "there"
	world := "world"

	if s.ConservativeIncrement(hello) != 1 {
		t.Errorf("Expected increment to set to 1")
	}
	if s.ConservativeIncrement(hello) != 2 {
		t.Errorf("Expected increment to set to 2")
	}
	s.ConservativeIncrement(there)

	exp := []struct {
		s string
		v uint32
	}{
		{hello, 2},
		{there, 1},
		{world, 0},
	}

	for _, e := range exp {
		if s.Count(e.s) != e.v {
			t.Errorf("Expected %v for %v, got %v", e.v, e.s, s.Count(e.s))
		}
	}

	if s.Del(hello, 1) != 1 {
		t.Errorf("Expected increment to set to 2")
	}
	exp = []struct {
		s string
		v uint32
	}{
		{hello, 1},
		{there, 1},
		{world, 0},
	}

	for _, e := range exp {
		if s.Count(e.s) != e.v {
			t.Errorf("Expected %v for %v, got %v", e.v, e.s, s.Count(e.s))
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
		v uint32
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

func TestCompress(t *testing.T) {
	s := NewSketch(8, 3)

	hello := "hello"
	there := "there"
	world := "world"

	s.Increment(hello)
	s.Increment(hello)
	s.Increment(there)

	for _, l := range s.sk {
		t.Log(l)
	}

	s.Compress()

	for _, l := range s.sk {
		t.Log(l)
		if len(l) != 4 {
			t.Errorf("Expected length 4, got %v\n", len(l))
		}
	}

	exp := []struct {
		s string
		v uint32
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

	lim := uint32(64)

	for i := 0; i < b.N; i++ {
		h1, h2 := hashn(s) // , 64, 64)
		var h uint32
		for j := uint32(0); j < 64; j++ {
			h += (h1 + j*h2) % lim
		}
	}
}
