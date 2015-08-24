package dmc

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/soniah/gosnmp"
)

type Freewave struct {
	Community string
}

func (f *Freewave) Name() string {
	return "Freewave"
}

func (f *Freewave) MatchString(s string) bool {
	return regexp.MustCompile("^Freewave").MatchString(s)
}

func (f *Freewave) Groups() []ModelType {
	return []ModelType{RadioModel}
}

func (f *Freewave) Group(g ModelType) bool {
	switch g {
	case RadioModel:
		return true
	default:
		return false
	}
}

func (f *Freewave) Identify(orig string, ip net.IP, timeout time.Duration, retries int) (*State, error) {

	community := env(f.Community, "FREEWAVE_COMMUNITY", "public")

	var snmp = &gosnmp.GoSNMP{
		Target:    ip.String(),
		Port:      161,
		Community: community,
		Version:   gosnmp.Version1,
		Timeout:   timeout,
		Retries:   retries,
	}
	if err := snmp.Connect(); err != nil {
		return nil, nil
	}
	defer snmp.Conn.Close()

	oid, err := sysObjectID(snmp)
	if oid == nil || err != nil {
		return nil, err
	}

	// not an freewave ...
	if (*oid) != ".1.3.6.1.4.1.29956.2.1.1" {
		return nil, nil
	}

	s := make(State)

	// default model
	s["model"] = "Freewave Spread Spectrum Radio"

	r, err := snmp.Get([]string{
		".1.3.6.1.2.1.1.1.0",
		".1.3.6.1.2.1.1.5.0",
		".1.3.6.1.2.1.1.6.0",
		".1.3.6.1.4.1.29956.3.1.1.1.1.12.1",
	})
	if r == nil || err != nil {
		return nil, err
	}

	for _, v := range r.Variables {
		switch v.Type {
		case gosnmp.OctetString:
			switch v.Name {
			case ".1.3.6.1.2.1.1.1.0":
				l := strings.Replace((string)(v.Value.([]byte)), "Freewave Technologies", "Freewave", -1)
				l = strings.Replace(strings.Replace(strings.Replace(l, " (", ";", -1), " ;", ";", -1), ")", "", -1)
				f := strings.Split(l, ";")
				if len(f) > 0 {
					s["model"] = f[0]
					for _, j := range f[1:] {
						if k := strings.Fields(j); len(k) > 2 {
							s[k[0]] = k[2]
						}
					}
				}
			case ".1.3.6.1.2.1.1.5.0":
				if a := (string)(v.Value.([]byte)); a != "" {
					s["name"] = a
				}
			case ".1.3.6.1.2.1.1.6.0":
				if a := (string)(v.Value.([]byte)); a != "" {
					s["location"] = a
				}
			}
		case gosnmp.Integer:
			switch v.Name {
			case ".1.3.6.1.4.1.29956.3.1.1.1.1.12.1":
				if a := strconv.Itoa(v.Value.(int)); a != "" {
					s["serial"] = a
				}
			}
		}
	}

	// done ...
	return &s, nil
}
