package dmc

import (
	//"fmt"

	"crypto/tls"
	//	"io/ioutil"
	"net"
	"net/http"
	//	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Cusp struct {
	Username string
	Password string
}

func (c *Cusp) Name() string {
	return "CSI Cusp"
}

func (c *Cusp) MatchString(s string) bool {
	return regexp.MustCompile("^CSI Cusp").MatchString(s)
}

func (c *Cusp) Groups() []ModelType {
	return []ModelType{StrongModel}
}

func (c *Cusp) Group(g ModelType) bool {
	switch g {
	case StrongModel:
		return true
	default:
		return false
	}
}

func (c *Cusp) Identify(hostname string, ip net.IP, timeout time.Duration, retries int) (*State, error) {

	username := env(c.Username, "CUSP_USERNAME", "default")
	password := env(c.Password, "CUSP_PASSWORD", "default")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cli := &http.Client{Transport: tr}

	request, err := http.NewRequest("GET", "https://"+ip.String()+"/admin/status.cgi", nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(username, password)

	resp, err := cli.Do(request)
	if resp == nil || err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if doc == nil || err != nil {
		return nil, err
	}

	s := State{Values: make(map[string]interface{})}
	s.Values["hostname"] = hostname
	s.Values["ipaddress"] = ip.String()

	doc.Find("div p").Each(func(i int, g *goquery.Selection) {
		if strings.Contains(g.Text(), "Administrator Access") {
			if p := strings.Fields(strings.Replace(g.Text(), "'", "", -1)); len(p) > 0 {
				s.Values["site"] = p[0]
			}
		}
	})

	doc.Find("div.box td").Each(func(i int, g *goquery.Selection) {
		if g.Next() != nil {
			switch g.Text() {
			case "Sensor serial number":
				if p := strings.Fields(g.Next().Text()); len(p) > 0 {
					s.Values["serial"] = p[0]
				}
			case "Data acquisition firmware revision":
				if p := strings.Fields(g.Next().Text()); len(p) > 1 {
					s.Values["hardware"] = strings.Join(p[0:len(p)-1], " ")
					s.Values["software"] = p[len(p)-1]
				}
			case "Sensor firmware revision":
				if p := strings.Fields(g.Next().Text()); len(p) > 0 {
					s.Values["firmware"] = p[0]
				}
			}

		}
	})

	if _, ok := s.Values["hardware"]; !ok {
		return nil, nil
	}

	if p := strings.Fields(s.Values["hardware"].(string)); len(p) > 0 {
		s.Values["model"] = "CSI Cusp " + strings.Replace(p[len(p)-1], "+", "", -1)
	}

	return &s, nil
}

func (c *Cusp) Status(hostname string, ip net.IP, timeout time.Duration, retries int) (*State, error) {

	username := env(c.Username, "CUSP_USERNAME", "default")
	password := env(c.Password, "CUSP_PASSWORD", "default")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cli := &http.Client{Transport: tr}

	request, err := http.NewRequest("GET", "https://"+ip.String()+"/admin/status.cgi", nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(username, password)

	resp, err := cli.Do(request)
	if resp == nil || err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if doc == nil || err != nil {
		return nil, err
	}

	s := State{Values: make(map[string]interface{})}
	s.Values["hostname"] = hostname
	s.Values["ipaddress"] = ip.String()

	doc.Find("div p").Each(func(i int, g *goquery.Selection) {
		if strings.Contains(g.Text(), "Administrator Access") {
			if p := strings.Fields(strings.Replace(g.Text(), "'", "", -1)); len(p) > 0 {
				s.Values["site"] = p[0]
			}
		}
	})

	doc.Find("div.box td").Each(func(i int, g *goquery.Selection) {
		if g.Next() != nil {

			if p := strings.Fields(g.Next().Text()); len(p) > 1 {
				switch g.Text() {
				case "Data acquisition firmware revision":
					s.Values["hardware"] = strings.Join(p[0:len(p)-1], " ")
					s.Values["software"] = p[len(p)-1]
				}
			}

			if p := strings.Fields(g.Next().Text()); len(p) > 0 {
				switch g.Text() {
				case "Sensor serial number":
					s.Values["serial"] = p[0]
				case "Sensor firmware revision":
					s.Values["firmware"] = p[0]
				case "Current system voltage":
					if v, err := strconv.ParseFloat(p[0], 64); err == nil {
						s.Values["voltage"] = v
					}
				case "Current system temperature":
					if v, err := strconv.ParseFloat(p[0], 64); err == nil {
						s.Values["temperature"] = v
					}
				case "GPS loss period":
					s.Values["quality"] = float64(10.0)
					if strings.Contains(p[0], "Never") {
						s.Values["loss"] = int(24 * 3600)
					} else if v, err := time.ParseDuration(p[0]); err == nil {
						if q := 100.0 - 90*v/(24*time.Hour); q > 10.0 {
							s.Values["quality"] = q
						}
						s.Values["loss"] = int(v / time.Second)
					}
				case "Timing system primary source":
					s.Values["timing"] = strings.Join(p, " ")
					switch {
					case strings.Contains(strings.Join(p, " "), "GPS"):
						s.Values["lock"] = "gps"
					case strings.Contains(strings.Join(p, " "), "NTP"):
						s.Values["lock"] = "ntp"
					default:
						s.Values["lock"] = "none"
					}
				case "GPS state":
					s.Values["gps"] = strings.Join(p, " ")
				case "Disk space free":
					if v, err := strconv.ParseInt(p[0], 10, 32); err == nil {
						s.Values["disk"] = v
					}
				case "Current X channel noise (over 1 min)":
					if v, err := strconv.ParseFloat(p[0], 64); err == nil {
						s.Values["x"] = v
					}
				case "Current Y channel noise (over 1 min)":
					if v, err := strconv.ParseFloat(p[0], 64); err == nil {
						s.Values["y"] = v
					}
				case "Current Z channel noise (over 1 min)":
					if v, err := strconv.ParseFloat(p[0], 64); err == nil {
						s.Values["z"] = v
					}
				}
			}
		}
	})

	if _, ok := s.Values["hardware"]; !ok {
		return nil, nil
	}

	if p := strings.Fields(s.Values["hardware"].(string)); len(p) > 0 {
		s.Values["model"] = "CSI Cusp " + strings.Replace(p[len(p)-1], "+", "", -1)
	}

	if _, ok := s.Values["loss"]; !ok {
		s.Values["loss"] = 0
	}
	if _, ok := s.Values["quality"]; !ok {
		s.Values["quality"] = float64(100)
	}

	s.Values["timestamp"] = time.Now().Unix()

	return &s, nil
}
