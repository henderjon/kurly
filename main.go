package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/alsm/ioprogress"
	"github.com/davidjpeacock/cli"
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
	var body io.Reader

	app := cli.NewApp()
	app.Name = "curly"
	app.Usage = "[options] URL"
	opts.getOptions(app)

	app.Action = func(c *cli.Context) error {
		var remote *url.URL
		var err error

		client.CheckRedirect = opts.checkRedirect
		opts.headers = c.StringSlice("header")
		opts.dataAscii = c.StringSlice("data")
		opts.dataAscii = append(opts.dataAscii, c.StringSlice("data-ascii")...)
		opts.dataBinary = c.StringSlice("data-binary")
		opts.dataRaw = c.StringSlice("data-raw")
		opts.dataURLEncode = c.StringSlice("data-urlencode")

		opts.ProcessData()

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
		if remote.Scheme == "" {
			remote.Scheme = "http"
			remote, _ = url.Parse(remote.String())
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

		if opts.fileUpload != "" {
			opts.method = "PUT"

			tr := &http.Transport{
				ExpectContinueTimeout: time.Duration(opts.expectTimeout) * time.Second,
			}
			client.Transport = tr
			opts.headers = append(opts.headers, "Expect: 100-continue")

			reader, err := os.Open(opts.fileUpload)
			if err != nil {
				log.Fatalf("Error opening %s\n", opts.fileUpload)
			}

			if !opts.silent {
				fi, err := reader.Stat()
				if err != nil {
					log.Fatalf("Unable to get file stats for %v\n", opts.fileUpload)
				}
				body = &ioprogress.Reader{
					Reader: reader,
					Size:   fi.Size(),
					DrawFunc: ioprogress.DrawTerminalf(os.Stderr, func(progress, total int64) string {
						return fmt.Sprintf(
							"%s %s",
							(ioprogress.DrawTextFormatBarWithIndicator(40, '>'))(progress, total),
							ioprogress.DrawTextFormatBytes(progress, total))
					}),
				}
			} else {
				body = reader
			}
		}

		if len(opts.data) > 0 {
			var data bytes.Buffer
			opts.method = "POST"
			opts.headers = append(opts.headers, "Content-Type: application/x-www-form-urlencoded")

			for i, d := range opts.data {
				data.WriteString(d)
				if i < len(opts.data)-1 {
					data.WriteRune('&')
				}
			}
			body = &data
		}

		req, err := http.NewRequest(opts.method, target, body)
		if err != nil {
			log.Fatalf("Error: unable to create http %s request; %s\n", opts.method, err)
		}
		req.Header.Set("User-Agent", opts.agent)
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Host", remote.Host)
		if body != nil {
			switch b := body.(type) {
			case *os.File:
				fi, err := b.Stat()
				if err != nil {
					log.Fatalf("Unable to get file stats for %v\n", opts.fileUpload)
				}
				req.ContentLength = fi.Size()
				req.Header.Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
			case *ioprogress.Reader:
				req.ContentLength = b.Size
				req.Header.Set("Content-Length", strconv.FormatInt(b.Size, 10))
			case *bytes.Buffer:
				req.Header.Set("Content-Length", strconv.FormatInt(int64(b.Len()), 10))
			}
		}
		setHeaders(req, opts.headers)

		Status.Println(">", req.Method, req.URL.Path, req.Proto)
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
						(ioprogress.DrawTextFormatBarWithIndicator(40, '<'))(progress, total),
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

func setHeaders(r *http.Request, h []string) {
	for _, header := range h {
		hParts := strings.Split(header, ": ")
		switch len(hParts) {
		case 0:
			//surely not
		case 1:
			//must be an empty Header or a delete
			switch {
			case strings.HasSuffix(header, ";"):
				r.Header.Set(strings.TrimSuffix(header, ";"), "")
			case strings.HasSuffix(header, ":"):
				r.Header.Del(strings.TrimSuffix(header, ":"))
			default:
			}
		case 2:
			//standard expected
			r.Header.Set(hParts[0], hParts[1])
		default:
			//more than expected, use first element as Header name
			//and rejoin the rest as header content
			r.Header.Set(hParts[0], strings.Join(hParts[1:], ": "))
		}
	}
}
