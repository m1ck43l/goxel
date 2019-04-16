package goxel

import (
	"net/http"
	"testing"
)

func TestHTTP(t *testing.T) {
	proxyURL = "http://127.0.0.1:8123"
	client, err := NewClient()

	if err != nil {
		t.Error("Error while creating http proxy", err)
	}

	tr := client.Transport.(*http.Transport)
	if tr.Proxy == nil {
		t.Error("Error while creating http proxy, proxy is nil")
	}
}

func TestHTTPS(t *testing.T) {
	proxyURL = "https://127.0.0.1:8123"
	client, err := NewClient()

	if err != nil {
		t.Error("Error while creating https proxy", err)
	}

	tr := client.Transport.(*http.Transport)
	if tr.Proxy == nil {
		t.Error("Error while creating https proxy, proxy is nil")
	}
}

func TestSocks5(t *testing.T) {
	proxyURL = "socks5://127.0.0.1:8123"
	client, err := NewClient()

	if err != nil {
		t.Error("Error while creating socks proxy", err)
	}

	tr := client.Transport.(*http.Transport)
	if tr.Dial == nil {
		t.Error("Error while creating socks proxy, Dial is nil")
	}
}

func TestBadProtocol(t *testing.T) {
	proxyURL = "ftp://127.0.0.1:8123"
	_, err := NewClient()

	if err == nil {
		t.Error("Error, shoud fail")
	}
}
