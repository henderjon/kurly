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
	"strconv"
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
	var outputFile *os.File
	var outputFilename string
	var remoteName bool
	var verbose bool
	var silent bool
	var target string
	var maxTime uint
	var remoteTime bool

	app := cli.NewApp()
	app.Name = "curly"
	app.Usage = "[options] URL"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "output, o",
			Usage:       "Filename to name url content to",
			Destination: &outputFilename,
		},
		cli.BoolFlag{
			Name:        "remote-name, O",
			Usage:       "Save output to file named with file part of URL",
			Destination: &remoteName,
		},
		cli.BoolFlag{
			Name:        "v",
			Usage:       "Verbose output",
			Destination: &verbose,
		},
		cli.BoolFlag{
			Name:        "silent, s",
			Usage:       "Mute curly entirely, operation without any output",
			Destination: &silent,
		},
		cli.UintFlag{
			Name:        "max-time, m",
			Usage:       "Maximum time to wait for an operation to complete in seconds",
			Destination: &maxTime,
		},
		cli.BoolFlag{
			Name:        "R",
			Usage:       "Set the timestamp of the local file to that of the remote file, if available",
			Destination: &remoteTime,
		},
	}

	app.Action = func(c *cli.Context) error {
		var remote *url.URL
		var err error
		if c.NArg() == 0 {
			cli.ShowAppHelp(c)
			os.Exit(0)
		}

		if verbose {
			Status.SetOutput(os.Stderr)
		}

		if maxTime > 0 {
			go func() {
				<-time.After(time.Duration(maxTime) * time.Second)
				log.Fatalf("Error: Maximum operation time of %d seconds expired, aborting\n", maxTime)
			}()

		}

		target = c.Args().Get(0)
		if remote, err = url.Parse(target); err != nil {
			log.Fatalf("Error: %s does not parse correctly as a URL\n", target)
		}

		if remoteName {
			outputFilename = path.Base(target)
		}
		if outputFilename != "" {
			if outputFile, err = os.Create(outputFilename); err != nil {
				log.Fatalf("Error: Unable to create file '%s' for output\n", outputFilename)
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

		clHeader := resp.Header.Get("Content-Length")
		for k, v := range resp.Header {
			Status.Println("<", k, v)
		}
		size, err := strconv.ParseInt(clHeader, 10, 64)
		if err != nil {
			size = 0
		}

		if !silent {
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
			if _, err = io.Copy(outputFile, progressR); err != nil {
				Status.Fatalf("Error: Failed to copy URL content; %s\n", err)
			}
		}
		if silent {
			if _, err = io.Copy(outputFile, resp.Body); err != nil {
				Status.Fatalf("Error: Failed to copy URL content; %s\n", err)
			}
		}

		if outputFilename != "" {
			outputFile.Close()
		}

		if rTime := resp.Header.Get("Last-Modified"); remoteTime && rTime != "" {
			if t, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", rTime); err == nil {
				os.Chtimes(outputFilename, t, t)
			}
		}
		return nil
	}

	app.Run(os.Args)
}
