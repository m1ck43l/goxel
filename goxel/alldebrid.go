package goxel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
)

var errors = map[int]string{
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
}

const (
	api   = "https://api.alldebrid.com"
	agent = "goxel"
)

func (s *AllDebridURLPreprocessor) initialize() {
	s.Client = &http.Client{}
	req, err := s.Client.Get(api + "/user/login?agent=" + agent + "&username=" + s.Login + "&password=" + s.Password)
	if err != nil {
		fmt.Printf("[ERROR] Following error occured while connecting to AllDebrid service: %v\n", err.Error())
		return
	}
	defer req.Body.Close()

	b, err := ioutil.ReadAll(req.Body)

	var resp LoginResponse
	err = json.Unmarshal(b, &resp)
	if err != nil {
		fmt.Printf("[ERROR] Following error occured while connecting to AllDebrid service: %v\n", err.Error())
		return
	}

	if !resp.Success {
		fmt.Printf("[ERROR] Following error occured while connecting to AllDebrid service: %v\n", errors[resp.Error])
		return
	}

	if !resp.User.Premium {
		fmt.Printf("[ERROR] Non premium user are not supported, bypassing.\n")
		return
	}

	fmt.Printf("[INFO] Successfully logged as [%v]\n", resp.User.Username)

	s.Token = resp.Token
	s.UseMe = true

	req, err = s.Client.Get(api + "/hosts/regexp")
	if err != nil {
		fmt.Printf("[ERROR] Can't retrieve hosts listing: %v\n", err.Error())
		return
	}
	defer req.Body.Close()

	b, err = ioutil.ReadAll(req.Body)

	var respD DomainsResponse
	err = json.Unmarshal(b, &respD)
	if err != nil {
		fmt.Printf("[ERROR] Can't retrieve hosts listing: %v\n", err.Error())
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
		s.initialize()
	}

	if !s.UseMe {
		return urls
	}

	output := make([]string, 0, len(urls))
	for _, url := range urls {
		var found bool
		for _, v := range s.Domains {
			if v.Match([]byte(url)) {
				req, err := s.Client.Get(api + "/link/unlock?agent=" + agent + "&token=" + s.Token + "&link=" + url)
				if err != nil {
					fmt.Printf("[ERROR] An error occured while debriding [%v]: %v\n", url, err.Error())
					continue
				}
				defer req.Body.Close()

				b, err := ioutil.ReadAll(req.Body)

				var resp LinkResponse
				err = json.Unmarshal(b, &resp)
				if err != nil {
					fmt.Printf("[ERROR] An error occured while debriding [%v]: %v\n", url, err.Error())
					continue
				}

				if !resp.Success {
					fmt.Printf("[ERROR] Ignoring [%v] due to an error: %v\n", url, errors[resp.Error])
				} else {
					output = append(output, resp.Infos.Link)
				}

				found = true
				break
			}
		}

		if !found {
			fmt.Printf("[INFO] Ignore alldebrid for [%v] as no domain matches the URL\n", url)
			output = append(output, url)
		}
	}
	return output
}
