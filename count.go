package probably

import (
	"fmt"
	"hash/crc32"
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
	h1 := crc32.Update(0, crc32.IEEETable, []byte(s))

	rv := make([]int, 0, d)

	for i := 0; i < d; i++ {
		h := int(crc32.Update(h1, crc32.IEEETable, []byte{byte(i)})) % lim
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
	for i, pos := range hashn(h, d, w) {
		val = (*s)[pos][i]
		(*s)[pos][i]++
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
