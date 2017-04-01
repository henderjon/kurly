package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	Status = log.New(os.Stderr, "", 0)
	Output = os.Stdout
)

func main() {
	out := flag.String("out", "", "output file")

	flag.Parse()

	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	// get the first non flag argument
	target := flag.Arg(0)
	if _, err := url.Parse(target); err != nil || !strings.HasPrefix(target, "http") || flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	}

	if flagset["out"] {
		var err error
		Output, err = os.Create(*out)
		if err != nil {
			flag.Usage()
			Status.Fatalln(err)
		}
		defer Output.Close()
	}

	go spinner(100 * time.Millisecond)

	resp, err := http.Get(target)
	if err != nil {
		Status.Fatal(err)
	}
	defer resp.Body.Close()

	n, err := io.Copy(Output, resp.Body)
	if err != nil {
		Status.Fatal(err)
	}

	Status.Printf("\nwrote %d bytes\n", n)
}

func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			Status.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}
