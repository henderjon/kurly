package main

import (
	"net/http"

	"github.com/urfave/cli"
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
	}
}

func (o *Options) checkRedirect(req *http.Request, via []*http.Request) error {
	o.redirectsTaken++

	if !o.followRedirect || o.redirectsTaken >= o.maxRedirects {
		return http.ErrUseLastResponse
	}

	return nil
}
