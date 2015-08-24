package dmc

import (
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Hongdian struct {
	Username string
	Password string
}

func (h *Hongdian) Name() string {
	return "Hongdian Radio"
}

func (h *Hongdian) Groups() []ModelType {
	return []ModelType{CellularModel}
}

func (h *Hongdian) Group(g ModelType) bool {
	switch g {
	case CellularModel:
		return true
	default:
		return false
	}
}

func (h *Hongdian) MatchString(s string) bool {
	return regexp.MustCompile("^Hongdian").MatchString(s)
}

func (h *Hongdian) Identify(orig string, ip net.IP, timeout time.Duration, retries int) (*State, error) {

	s := make(State)

	// default model
	s["model"] = "Hongdian Cellular Modem"

	cli := &http.Client{}
	pages := []string{"status_main.cgi", "lan_setup.cgi"}
	for _, p := range pages {
		r, err := h.discover(cli, "http://"+ip.String()+"/"+p)
		if err != nil {
			return nil, err
		}
		for key, value := range r {
			switch key {
			case "hostname":
				s["name"] = value
			case "pattern":
				s["version"] = value
			case "soft":
				s["software"] = value
			case "hard":
				s["firmware"] = value
			case "num":
				s["serial"] = value
			}
		}
	}

	return &s, nil
}

func (h *Hongdian) discover(cli *http.Client, url string) (map[string]string, error) {
	username := env(h.Username, "HONGDIAN_USERNAME", "admin")
	password := env(h.Password, "HONGDIAN_PASSWORD", "admin")

	request, err := http.NewRequest("GET", url, nil)
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

	results := make(map[string]string)

	doc.Find("div.setting input").Each(func(i int, s *goquery.Selection) {
		k, _ := s.Attr("name")
		v, _ := s.Attr("value")
		if k == "" || v == "" {
			return
		}
		results[k] = v

	})

	doc.Find("div.setting").Each(func(i int, s *goquery.Selection) {
		p := strings.Fields(strings.Replace(s.Text(), ")(", ") ", -1))
		if !(len(p) > 0) {
			return
		}

		switch p[0] {
		case "Capture(syspro.pattern)":
			results["pattern"] = p[1]
		case "Capture(syspro.num)":
			results["num"] = p[1]
		case "Capture(syspro.hard_release)":
			results["hard"] = p[1]
		case "Capture(syspro.soft_release)":
			results["soft"] = p[1]
		case "Capture(syslan.macaddr)":
			results["macaddr"] = p[1]
		case "Capture(share.ip_show)":
			results["ip"] = p[1]
		case "Capture(wireless.signal)":
			results["sig"] = p[1]
		}

	})

	return results, nil
}
