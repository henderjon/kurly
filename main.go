package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/mitchellh/ioprogress"
	"github.com/urfave/cli"
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

var outputFile string
var url string

func main() {
	app := cli.NewApp()
	app.Name = "working-title"
	app.Usage = "[options] <URL>"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "output, o",
			Usage:       "output filename",
			Destination: &outputFile,
		},
		cli.StringFlag{
			Name:        "url",
			Usage:       "URL",
			Destination: &url,
		},
	}

	app.Run(os.Args)

	if outputFile != "" {
		var err error
		Output, err = os.Create(outputFile)
		if err != nil {
			Status.Fatalln(err)
		}
		defer Output.Close()
	}

	testForStdoutRedirect()

	resp, err := http.Get(url)
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
