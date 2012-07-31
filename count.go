package probably

import (
	"fmt"
	"math"
)

// A count-min sketcher.
type Sketch [][]uint32

// Create a new count-min sketch with the given width and depth.
func NewSketch(w, d int) *Sketch {
	if d < 1 || w < 1 {
		panic("Dimensions must be positive")
	}

	rv := make(Sketch, d)
	for i := 0; i < d; i++ {
		rv[i] = make([]uint32, w)
	}

	return &rv
}

func (s Sketch) String() string {
	return fmt.Sprintf("{Sketch %dx%d}", len(s[0]), len(s))
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
func (s *Sketch) Increment(h string) (val uint32) {
	w := len((*s)[0])
	d := len(*s)
	val = math.MaxUint32
	for i, pos := range hashn(h, d, w) {
		v := (*s)[i][pos] + 1
		(*s)[i][pos] = v
		if v < val {
			val = v
		}
	}
	return val
}

// Get the estimate count for the given input.
func (s Sketch) Count(h string) uint32 {
	var min uint32 = math.MaxUint32
	w := len(s[0])
	d := len(s)
	for i, pos := range hashn(h, d, w) {
		v := s[i][pos]
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
