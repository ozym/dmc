package dmc

import (
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/ozym/qdp"
)

type Quanterra struct {
}

func (q *Quanterra) Name() string {
	return "Quanterra"
}

func (q *Quanterra) MatchString(s string) bool {
	return regexp.MustCompile("^Quanterra").MatchString(s)
}

func (q *Quanterra) Groups() []ModelType {
	return []ModelType{DataloggerModel}
}

func (q *Quanterra) Group(g ModelType) bool {
	switch g {
	case DataloggerModel:
		return true
	default:
		return false
	}
}

func (q *Quanterra) Identify(orig string, ip net.IP, timeout time.Duration, retries int) (*State, error) {

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

	return nil, nil
}

func (q *Quanterra) discover(ip net.IP, port string, timeout time.Duration) (*State, error) {

	ans, err := qdp.ReadSerial(ip.String(), port, timeout)
	if ans == nil || err != nil {
		return nil, err
	}

	s := make(State)

	switch port {
	case "6330":
		s["model"] = "Quanterra Q330+"
	default:
		s["model"] = "Quanterra Q330"
	}

	s["version"] = ans.Version
	s["serial"] = ans.Serial
	s["sysver"] = ans.SysVer

	return &s, nil
}
