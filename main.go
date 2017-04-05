package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/mitchellh/ioprogress"
)

var (
	Status         = log.New(os.Stderr, "", 0)
	Output         = os.Stdout
	stdoutRedirect bool
)

func testForStdoutRedirect() {
	info, err := os.Stdin.Stat()
	if err != nil {
		return
	}
	if (info.Mode() & os.ModeCharDevice) != os.ModeCharDevice {
		stdoutRedirect = true
	}
}

func main() {
	o := flag.String("o", "", "output file")

	flag.Parse()

	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	// get the first non flag argument
	target := flag.Arg(0)
	if _, err := url.Parse(target); err != nil || !strings.HasPrefix(target, "http") || flag.NArg() > 1 {
		flag.Usage()
		os.Exit(1)
	}

	if flagset["o"] {
		var err error
		Output, err = os.Create(*o)
		if err != nil {
			flag.Usage()
			Status.Fatalln(err)
		}
		defer Output.Close()
	}

	testForStdoutRedirect()

	resp, err := http.Get(target)
	if err != nil {
		Status.Fatal(err)
	}
	defer resp.Body.Close()

	clHeader := resp.Header.Get("Content-Length")
	size, err := strconv.ParseInt(clHeader, 10, 64)
	if err != nil {
		size = 0
	}

	progressR := &ioprogress.Reader{
		Reader: resp.Body,
		Size:   size,
		DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
			return fmt.Sprintf(
				"%s %s",
				(ioprogress.DrawTextFormatBar(40))(progress, total),
				ioprogress.DrawTextFormatBytes(progress, total))
		}),
	}

	n, err := io.Copy(Output, progressR)
	if err != nil {
		Status.Fatal(err)
	}

	Status.Printf("\nwrote %d bytes\n", n)
}
