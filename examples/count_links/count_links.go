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

const (
	numWorkers = 8
	countMinW  = 1000000
	countMinD  = 8
	maxRecords = 100
	logError   = 0.0001
)

func maybeFatal(err error) {
	if err != nil {
		log.Fatalf("Error:  %v", err)
	}
}

func streamWorker(chin <-chan string,
	chout chan<- *probably.StreamTop,
	chcount chan<- *probably.HyperLogLog) {
	st := probably.NewStreamTop(countMinW, countMinD, maxRecords)
	hll := probably.NewHyperLogLog(logError)

	hash := dgohash.NewSuperFastHash()

	for b := range chin {
		links := strings.Split(b, " ")[1:]
		if len(links) > 1 {
			for i, l1 := range links {
				for _, l2 := range links[i+1:] {
					pair := l1 + " " + l2
					st.Add(pair)

					hash.Reset()
					hash.Write([]byte(pair))
					hll.Add(hash.Sum32())
				}
			}
		}
	}

	chout <- st
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
	outch := make(chan *probably.StreamTop)
	outchcount := make(chan *probably.HyperLogLog)

	for i := 0; i < numWorkers; i++ {
		go streamWorker(bch, outch, outchcount)
	}

	readFile(os.Args[1], bch)

	close(bch)

	st := probably.NewStreamTop(countMinW, countMinD, maxRecords)
	hll := probably.NewHyperLogLog(logError)

	for i := 0; i < numWorkers; i++ {
		st.Merge(<-outch)
		hll.Merge(<-outchcount)
	}

	log.Printf("Cardinality estimate:  %s",
		humanize.Comma(int64(hll.Count())))

	for _, p := range st.GetTop()[:20] {
		log.Printf("  %v\t->\t%s",
			p.Key, humanize.Comma(int64(p.Count)))
	}
}
