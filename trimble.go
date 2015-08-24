package dmc

import (
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Trimble struct {
	Username string
	Password string
}

func (t *Trimble) Name() string {
	return "Trimble"
}

func (t *Trimble) MatchString(s string) bool {
	return regexp.MustCompile("^Trimble").MatchString(s)
}

func (t *Trimble) Groups() []ModelType {
	return []ModelType{GNSSModel}
}

func (t *Trimble) Group(g ModelType) bool {
	switch g {
	case GNSSModel:
		return true
	default:
		return false
	}
}

func (t *Trimble) Identify(orig string, ip net.IP, timeout time.Duration, retries int) (*State, error) {
	switch {
	case strings.Contains(orig, "NetRS"):
		if s, err := t.IdentifyNetRS(ip, timeout, retries); s != nil && err == nil {
			return s, nil
		}
		if s, err := t.IdentifyNetR9(ip, timeout, retries); s != nil && err == nil {
			return s, nil
		}
	case strings.Contains(orig, "NetR9"):
		if s, err := t.IdentifyNetR9(ip, timeout, retries); s != nil && err == nil {
			return s, nil
		}
		if s, err := t.IdentifyNetRS(ip, timeout, retries); s != nil && err == nil {
			return s, nil
		}
	}
	return nil, nil
}

func (t *Trimble) IdentifyNetRS(ip net.IP, timeout time.Duration, retries int) (*State, error) {
	username := env(t.Username, "NETRS_USERNAME", "sysadmin")
	password := env(t.Password, "NETRS_PASSWORD", "")

	return t.identify(username, password, ip, timeout, retries)
}

func (t *Trimble) IdentifyNetR9(ip net.IP, timeout time.Duration, retries int) (*State, error) {
	username := env(t.Username, "NETR9_USERNAME", "admin")
	password := env(t.Password, "NETR9_PASSWORD", "")

	return t.identify(username, password, ip, timeout, retries)
}

func (t *Trimble) identify(username, password string, ip net.IP, timeout time.Duration, retries int) (*State, error) {

	cli := &http.Client{}

	s := make(State)

	for _, n := range []string{"serialNumber", "FirmwareVersion", "RefStation"} {
		r, err := t.show(cli, username, password, ip, n)
		if err != nil {
			return nil, err
		}
		if f := strings.Fields(r); len(f) > 1 {
			switch {
			case strings.Contains(r, "SerialNumber"):
				s["serial"] = strings.Replace(f[len(f)-1], "sn=", "", -1)
			case strings.Contains(r, "FirmwareVersion"):
				s["firmware"] = strings.Replace(f[1], "version=", "", -1)
			case strings.Contains(r, "RefStation"):
				for _, a := range f {
					if b := strings.Split(a, "="); len(b) > 1 {
						switch b[0] {
						case "lat":
							if v, err := strconv.ParseFloat(b[1], 64); err == nil {
								s["latitude"] = v
							}
						case "lon":
							if v, err := strconv.ParseFloat(b[1], 64); err == nil {
								s["longitude"] = v
							}
						case "height":
							if v, err := strconv.ParseFloat(b[1], 64); err == nil {
								s["height"] = v
							}
						case "Code":
							s["site"] = strings.Replace(b[1], "'", "", -1)
							s["model"] = "Trimble NetR9"
						case "CmrStationName":
							s["site"] = strings.Replace(b[1], "'", "", -1)
							s["model"] = "Trimble NetRS"
						}
					}

				}

			}
		}
	}

	// had to have found a model
	if _, ok := s["model"]; !ok {
		return nil, nil
	}

	return &s, nil
}

func (t *Trimble) show(cli *http.Client, username, password string, ip net.IP, value string) (string, error) {

	request, err := http.NewRequest("GET", "http://"+ip.String()+"/prog/show?"+value, nil)
	if err != nil {
		return "", err
	}
	request.SetBasicAuth(username, password)

	resp, err := cli.Do(request)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return (string)(body), nil
}
