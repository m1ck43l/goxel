# <img src="https://user-images.githubusercontent.com/918212/56085390-af18c780-5e42-11e9-9ae7-7ba453502ddb.png" width="300">


[![GoDoc](https://godoc.org/github.com/m1ck43l/goxel?status.svg)](https://godoc.org/github.com/m1ck43l/goxel) [![Build Status](https://travis-ci.org/m1ck43l/goxel.svg?branch=master)](https://travis-ci.org/m1ck43l/goxel/builds) [![Go Report Card](https://goreportcard.com/badge/github.com/m1ck43l/goxel)](https://goreportcard.com/report/github.com/m1ck43l/goxel) [![Coverage Status](https://coveralls.io/repos/github/m1ck43l/goxel/badge.svg?branch=master&_=0.11)](https://coveralls.io/github/m1ck43l/goxel?branch=master)


*GoXel - download accelerator written in Go*

GoXel is a Go package for faster downloads from the internet:

* Monitor download progress
* Resume incomplete downloads
* Guess filename from URL path
* Download batches of files concurrently

Requires Go v1.8+

GoXel was inspired by axel (https://github.com/axel-download-accelerator/axel)

## Build

```
$ make clean && make deps && make && make test
```

Make will create the goxel executable in the bin directory

## Usage

```
$ bin/goxel -h
GoXel is a download accelerator written in Go
Usage: goxel [options] [url1] [url2] [url...]
      --alldebrid-password string         Alldebrid password, can also be passed in the GOXEL_ALLDEBRID_PASSWD environment variable                                                                                 
      --alldebrid-username string         Alldebrid username, can also be passed in the GOXEL_ALLDEBRID_USERNAME environment variable                                                                               
      --buffer-size int                   Buffer size in KB (default 256)
  -f, --file string                       File containing links to download (1 per line)
      --header header-name=header-value   Extra header(s) (default [])
  -h, --help                              This information
      --insecure                          Bypass SSL validation
      --max-conn int                      Max number of connections (default 8)
  -m, --max-conn-file int                 Max number of connections per file (default 4)
      --no-resume                         Don't resume downloads
  -o, --output string                     Output directory
      --overwrite                         Overwrite existing file(s)
  -p, --proxy string                      Proxy string: (http|https|socks5)://0.0.0.0:0000
  -q, --quiet                             No stdout output
  -s, --scroll                            Scroll output instead of in place display
      --version                           Version

Visit https://github.com/m1ck43l/goxel/issues to report bugs.
```

## Benchmark

This benchmark compares Axel and GoXel for multiple downloads using files from https://www.thinkbroadband.com/download.
All links were done using a broadhand connection: 455.0 Mbit/s download, 276.4 Mbit/s upload, lantency 3ms over WiFi.

![Benchmark](https://user-images.githubusercontent.com/918212/56504862-2e308e80-651a-11e9-96de-398bf263b060.png)


## Contributing

Pull requests for new features, bug fixes, and suggestions are welcome!

## License

[Apache 2](https://github.com/m1ck43l/goxel/blob/master/LICENSE)
