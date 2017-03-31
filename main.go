package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	url := flag.String("url", "", "url")
	out := flag.String("out", "", "output file")

	flag.Parse()

	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if flagset["url"] && flagset["out"] {
		go spinner(100 * time.Millisecond)

		f, err := os.Create(*out)
		defer f.Close()

		resp, err := http.Get(*url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		n, err := io.Copy(f, resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\nwrote %d bytes\n", n)
	} else {
		flag.Usage()
	}
}

func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}
