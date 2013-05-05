package probably

import (
	"fmt"
	"hash/fnv"
	"math"
	"sort"
)

// A count-min sketcher.
type Sketch struct {
	sk        [][]uint32
	rowCounts []uint32
}

// Create a new count-min sketch with the given width and depth.
func NewSketch(w, d int) *Sketch {
	if d < 1 || w < 1 {
		panic("Dimensions must be positive")
	}

	s := &Sketch{}

	s.sk = make([][]uint32, d)
	for i := 0; i < d; i++ {
		s.sk[i] = make([]uint32, w)
	}

	s.rowCounts = make([]uint32, d)

	return s
}

func (s Sketch) String() string {
	return fmt.Sprintf("{Sketch %dx%d}", len(s.sk[0]), len(s.sk))
}

func hashn(s string, d, lim int) []int {

	fnv1a := fnv.New32a()
	fnv1a.Write([]byte(s))
	h1 := fnv1a.Sum32()

	// inlined jenkins one-at-a-time hash
	h2 := uint32(0)
	for _, c := range s {
		h2 += uint32(c)
		h2 += h2 << 10
		h2 ^= h2 >> 6
	}
	h2 += (h2 << 3)
	h2 ^= (h2 >> 11)
	h2 += (h2 << 15)

	rv := make([]int, 0, d)

	for i := 0; i < d; i++ {
		h := int(h1) + i*int(h2)
		h %= lim
		if h < 0 {
			h = 0 - h
		}
		rv = append(rv, h)
	}

	return rv
}

// Add 'count' occurences of the given input
func (s *Sketch) Add(h string, count uint32) (val uint32) {
	w := len(s.sk[0])
	d := len(s.sk)
	val = math.MaxUint32
	for i, pos := range hashn(h, d, w) {
		s.rowCounts[i] += count
		v := s.sk[i][pos] + count
		s.sk[i][pos] = v
		if v < val {
			val = v
		}
	}
	return val
}

// Delete 'count' occurences of the given input
func (s *Sketch) Del(h string, count uint32) (val uint32) {
	w := len(s.sk[0])
	d := len(s.sk)
	val = math.MaxUint32
	for i, pos := range hashn(h, d, w) {
		s.rowCounts[i] -= count
		v := s.sk[i][pos] - count
		if v > s.sk[i][pos] { // did we wrap-around?
			v = 0
		}
		s.sk[i][pos] = v
		if v < val {
			val = v
		}
	}
	return val
}

// Increment the count for the given input.
func (s *Sketch) Increment(h string) (val uint32) {
	return s.Add(h, 1)
}

// Increment the count (conservatively) for the given input.
func (s *Sketch) ConservativeIncrement(h string) (val uint32) {
	return s.ConservativeAdd(h, 1)
}

// Add the count (conservatively) for the given input.
func (s *Sketch) ConservativeAdd(h string, count uint32) (val uint32) {
	w := len(s.sk[0])
	d := len(s.sk)
	hashes := hashn(h, d, w)

	val = math.MaxUint32
	for i, pos := range hashes {
		v := s.sk[i][pos]
		if v < val {
			val = v
		}
	}

	val += count

	// Conservative update means no counter is increased to more than the
	// size of the smallest counter plus the size of the increment.  This technique
	// first described in Cristian Estan and George Varghese. 2002. New directions in
	// traffic measurement and accounting. SIGCOMM Comput. Commun. Rev., 32(4).

	for i, pos := range hashes {
		v := s.sk[i][pos]
		if v < val {
			s.rowCounts[i] += (val - s.sk[i][pos])
			s.sk[i][pos] = val
		}
	}
	return val
}

// Get the estimate count for the given input.
func (s Sketch) Count(h string) uint32 {
	var min uint32 = math.MaxUint32
	w := len(s.sk[0])
	d := len(s.sk)
	for i, pos := range hashn(h, d, w) {
		v := s.sk[i][pos]
		if v < min {
			min = v
		}
	}
	return min
}

/*
   CountMinMean described in:
   Fan Deng and Davood Raﬁei. 2007. New estimation algorithms for streaming data: Count-min can do more.
   http://webdocs.cs.ualberta.ca/~fandeng/paper/cmm.pdf

   Sketch Algorithms for Estimating Point Queries in NLP
   Amit Goyal, Hal Daume III and Graham Cormode
   EMNLP-CONLL 2012
   http://www.umiacs.umd.edu/~amit/Papers/goyalPointQueryEMNLP12.pdf
*/

// Get the estimate count for the given input, using the count-min-mean
// heuristic.  This gives more accurate results than Count() for low-frequency
// counts at the cost of larger under-estimation error.  For tasks sensitive to
// under-estimation, use the regular Count() and only call ConservativeAdd()
// and ConservativeIncrement() when constructing your sketch.
func (s Sketch) CountMinMean(h string) uint32 {
	var min uint32 = math.MaxUint32
	w := len(s.sk[0])
	d := len(s.sk)
	residues := make([]float64, d)
	for i, pos := range hashn(h, d, w) {
		v := s.sk[i][pos]
		noise := float64(s.rowCounts[i]-s.sk[i][pos]) / float64(w-1)
		residues[i] = float64(v) - noise
		// negative count doesn't make sense
		if residues[i] < 0 {
			residues[i] = 0
		}
		if v < min {
			min = v
		}
	}

	sort.Float64s(residues)
	var median uint32
	if d%2 == 1 {
		median = uint32(residues[(d+1)/2])
	} else {
		// integer average without overflow
		x := uint32(residues[d/2])
		y := uint32(residues[d/2+1])
		median = (x & y) + (x^y)/2
	}

	// count estimate over the upper-bound (min) doesn't make sense
	if min < median {
		return min
	}

	return median
}

// Merge the given sketch into this one.
func (s *Sketch) Merge(from *Sketch) {
	if len(s.sk) != len(from.sk) || len(s.sk[0]) != len(from.sk[0]) {
		panic("Can't merge different sketches with different dimensions")
	}

	for i, l := range from.sk {
		for j, v := range l {
			s.sk[i][j] += v
		}
	}
}
