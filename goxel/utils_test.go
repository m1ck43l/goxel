package goxel

import (
	"net/http"
	"testing"
)

func TestHTTP(t *testing.T) {
	goxel.Proxy = "http://127.0.0.1:8123"
	client, err := NewClient()

	if err != nil {
		t.Error("Error while creating http proxy", err)
	}

	tr := client.Transport.(*http.Transport)
	if tr.Proxy == nil {
		t.Error("Error while creating http proxy, proxy is nil")
	}
}

func TestHttpError(t *testing.T) {
	goxel.Proxy = "http://" + string([]byte{0x7f, 0x7f}) + ":1234"
	_, err := NewClient()

	if err == nil {
		t.Error("Error should be thrown")
	}
}

func TestHTTPS(t *testing.T) {
	goxel.Proxy = "https://127.0.0.1:8123"
	client, err := NewClient()

	if err != nil {
		t.Error("Error while creating https proxy", err)
	}

	tr := client.Transport.(*http.Transport)
	if tr.Proxy == nil {
		t.Error("Error while creating https proxy, proxy is nil")
	}
}

func TestHttpsError(t *testing.T) {
	goxel.Proxy = "https://" + string([]byte{0x7f, 0x7f}) + ":1234"
	_, err := NewClient()

	if err == nil {
		t.Error("Error should be thrown")
	}
}

func TestSocks5(t *testing.T) {
	goxel.Proxy = "socks5://127.0.0.1:8123"
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
	goxel.Proxy = "ftp://127.0.0.1:8123"
	_, err := NewClient()

	if err == nil {
		t.Error("Error, shoud fail")
	}
}
