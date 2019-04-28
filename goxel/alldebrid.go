package goxel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

var aderrors = map[int]string{
	1:  "Invalid token",
	2:  "Invalid user or password",
	3:  "Geolock protection active, please login from the website",
	4:  "User is banned",
	5:  "Please provide both username and password for authentification, or a valid token",
	30: "This link is not supported.",
	31: "This link is not available on the file hoster website.",
	32: "Host under maintenance or not available.",
	33: "You have reached the free trial limit (7 days // 25GB downloaded or host uneligible for free trial).",
	34: "Too many concurrent downloads.",
	35: "All servers are full for this host, please retry later.",
	36: "You have reached the download limit for this host.",
	37: "You must be premium to process this link.",
	38: "Link is password protected.",
	39: "Generic unlocking error.",
}

// LoginResponse represents the Alldebrid login JSON response, it contains the authentication Token
// that will be used in all calls
type LoginResponse struct {
	Success bool          `json:"success"`
	Token   string        `json:"token"`
	User    AllDebridUser `json:"user"`
	Error   int           `json:"errorCode"`
}

// AllDebridUser represents a Alldebrid User as returned by the login call
type AllDebridUser struct {
	Premium  bool   `json:"isPremium"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// DomainsResponse represents the list of domain's regexp supported by Alldebrid
type DomainsResponse struct {
	Success bool              `json:"success"`
	Domains map[string]string `json:"hosts"`
}

// LinkResponse represents the answer of the link's debriding request
type LinkResponse struct {
	Success bool      `json:"success"`
	Error   int       `json:"errorCode"`
	Infos   LinkInfos `json:"infos"`
}

// LinkInfos contains link related information as sent by Alldebrid
type LinkInfos struct {
	Link     string `json:"link"`
	Filename string `json:"filename"`
}

// AllDebridURLPreprocessor implements the UrlPreprocessor interface.
// It handles the conversion of links after the debriding
type AllDebridURLPreprocessor struct {
	Client                 *http.Client
	Login, Password, Token string
	Initialized, UseMe     bool
	Domains                map[string]*regexp.Regexp
	API                    string
}

const (
	api   = "https://api.alldebrid.com"
	agent = "goxel"
)

func (s *AllDebridURLPreprocessor) initialize(url string) {
	if url != "" {
		s.API = url
	} else {
		s.API = api
	}

	s.Client, _ = NewClient()
	req, err := s.Client.Get(s.API + "/user/login?agent=" + agent + "&username=" + s.Login + "&password=" + s.Password)
	if err != nil {
		cMessages <- NewErrorMessage("ALLDEBRID", fmt.Sprintf("Following error occurred while connecting to AllDebrid service: %v", err.Error()))
		return
	}
	defer req.Body.Close()

	b, _ := ioutil.ReadAll(req.Body)

	var resp LoginResponse
	err = json.Unmarshal(b, &resp)
	if err != nil {
		cMessages <- NewErrorMessage("ALLDEBRID", fmt.Sprintf("Following error occurred while connecting to AllDebrid service: %v", err.Error()))
		return
	}

	if !resp.Success {
		cMessages <- NewErrorMessage("ALLDEBRID", fmt.Sprintf("Following error occurred while connecting to AllDebrid service: %v", aderrors[resp.Error]))
		return
	}

	if !resp.User.Premium {
		cMessages <- NewWarningMessage("ALLDEBRID", "Non premium user are not supported, bypassing.")
		return
	}

	cMessages <- NewInfoMessage("ALLDEBRID", fmt.Sprintf("Successfully logged as [%v]", resp.User.Username))

	s.Token = resp.Token
	s.UseMe = true

	req, err = s.Client.Get(s.API + "/hosts/regexp")
	if err != nil {
		cMessages <- NewErrorMessage("ALLDEBRID", fmt.Sprintf("Can't retrieve hosts listing: %v", err.Error()))
		return
	}
	defer req.Body.Close()

	b, _ = ioutil.ReadAll(req.Body)

	var respD DomainsResponse
	err = json.Unmarshal(b, &respD)
	if err != nil {
		cMessages <- NewErrorMessage("ALLDEBRID", fmt.Sprintf("Can't retrieve hosts listing: %v", err.Error()))
		return
	}

	s.Domains = make(map[string]*regexp.Regexp, len(respD.Domains))
	for k, v := range respD.Domains {
		s.Domains[k] = regexp.MustCompile(v)
	}

	s.Initialized = true
}

func (s *AllDebridURLPreprocessor) process(urls []string) []string {
	if !s.Initialized {
		s.initialize("")
	}

	if !s.UseMe {
		return urls
	}

	output := make([]string, 0, len(urls))
	for _, url := range urls {
		var found bool
		for _, v := range s.Domains {
			if v.Match([]byte(url)) {
				req, err := s.Client.Get(s.API + "/link/unlock?agent=" + agent + "&token=" + s.Token + "&link=" + url)
				if err != nil {
					cMessages <- NewErrorMessage("ALLDEBRID", fmt.Sprintf("An error occurred while debriding [%v]: %v", url, err.Error()))
					continue
				}
				defer req.Body.Close()

				b, _ := ioutil.ReadAll(req.Body)

				var resp LinkResponse
				err = json.Unmarshal(b, &resp)
				if err != nil {
					cMessages <- NewErrorMessage("ALLDEBRID", fmt.Sprintf("An error occurred while debriding [%v]: %v", url, err.Error()))
					continue
				}

				if !resp.Success {
					cMessages <- NewErrorMessage("ALLDEBRID", fmt.Sprintf("Ignoring [%v] due to an error: %v", url, aderrors[resp.Error]))
				} else {
					output = append(output, resp.Infos.Link)
				}

				found = true
				break
			}
		}

		if !found {
			cMessages <- NewWarningMessage("ALLDEBRID", fmt.Sprintf("Ignore alldebrid for [%v] as no domain matches the URL", url))
			output = append(output, url)
		}
	}
	return output
}
