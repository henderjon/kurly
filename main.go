package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/alsm/ioprogress"
	"github.com/urfave/cli"
)

var (
	client = http.Client{}
	Status = log.New(ioutil.Discard, "", 0)
)

func init() {
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version, V",
		Usage: "print the version",
	}
}

func main() {
	var target string
	var outputFile *os.File
	var opts Options

	app := cli.NewApp()
	app.Name = "curly"
	app.Usage = "[options] URL"
	opts.getOptions(app)

	app.Action = func(c *cli.Context) error {
		var remote *url.URL
		var err error

		client.CheckRedirect = opts.checkRedirect

		if c.NArg() == 0 {
			cli.ShowAppHelp(c)
			os.Exit(0)
		}

		if opts.verbose {
			Status.SetOutput(os.Stderr)
		}

		if opts.maxTime > 0 {
			go func() {
				<-time.After(time.Duration(opts.maxTime) * time.Second)
				log.Fatalf("Error: Maximum operation time of %d seconds expired, aborting\n", opts.maxTime)
			}()

		}

		target = c.Args().Get(0)
		if remote, err = url.Parse(target); err != nil {
			log.Fatalf("Error: %s does not parse correctly as a URL\n", target)
		}

		if opts.remoteName {
			opts.outputFilename = path.Base(target)
		}
		if opts.outputFilename != "" {
			if outputFile, err = os.Create(opts.outputFilename); err != nil {
				log.Fatalf("Error: Unable to create file '%s' for output\n", opts.outputFilename)
			}
		} else {
			outputFile = os.Stdout
		}

		req, err := http.NewRequest("GET", target, nil)
		if err != nil {
			log.Fatalf("Error: unable to create http request; %s\n", err)
		}
		req.Header.Set("User-Agent", "Curly_Fries/1.0")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Host", remote.Host)

		for k, v := range req.Header {
			Status.Println(">", k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error: Unable to get URL; %s\n", err)
		}
		defer resp.Body.Close()

		Status.Printf("< %s %s", resp.Proto, resp.Status)

		for k, v := range resp.Header {
			Status.Println("<", k, v)
		}

		if !opts.silent {
			progressR := &ioprogress.Reader{
				Reader: resp.Body,
				Size:   resp.ContentLength,
				DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
					return fmt.Sprintf(
						"%s %s",
						(ioprogress.DrawTextFormatBar(40))(progress, total),
						ioprogress.DrawTextFormatBytes(progress, total))
				}),
			}
			if _, err = io.Copy(outputFile, progressR); err != nil {
				Status.Fatalf("Error: Failed to copy URL content; %s\n", err)
			}
		}
		if opts.silent {
			if _, err = io.Copy(outputFile, resp.Body); err != nil {
				Status.Fatalf("Error: Failed to copy URL content; %s\n", err)
			}
		}

		if opts.outputFilename != "" {
			outputFile.Close()
		}

		if rTime := resp.Header.Get("Last-Modified"); opts.remoteTime && rTime != "" {
			if t, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", rTime); err == nil {
				os.Chtimes(opts.outputFilename, t, t)
			}
		}
		return nil
	}

	app.Run(os.Args)
}
