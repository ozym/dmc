package dmc

import (
	"fmt"

	//	"bufio"
	//"io/ioutil"
	"net"
	"net/http"
	//	"net/url"
	"regexp"
	"strings"
	"time"

	//	"bytes"

	"code.google.com/p/mlab-ns2/gae/ns/digest"
	"github.com/PuerkitoBio/goquery"
	//	"golang.org/x/net/html"
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

	/*
		conn, err := net.Dial("tcp", ip.String()+":9999")
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		buf := bufio.NewReader(conn)
		str, err := buf.ReadString('\n')
		if err != nil {
			return nil, err
		}
		fmt.Println(str)
	*/

	/*
		package httpDigestAuth
	*/

	//return strings.Contains(str, "dig1"), nil

	/*
		switch {
		case strings.HasSuffix(orig, "+"):
			if s, _ := q.discover(ip, "6330", timeout); s != nil {
				return s, nil
			}
			if s, _ := q.discover(ip, "5330", timeout); s != nil {
				return s, nil
			}
		default:
			if s, _ := q.discover(ip, "5330", timeout); s != nil {
				return s, nil
			}
			if s, _ := q.discover(ip, "6330", timeout); s != nil {
				return s, nil
			}
		}
	*/

	transport := digest.NewTransport("rock", "kmi")

	client, err := transport.Client()
	if err != nil {
		return nil, err
	}

	s := make(State)

	/*
		c, err := r.menuload(ip, client)
		fmt.Println(c)
		if c == nil || err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(c.Body)
		fmt.Println((string)(body))
		c.Body.Close()
	*/
	/*
		for a, b := range c {
			s[a] = b
		}
	*/

	d, err := r.homeload(ip, client)
	fmt.Println(d)
	if d == nil || err != nil {
		return nil, err
	}
	/*
		body, err := ioutil.ReadAll(d.Body)
		fmt.Println((string)(body))
		d.Body.Close()
	*/
	for a, b := range d {
		s[a] = b
	}

	client.Get("http://" + ip.String() + "/logoff")

	/*
		resp, err := client.Get("http://" + ip.String() + "/homeload")
		if resp == nil || err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromResponse(resp)
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
						s["filesystem"] = b[len(b)-1]
					case strings.Contains(y, "Update"):
						s["hardware"] = strings.Join(b[0:2], " ")
						s["update"] = b[len(b)-1]
					case strings.Contains(y, "Serial number: 2149"):
						s["serial"] = b[len(b)-1]
					}
				}
			}
		})
	*/

	return &s, nil
}

func (r *Rock) cleanup(s string) string {
	return regexp.MustCompile("<[a-z0-9#=\"\\s\\_\\-\\+/]+>").ReplaceAllString(s, " ")
}

//func (r *Rock) menuload(ip net.IP, client *http.Client) (map[string]interface{}, error) {
func (r *Rock) menuload(ip net.IP, client *http.Client) (*http.Response, error) {

	/*
		transport := digest.NewTransport("rock", "kmi")

		client, err := transport.Client()
		if err != nil {
			return nil, err
		}
	*/

	resp, err := client.Get("http://" + ip.String() + "/menuload")
	if resp == nil || err != nil {
		fmt.Println(err)
		return nil, err
	}

	/*
		doc, err := goquery.NewDocumentFromResponse(resp)
		if doc == nil || err != nil {
			fmt.Println(err)
			return nil, err
		}
	*/

	/*
		h, err := doc.Html()
		if h == "" || err != nil {
			return nil, err
		}

		s := make(map[string]interface{})

		switch {
		case strings.Contains(h, "rockhound.jpg"):
			s["model"] = "Kinemetrics Slate"
		case strings.Contains(h, "basalt.jpg"):
			s["model"] = "Kinemetrics Basalt"
		}

		doc.Find("h4").Each(func(i int, g *goquery.Selection) {
			h, _ := g.Html()
			y := strings.Fields(r.cleanup(h))
			fmt.Println(strings.Join(y, " "))
			if !(len(y) > 1) {
				return
			}
			switch y[0] {
			case "Station":
				s["code"] = y[1]
			}
		})

		return s, nil
	*/
	return resp, nil
}

func (r *Rock) homeload(ip net.IP, client *http.Client) (map[string]interface{}, error) {

	resp, err := client.Get("http://" + ip.String() + "/homeload")
	if resp == nil || err != nil {
		fmt.Println(err)
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if doc == nil || err != nil {
		fmt.Println(err)
		return nil, err
	}

	s := make(map[string]interface{})

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
					s["filesystem"] = b[len(b)-1]
				case strings.Contains(y, "Update"):
					s["hardware"] = strings.Join(b[0:2], " ")
					s["update"] = b[len(b)-1]
				case strings.Contains(y, "Serial number: 2149"):
					s["serial"] = b[len(b)-1]
				}
			}
		}
	})

	return s, nil
}

/*
func (r *Rock) Identify(timeout time.Duration) (bool, error) {
	conn, err := net.Dial("tcp", r.IP.String()+":9999")
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	buf := bufio.NewReader(conn)
	str, err := buf.ReadString('\n')
	if err != nil {
		return false, err
	}
	fmt.Println(str)
	return strings.Contains(str, "dig1"), nil
*/

/*
	cli := &http.Client{}

		fmt.Println("in!\n")

			request, err := http.NewRequest("GET", "http://"+r.IP.String()+"/homeload", nil)
			if err != nil {
				fmt.Println(err)
				return false, err
			}
			request.SetBasicAuth("rock", "kmi")

			fmt.Println("2")
			resp, err := cli.Do(request)
			if resp == nil || err != nil {
				fmt.Println(err)
				return false, err
			}
			defer resp.Body.Close()
			fmt.Println(resp)
			fmt.Println("out!\n")

			body, err := ioutil.ReadAll(resp.Body)
			fmt.Println((string)(body))

			return strings.Contains((string)(body), "Canterbury Seismic Instruments"), nil
*/

/*
	//9999
	return false, nil
}
*/

/*
func digestAuthParams(r *http.Response) map[string]string {
	s := strings.SplitN(r.Header.Get("Www-Authenticate"), " ", 2)
	if len(s) != 2 || s[0] != "Digest" {
		return nil
	}

	result := map[string]string{}
	for _, kv := range strings.Split(s[1], ",") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		result[strings.Trim(parts[0], "\" ")] = strings.Trim(parts[1], "\" ")
	}
	return result
}
*/
