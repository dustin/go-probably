package main

import (
	"bufio"
	"compress/bzip2"
	"io"
	"log"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/dustin/go-probably"
)

const (
	numWorkers = 4
	countMinW  = 10000
	countMinD  = 4
	maxRecords = 100
)

func maybeFatal(err error) {
	if err != nil {
		log.Fatalf("Error:  %v", err)
	}
}

func streamWorker(chin <-chan string, chout chan<- *probably.StreamTop) {
	st := probably.NewStreamTop(countMinW, countMinD, maxRecords)

	for b := range chin {
		links := strings.Split(b, " ")[1:]
		for _, l := range links {
			st.Add(l)
		}
	}

	chout <- st
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

	for i := 0; i < numWorkers; i++ {
		go streamWorker(bch, outch)
	}

	readFile(os.Args[1], bch)

	close(bch)

	st := probably.NewStreamTop(countMinW, countMinD, maxRecords)

	for i := 0; i < numWorkers; i++ {
		st.Merge(<-outch)
	}

	for _, p := range st.GetTop()[:20] {
		log.Printf("  %v\t->\t%v", p.Key, p.Count)
	}
}