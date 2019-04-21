package goxel

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/net/proxy"
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getWidth() uint {
	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return uint(100)
	}
	return uint(ws.Col)
}

func fmtDuration(d uint64) string {
	h := d / 3600
	m := (d - h*3600) / 60
	s := d - m*60 - h*3600

	if h > 99 {
		return fmt.Sprintf(" > 99 h ")
	}
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// headerFlag is used to parse headers on the CLI
// It allows multiple elements to be passed
type headerFlag []string

func (h *headerFlag) String() string {
	return fmt.Sprintf("%v", *h)
}

func (h *headerFlag) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func (h *headerFlag) Type() string {
	return "header-name=header-value"
}

// counter allows for an atomic counter
type counter struct {
	v   int
	mux sync.Mutex
}

func (c *counter) inc() {
	c.mux.Lock()
	c.v++
	c.mux.Unlock()
}

func (c *counter) dec() {
	c.mux.Lock()
	c.v--
	c.mux.Unlock()
}

// NewClient returns a HTTP client with the requested configuration
// It supports HTTP and SOCKS proxies
func NewClient() (*http.Client, error) {
	client := &http.Client{}

	if goxel.Proxy != "" {
		re := regexp.MustCompile(`^(http|https|socks5)://`)
		protocol := re.Find([]byte(goxel.Proxy))

		if protocol != nil {
			var transport *http.Transport

			if string(protocol) == "http://" || string(protocol) == "https://" {
				pURL, err := url.Parse(goxel.Proxy)
				if err != nil {
					return client, errors.New("Invalid proxy URL")
				}

				transport = &http.Transport{
					Proxy: http.ProxyURL(pURL),
				}
			} else if string(protocol) == "socks5://" {
				dialer, _ := proxy.SOCKS5("tcp", strings.Replace(goxel.Proxy, "socks5://", "", 1), nil, proxy.Direct)
				transport = &http.Transport{
					Dial: dialer.Dial,
				}
			} else {
				return client, errors.New("Invalid proxy protocol")
			}

			if transport != nil {
				client = &http.Client{
					Transport: transport,
				}
			}
		} else {
			return client, errors.New("Invalid proxy URL")
		}

	}

	return client, nil
}
