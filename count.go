package probably

import (
	"fmt"
	"math"
)

// A count-min sketcher.
type Sketch [][]uint64

// Create a new count-min sketch with the given width and depth.
func NewSketch(w, d int) *Sketch {
	if d < 1 || w < 1 {
		panic("Dimensions must be positive")
	}

	rv := make(Sketch, w)
	for i := 0; i < w; i++ {
		rv[i] = make([]uint64, d)
	}

	return &rv
}

func (s Sketch) String() string {
	return fmt.Sprintf("{Sketch %dx%d}", len(s), len(s[0]))
}

func hashn(s string, d, lim int) []int {
	hval := uint32(0)
	for _, c := range s {
		hval = (hval << 3) ^ hval ^ uint32(c)
	}

	rv := make([]int, 0, d)

	for i := 0; i < d; i++ {
		h2 := hval + uint32(i)
		h2 += (h2 << 10)
		h2 ^= (h2 >> 6)

		h := int(h2) % lim
		if h < 0 {
			h = 0 - h
		}
		rv = append(rv, h)
	}

	return rv
}

// Increment the count for the given input.
func (s *Sketch) Increment(h string) (val uint64) {
	d := len((*s)[0])
	w := len(*s)
	val = math.MaxUint64
	for i, pos := range hashn(h, d, w) {
		v := (*s)[pos][i] + 1
		(*s)[pos][i] = v
		if v < val {
			val = v
		}
	}
	return val
}

// Get the estimate count for the given input.
func (s Sketch) Count(h string) uint64 {
	var min uint64 = math.MaxUint64
	d := len(s[0])
	w := len(s)
	for i, pos := range hashn(h, d, w) {
		v := s[pos][i]
		if v < min {
			min = v
		}
	}
	return min
}

// Merge the given sketch into this one.
func (s *Sketch) Merge(from *Sketch) {
	if len(*s) != len(*from) || len((*s)[0]) != len((*from)[0]) {
		panic("Can't merge different sketches with different dimensions")
	}

	for i, l := range *from {
		for j, v := range l {
			(*s)[i][j] += v
		}
	}
}
