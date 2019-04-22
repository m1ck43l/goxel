package goxel

import (
	"fmt"
	"net/http"
	"testing"
)

func SetupAlldebridTest() {
	http.HandleFunc("/user/login", func(w http.ResponseWriter, r *http.Request) {
		gets := r.URL.Query()["username"]
		username := gets[0]

		switch username {
		case "test1":
			fmt.Fprintf(w, "{bad=json}")
		case "test2":
			fmt.Fprintf(w, "{\"success\":false, \"errorCode\": 2}")
		case "test3":
			fmt.Fprintf(w, "{\"success\":true, \"token\": \"alldebridtoken\", \"user\": {\"isPremium\":false, \"username\": \"alldebrid\", \"email\": \"alldebrid@mail.com\"}}")
		case "test4":
			fmt.Fprintf(w, "{\"success\":true, \"token\": \"alldebridtoken\", \"user\": {\"isPremium\":true, \"username\": \"alldebrid\", \"email\": \"alldebrid@mail.com\"}}")
		}
	})

	http.HandleFunc("/hosts/regexp", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "{\"success\":true, \"hosts\": {\"test\": \".*test.*\"}}")
	})

	http.HandleFunc("/link/unlock", func(w http.ResponseWriter, r *http.Request) {
		gets := r.URL.Query()["token"]
		token := gets[0]

		if token != "alldebridtoken" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		getsu := r.URL.Query()["link"]
		url := getsu[0]

		if url == "http://upload.com/test/video.mp4" {
			fmt.Fprintf(w, "{\"success\":false, \"errorCode\": 30}")
		} else {
			fmt.Fprintf(w, "{\"success\":true, \"infos\": {\"link\": \"http://test.com/ok.mp4\", \"filename\": \"test\"}}")
		}
	})
}

func TestServerError(t *testing.T) {
	alldebrid := AllDebridURLPreprocessor{}
	alldebrid.initialize("http://127.0.0.1:8080")

	if alldebrid.Initialized || alldebrid.UseMe {
		t.Error("Alldebrid should not be usable")
	}
}

func TestBadJsonResponse(t *testing.T) {
	alldebrid := AllDebridURLPreprocessor{
		Login: "test1",
	}
	alldebrid.initialize("http://127.0.0.1:8080")

	if alldebrid.Initialized || alldebrid.UseMe {
		t.Error("Alldebrid should not be usable")
	}
}

func TestErrorLogin(t *testing.T) {
	alldebrid := AllDebridURLPreprocessor{
		Login: "test2",
	}
	alldebrid.initialize("http://127.0.0.1:8080")

	if alldebrid.Initialized || alldebrid.UseMe {
		t.Error("Alldebrid should not be usable")
	}
}

func TestNotPremium(t *testing.T) {
	alldebrid := AllDebridURLPreprocessor{
		Login: "test3",
	}
	alldebrid.initialize("http://127.0.0.1:8080")

	if alldebrid.Initialized || alldebrid.UseMe {
		t.Error("Alldebrid should not be usable")
	}
}

func TestLoginOkAndPremium(t *testing.T) {
	alldebrid := AllDebridURLPreprocessor{
		Login: "test4",
	}
	alldebrid.initialize("http://127.0.0.1:8080")

	if !alldebrid.UseMe || alldebrid.Token != "alldebridtoken" {
		t.Error("Alldebrid should be usable")
	}
}

func TestHosts(t *testing.T) {
	alldebrid := AllDebridURLPreprocessor{
		Login: "test4",
	}
	alldebrid.initialize("http://127.0.0.1:8080")

	if !alldebrid.UseMe || alldebrid.Token != "alldebridtoken" || !alldebrid.Initialized {
		t.Error("Alldebrid should be usable and initialized")
	}
}

func TestNoUrlMatching(t *testing.T) {
	alldebrid := AllDebridURLPreprocessor{
		Login: "test4",
	}
	alldebrid.initialize("http://127.0.0.1:8080")

	if !alldebrid.UseMe || alldebrid.Token != "alldebridtoken" || !alldebrid.Initialized {
		t.Error("Alldebrid should be usable and initialized")
	}

	urls := alldebrid.process([]string{"http://upload.com/video.mp4"})
	if len(urls) != 1 || urls[0] != "http://upload.com/video.mp4" {
		t.Error("Url should have stayed unchanged")
	}
}

func TestUrlMatchingButInvalidJson(t *testing.T) {
	alldebrid := AllDebridURLPreprocessor{
		Login: "test4",
	}
	alldebrid.initialize("http://127.0.0.1:8080")

	if !alldebrid.UseMe || alldebrid.Token != "alldebridtoken" || !alldebrid.Initialized {
		t.Error("Alldebrid should be usable and initialized")
	}
	alldebrid.Token = "badtoken"

	urls := alldebrid.process([]string{"http://upload.com/test/video.mp4"})
	if len(urls) != 1 || urls[0] != "http://upload.com/test/video.mp4" {
		t.Error("Url should have stayed unchanged")
	}
}

func TestUnlinkError(t *testing.T) {
	alldebrid := AllDebridURLPreprocessor{
		Login: "test4",
	}
	alldebrid.initialize("http://127.0.0.1:8080")

	if !alldebrid.UseMe || alldebrid.Token != "alldebridtoken" || !alldebrid.Initialized {
		t.Error("Alldebrid should be usable and initialized")
	}

	urls := alldebrid.process([]string{"http://upload.com/test/video.mp4"})
	if len(urls) > 0 {
		t.Error("Urls should be empty")
	}
}

func TestUnlinkSuccess(t *testing.T) {
	alldebrid := AllDebridURLPreprocessor{
		Login: "test4",
	}
	alldebrid.initialize("http://127.0.0.1:8080")

	if !alldebrid.UseMe || alldebrid.Token != "alldebridtoken" || !alldebrid.Initialized {
		t.Error("Alldebrid should be usable and initialized")
	}

	urls := alldebrid.process([]string{"http://upload.com/test/video-ok.mp4"})
	if len(urls) != 1 || urls[0] != "http://test.com/ok.mp4" {
		t.Error("Url should be debrided")
	}
}
