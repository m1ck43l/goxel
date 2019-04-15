# goxel

[![GoDoc](https://godoc.org/github.com/m1ck43l/goxel?status.svg)](https://godoc.org/github.com/m1ck43l/goxel) [![Build Status](https://travis-ci.org/m1ck43l/goxel.svg?branch=master)](https://travis-ci.org/m1ck43l/goxel/builds) [![Go Report Card](https://goreportcard.com/badge/github.com/m1ck43l/goxel)](https://goreportcard.com/report/github.com/m1ck43l/goxel)

*GoXel - download accelerator written in Go*

GoXel is a Go package for faster downloads from the internet:

* Monitor download progress
* Resume incomplete downloads
* Guess filename from URL path
* Download batches of files concurrently

Requires Go v1.8+

## Build

```
$ make
```

Make will create the goxel executable in the bin directory

## Usage

```
$ bin/goxel -h
goxel [options] [url1] [url2] [url...]
  -alldebrid-password string
        Alldebrid password, can also be passed in the GOXEL_ALLDEBRID_PASSWD environment variable
  -alldebrid-username string
        Alldebrid username, can also be passed in the GOXEL_ALLDEBRID_USERNAME environment variable
  -file string
        Links file
  -header value
        Extra header(s), key=value
  -insecure
        Bypass SSL validation
  -max-conn int
        Max number of connections (default 8)
  -max-conn-file int
        Max number of connections per file (default 4)
  -no-override
        Do not override existing file(s)
  -output string
        Output directory
  -proxy string
        Proxy string: (http|https|socks5)://0.0.0.0:0000
  -quiet
        No stdout output
  -version
        Version

Visit https://github.com/m1ck43l/goxel/issues to report bugs.
```

## Contributing

Pull requests for new features, bug fixes, and suggestions are welcome!

## License

[Apache 2](https://github.com/m1ck43l/goxel/blob/master/LICENSE)