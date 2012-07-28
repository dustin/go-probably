package main

import (
	"bufio"
	"compress/bzip2"
	"io"
	"log"
	"os"
	"strings"

	"github.com/dustin/go-probably"
)

const (
	numWorkers = 8
	countMinW  = 100000
	countMinD  = 8
	maxRecords = 1000
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

func main() {
	f, err := os.Open(os.Args[1])
	maybeFatal(err)
	defer f.Close()

	z := bzip2.NewReader(f)

	br := bufio.NewReader(z)

	bch := make(chan string, 1024)
	outch := make(chan *probably.StreamTop)

	for i := 0; i < numWorkers; i++ {
		go streamWorker(bch, outch)
	}

	for i := 0; ; i++ {
		b, err := br.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		maybeFatal(err)

		bch <- string(b[:len(b)-1])

		if i%100000 == 0 {
			log.Printf("Processed %v lines", i)
		}
	}

	close(bch)

	st := probably.NewStreamTop(countMinW, countMinD, maxRecords)

	for i := 0; i < numWorkers; i++ {
		st.Merge(<-outch)
	}

	for _, p := range st.GetTop() {
		log.Printf("  %v -> %v", p.Key, p.Count)
	}
}
