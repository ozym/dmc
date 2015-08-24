package dmc

import (
	"crypto/tls"
	"net"
	"net/http"
	"regexp"
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

func (c *Cusp) Identify(orig string, ip net.IP, timeout time.Duration, retries int) (*State, error) {

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

	s := make(State)

	/*
	   <div class='boxhdr'>Firmware and sensor parameters:</div><div class='box'><table bgcolor='#FFFFFF' border='0'>
	   <tr><td>&nbsp;</td></tr>
	   <tr><td width='250'>Data acquisition firmware revision</td><td width='200'><font color='#0000FF'>&nbsp;&nbsp;Cusp 3D+ r400.63</font></td></tr>
	   <tr><td>Sensor serial number</td><td><font color='#0000FF'>&nbsp;&nbsp;42108</font></td></tr>
	   <tr><td>Sensor firmware revision</td><td><font color='#0000FF'>&nbsp;&nbsp;7.007</font></td></tr>
	   <tr><td>&nbsp;</td></tr>
	   </table>
	   </div>
	*/

	doc.Find("div p").Each(func(i int, g *goquery.Selection) {
		if strings.Contains(g.Text(), "Administrator Access") {
			if p := strings.Fields(strings.Replace(g.Text(), "'", "", -1)); len(p) > 0 {
				s["site"] = p[0]
			}
		}
	})

	doc.Find("div.box td").Each(func(i int, g *goquery.Selection) {
		//fmt.Println("--->", g.Text(), "<<<---")
		if g.Next() != nil {
			switch g.Text() {
			case "Sensor serial number":
				if p := strings.Fields(g.Next().Text()); len(p) > 0 {
					s["serial"] = p[0]
				}
			case "Data acquisition firmware revision":
				if p := strings.Fields(g.Next().Text()); len(p) > 1 {
					s["hardware"] = strings.Join(p[0:len(p)-1], " ")
					s["software"] = p[len(p)-1]
				}
			case "Sensor firmware revision":
				if p := strings.Fields(g.Next().Text()); len(p) > 0 {
					s["firmware"] = p[0]
				}
			}

		}
	})

	if _, ok := s["hardware"]; !ok {
		return nil, nil
	}

	if p := strings.Fields(s["hardware"].(string)); len(p) > 0 {
		s["model"] = "CSI Cusp " + strings.Replace(p[len(p)-1], "+", "", -1)
	}

	return &s, nil
}
