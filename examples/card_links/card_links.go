package main

import (
	"bufio"
	"compress/bzip2"
	"io"
	"log"
	"os"
	"strings"

	"github.com/dgryski/dgohash"

	"github.com/dustin/go-humanize"
	"github.com/dustin/go-probably"
)

// 2012/07/30 17:20:03 Initializing a slice of 134217728 for 0.0001 from m=10400, k=27, k_comp=5
// 2012/07/30 17:20:43 Initializing a slice of 0 for 1e-05 from m=104000, k=34, k_comp=-2

const (
	numWorkers = 8
	logError   = 0.0001
)

func maybeFatal(err error) {
	if err != nil {
		log.Fatalf("Error:  %v", err)
	}
}

func streamWorker(chin <-chan string,
	chcount chan<- *probably.HyperLogLog) {
	hll := probably.NewHyperLogLog(logError)

	hash := dgohash.NewSuperFastHash()

	for b := range chin {
		links := strings.Split(b, " ")[1:]
		if len(links) > 1 {
			for i, l1 := range links {
				for _, l2 := range links[i+1:] {
					pair := l1 + " " + l2

					hash.Reset()
					hash.Write([]byte(pair))
					hll.Add(hash.Sum32())
				}
			}
		}
	}

	chcount <- hll
}

func readFile(fn string, ch chan<- string) {
	f, err := os.Open(fn)
	maybeFatal(err)
	defer f.Close()

	br := bufio.NewReader(bzip2.NewReader(f))

	for i := 0; ; i++ {
		b, err := br.ReadBytes('\n')
		switch err {
		case io.EOF:
			log.Printf("Processed %s lines total",
				humanize.Comma(int64(i)))
			return
		case nil:
			ch <- string(b[:len(b)-1])
		default:
			log.Fatalf("Error reading input: %v", err)
		}

		if i%100000 == 0 {
			log.Printf("Processed %s lines",
				humanize.Comma(int64(i)))
		}
	}
}

func main() {
	bch := make(chan string, 1024)
	outchcount := make(chan *probably.HyperLogLog)

	for i := 0; i < numWorkers; i++ {
		go streamWorker(bch, outchcount)
	}

	readFile(os.Args[1], bch)

	close(bch)

	// st := probably.NewStreamTop(countMinW, countMinD, maxRecords)
	// hll := probably.NewHyperLogLog(logError)

	// Grab the first worker's results and merge the rest of them in.
	hll := <-outchcount

	for i := 0; i < numWorkers-1; i++ {
		hll.Merge(<-outchcount)
	}

	log.Printf("Cardinality estimate:  %s",
		humanize.Comma(int64(hll.Count())))
}
