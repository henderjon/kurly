package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/davidjpeacock/cli"
)

type Options struct {
	outputFilename string
	fileUpload     string
	remoteName     bool
	verbose        bool
	maxTime        uint
	remoteTime     bool
	followRedirect bool
	maxRedirects   uint
	redirectsTaken uint
	silent         bool
	method         string
	headers        []string
	agent          string
	expectTimeout  uint
	data           []string
	dataAscii      []string
	dataRaw        []string
	dataBinary     []string
	dataURLEncode  []string
}

func (o *Options) getOptions(app *cli.App) {
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "output, o",
			Usage:       "Filename to name url content to",
			Destination: &o.outputFilename,
		},
		cli.StringFlag{
			Name:        "upload-file, T",
			Usage:       "File to upload",
			Destination: &o.fileUpload,
		},
		cli.BoolFlag{
			Name:        "remote-name, O",
			Usage:       "Save output to file named with file part of URL",
			Destination: &o.remoteName,
		},
		cli.BoolFlag{
			Name:        "v",
			Usage:       "Verbose output",
			Destination: &o.verbose,
		},
		cli.UintFlag{
			Name:        "max-time, m",
			Usage:       "Maximum time to wait for an operation to complete in seconds",
			Destination: &o.maxTime,
		},
		cli.BoolFlag{
			Name:        "R",
			Usage:       "Set the timestamp of the local file to that of the remote file, if available",
			Destination: &o.remoteTime,
		},
		cli.BoolFlag{
			Name:        "location, L",
			Usage:       "Follow 3xx redirects",
			Destination: &o.followRedirect,
		},
		cli.UintFlag{
			Name:        "max-redirs",
			Usage:       "Maximum number of 3xx redirects to follow",
			Destination: &o.maxRedirects,
			Value:       10,
		},
		cli.BoolFlag{
			Name:        "silent, s",
			Usage:       "Mute curly entirely, operation without any output",
			Destination: &o.silent,
		},
		cli.StringFlag{
			Name:        "request, X",
			Usage:       "HTTP method to use",
			Destination: &o.method,
			Value:       "GET",
		},
		cli.StringFlag{
			Name:        "user-agent, A",
			Usage:       "User agent to set for this request",
			Destination: &o.agent,
			Value:       "Curly_Fries/1.0",
		},
		cli.StringSliceFlag{
			Name:  "header, H",
			Usage: "Extra headers to be sent with the request",
		},
		cli.UintFlag{
			Name:        "expect100-timeout",
			Usage:       "Timeout in seconds for Expect: 100-continue wait period",
			Destination: &o.expectTimeout,
			Value:       1,
		},
		cli.StringSliceFlag{
			Name:  "data, d",
			Usage: "Sends the specified data in a POST request to the server",
		},
		cli.StringSliceFlag{
			Name:  "data-ascii",
			Usage: "The same as --data, -d",
		},
		cli.StringSliceFlag{
			Name:  "data-raw",
			Usage: "Basically the same as --data-binary (no @ interpretation)",
		},
		cli.StringSliceFlag{
			Name:  "data-binary",
			Usage: "Sends the data as binary",
		},
		cli.StringSliceFlag{
			Name:  "data-urlencode",
			Usage: "Sends the data as urlencoded ascii",
		},
	}
}

func (o *Options) checkRedirect(req *http.Request, via []*http.Request) error {
	o.redirectsTaken++

	if !o.followRedirect || o.redirectsTaken >= o.maxRedirects {
		return http.ErrUseLastResponse
	}

	return nil
}

func (o *Options) ProcessData() {
	var uriEncodes url.Values
	for _, d := range o.dataAscii {
		parts := strings.SplitN(d, "=", 2)
		if len(parts) == 1 {
			o.data = append(o.data, d)
			continue
		}
		if strings.HasPrefix(parts[1], "@") {
			data, err := ioutil.ReadFile(strings.TrimPrefix(parts[1], "@"))
			if err != nil {
				log.Fatalf("Unable to read file %s for data element %s\n", strings.TrimPrefix(parts[1], "@"), parts[0])
			}
			data = []byte(strings.Replace(string(data), "\r", "", -1))
			data = []byte(strings.Replace(string(data), "\n", "", -1))
			o.data = append(o.data, fmt.Sprintf("%s=%s", parts[0], string(data)))
		} else {
			o.data = append(o.data, d)
		}
	}
	for _, d := range o.dataRaw {
		parts := strings.SplitN(d, "=", 2)
		o.data = append(o.data, fmt.Sprintf("%s=%s", parts[0], parts[1]))
	}
	for _, d := range o.dataBinary {
		parts := strings.SplitN(d, "=", 2)
		o.data = append(o.data, fmt.Sprintf("%s=%s", parts[0], parts[1]))
	}
	for _, d := range o.dataURLEncode {
		parts := strings.SplitN(d, "=", 2)
		uriEncodes.Add(parts[0], parts[1])
	}
	if len(uriEncodes) > 0 {
		o.data = append(o.data, uriEncodes.Encode())
	}
}
