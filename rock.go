package dmc

import (
	"net"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"time"

	"code.google.com/p/mlab-ns2/gae/ns/digest"
	"github.com/PuerkitoBio/goquery"
)

type Rock struct {
}

func (r *Rock) Name() string {
	return "Rock"
}

func (r *Rock) MatchString(s string) bool {
	return regexp.MustCompile("^Kinemetrics").MatchString(s)
}

func (r *Rock) Groups() []ModelType {
	return []ModelType{StrongModel}
}

func (r *Rock) Group(g ModelType) bool {
	switch g {
	case StrongModel:
		return true
	default:
		return false
	}
}

func (r *Rock) Identify(orig string, ip net.IP, timeout time.Duration, retries int) (*State, error) {

	transport := digest.NewTransport("rock", "kmi")
	cookieJar, _ := cookiejar.New(nil)

	client := &http.Client{
		Transport: transport,
		Jar:       cookieJar,
	}

	s := State{Values: make(map[string]interface{})}

	resp, err := client.Get("http://" + ip.String() + "/")
	if resp == nil || err != nil {
		return nil, err
	}
	defer client.Get("http://" + ip.String() + "/logoff")

	resp, err = client.Get("http://" + ip.String() + "/menuload")
	if resp == nil || err != nil {
		// there's a bug in golang http which means it may work the next time ...
		resp, err = client.Get("http://" + ip.String() + "/menuload")
		if resp == nil || err != nil {
			return nil, err
		}
	}
	defer client.Get("http://" + ip.String() + "/logoff")

	doc, err := goquery.NewDocumentFromResponse(resp)
	if doc == nil || err != nil {
		return nil, err
	}

	raw, err := doc.Html()
	if raw == "" || err != nil {
		return nil, err
	}

	switch {
	case strings.Contains(raw, "rockhound.jpg"):
		s.Values["model"] = "Kinemetrics Slate"
	case strings.Contains(raw, "basalt.jpg"):
		s.Values["model"] = "Kinemetrics Basalt"
	default:
		s.Values["model"] = "Kinemetrics Rock"
	}

	doc.Find("h4").Each(func(i int, g *goquery.Selection) {
		h, _ := g.Html()
		y := strings.Fields(regexp.MustCompile("<[a-z0-9#=\"\\s\\_\\-\\+/]+>").ReplaceAllString(h, " "))
		if !(len(y) > 1) {
			return
		}
		switch y[0] {
		case "Station":
			s.Values["code"] = y[1]
		}
	})

	resp, err = client.Get("http://" + ip.String() + "/homeload")
	if resp == nil || err != nil {
		// there's a bug in golang http which means it may work the next time ...
		resp, err = client.Get("http://" + ip.String() + "/homeload")
		if resp == nil || err != nil {
			return nil, err
		}
	}
	defer client.Get("http://" + ip.String() + "/logoff")

	doc, err = goquery.NewDocumentFromResponse(resp)
	if doc == nil || err != nil {
		return nil, err
	}

	doc.Find("p").Each(func(i int, g *goquery.Selection) {
		h, err := g.Html()
		if h == "" || err != nil {
			return
		}
		x := strings.Split(h, "<br/>")
		for _, y := range x {
			b := strings.Fields(y)
			if len(b) > 1 {
				switch {
				case strings.Contains(y, "filesystem"):
					s.Values["filesystem"] = b[len(b)-1]
				case strings.Contains(y, "Update"):
					s.Values["hardware"] = strings.Join(b[0:2], " ")
					s.Values["update"] = b[len(b)-1]
				case strings.Contains(y, "Serial number"):
					s.Values["serial"] = b[len(b)-1]
				case strings.Contains(y, "Software version"):
					s.Values["software"] = b[len(b)-1]
				}
			}
		}
	})

	return &s, nil
}
