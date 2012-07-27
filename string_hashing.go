package probably

import (
	"hash/crc32"
)

// Adaptor for hashing strings.
type HashableString string

func (h HashableString) Hash(d int) int {
	h1 := crc32.Update(0, crc32.IEEETable, []byte(h))
	return int(crc32.Update(h1, crc32.IEEETable, []byte{byte(d)}))
}
