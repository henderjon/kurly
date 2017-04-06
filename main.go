package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/mitchellh/ioprogress"
	"github.com/urfave/cli"
)

func main() {
	var outputFile string
	var target string

	app := cli.NewApp()
	app.Name = "working-title"
	app.Usage = "[options] URL"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "output, o",
			Usage:       "output filename",
			Destination: &outputFile,
		},
	}

	app.Action = func(c *cli.Context) error {
		if c.NArg() > 0 {
			target = c.Args().Get(0)
		}
		if outputFile != "" {
			f, err := os.Create(outputFile)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			resp, err := http.Get(target)
			if err != nil {
				panic(err)
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

			n, err := io.Copy(f, progressR)
			if err != nil {
				panic(err)
			}
			fmt.Printf("\nwrote %d bytes\n", n)
		}
		return nil
	}

	app.Run(os.Args)
}
