# Curly Fries, AKA curly

Curly Fries is an alternative to the widely popular curl program.

curly is designed to operate in a similar manner to curl, with select features.
Notably, curly is not aiming for feature parity, but common flags and mechanisms
particularly within the HTTP(S) realm are to be expected.

The current authors are not security experts, but want to contribute to the fledging
movement of replacing key tools and services with equivalents based on modern
and safe languages.  We recognize that people are fallible (including
ourselves), and for this reason believe we need all the help we can get.

Several languages exist which could be used to fulfill our goal, but in this case
we picked Golang.

## Installation

From source, assuming you have a recent version of Go installed on your system,
you can simply:

`go get github.com/davidjpeacock/curly`

## Binary download

Binaries are provided for the following platforms, download here and rename
as appropriate.

* [Linux amd64](https://github.com/davidjpeacock/curly/releases/download/untagged-31c3ab9c5e97121cd366/curly-linux-amd64-v1.0.0.tar.gz)
* [Linux arm](https://github.com/davidjpeacock/curly/releases/download/untagged-31c3ab9c5e97121cd366/curly-linux-arm-v1.0.0.tar.gz)
* [Mac OS X amd64](https://github.com/davidjpeacock/curly/releases/download/untagged-31c3ab9c5e97121cd366/curly-osx-amd64-v1.0.0.tar.gz)
* [Windows amd64](https://github.com/davidjpeacock/curly/releases/download/untagged-31c3ab9c5e97121cd366/curly-windows-amd64-v1.0.0.zip)

## Usage

See `curly --help` for usage information.

## Examples

Verbose output, showing headers
```
$ curly -v https://httpbin.org/ip
> GET /ip HTTP/1.1
> User-Agent [Curly_Fries/1.0]
> Accept [*/*]
> Host [httpbin.org]
< HTTP/1.1 200 OK
< Date [Thu, 27 Apr 2017 22:46:57 GMT]
< Content-Type [application/json]
< Access-Control-Allow-Credentials [true]
< Content-Length [31]
< Via [1.1 vegur]
< Connection [keep-alive]
< Server [gunicorn/19.7.1]
< Access-Control-Allow-Origin [*]
[<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<] 31 B/31 B
{
      "origin": "8.41.12.77"
}
```

Download file, preserving remote filename, timestamp, and following redirects
```
$ curly -R -O -L http://cdimage.debian.org/debian-cd/current/amd64/iso-cd/debian-8.7.1-amd64-netinst.iso
[<<<<<<                                ] 41.2 MB/260 MB
```

Upload file
```
$ curly -T ~/Downloads/image.jpeg https://httpbin.org/put
```

Posting elements with -d
```
$ curly -d bingo=bongo https://httpbin.org/post
```

## Roadmap

Succinctly, we're planning to cover all curl-like features relevant to HTTP(S), and would
love you to help.

## Contributing

Bug reports, feature requests, and pull request are all welcome.

## License

Curly Fries AKA curly is Copyright (c) 2017 David J Peacock and Al S-M, and
is published under the Apache 2.0 license.
