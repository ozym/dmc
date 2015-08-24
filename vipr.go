package dmc

import (
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type ViPR struct {
	Username string
	Password string
}

func (v *ViPR) Name() string {
	return "ViPR Radio"
}

func (v *ViPR) Groups() []ModelType {
	return []ModelType{RadioModel}
}

func (v *ViPR) Group(g ModelType) bool {
	switch g {
	case RadioModel:
		return true
	default:
		return false
	}
}

func (v *ViPR) MatchString(s string) bool {
	return regexp.MustCompile("^ViPR").MatchString(s)
}

func (v *ViPR) Identify(orig string, ip net.IP, timeout time.Duration, retries int) (*State, error) {

	s := make(State)

	// default model
	s["model"] = "ViPR Radio"

	cli := &http.Client{}
	pages := []string{"UStatus.html", "ViPRDiag.html", "IPSetting.html"}
	for _, p := range pages {
		r, err := v.discover(cli, "http://"+ip.String()+"/"+p)
		if err != nil {
			return nil, err
		}
		for key, value := range r {
			if t := value; key == "name" {
				s[key] = t
			}
			if t := value; key == "version" {
				s[key] = strings.Replace(t, "Dataradio ViPR ", "", -1)
			}
			if t := value; key == "firmware" {
				s[key] = t
			}
			if t := value; key == "macaddr" {
				s[key] = t
			}
		}
	}

	return &s, nil
}

func (v *ViPR) discover(cli *http.Client, url string) (map[string]string, error) {

	user := env(v.Username, "VIPR_USERNAME", "Admin")
	password := env(v.Password, "VIPR_PASSWORD", "ADMINISTRATOR")

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(user, password)

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

	// mac address is hidden in javascript and non changeable
	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		p := strings.Fields(s.Text())
		if len(p) > 0 && s.Next() != nil {
			r := strings.Fields(strings.Replace(s.Next().Text(), "\"", "", -1))
			if strings.HasPrefix(s.Text(), "Dataradio") {
				results["version"] = strings.Join(p, " ")
			}
			if strings.HasPrefix(s.Text(), "MAC") {
				if _, ok := results["macaddr"]; !ok && len(r) > 10 {
					results["macaddr"] = strings.Join(r[5:10], ":")
				}
			}
			if strings.HasPrefix(s.Text(), "RSSI") {
				if len(r) > 1 {
					results["signal"] = r[0]
				}
			}

			switch strings.Join(p, " ") {
			case "Modem Firmware Version":
				results["firmware"] = strings.Replace(strings.Replace(strings.Join(r, " "), "(", "", -1), ")", "", -1)
			case "Station Name":
				results["name"] = strings.Join(r, " ")
			case "Unit Status":
				results["status"] = strings.Join(r, " ")
			case "DC Input Voltage":
				results["supply"] = strings.Join(r, " ")
			case "Transceiver Temperature":
				results["temperature"] = strings.Join(r, " ")
			}
		}
	})

	return results, nil
}
