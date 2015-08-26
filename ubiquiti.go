package dmc

import (
	"net"
	"regexp"
	"time"

	"github.com/soniah/gosnmp"
)

type Ubiquiti struct {
	Community string
}

func (m *Ubiquiti) Name() string {
	return "Ubiquiti"
}

func (m *Ubiquiti) MatchString(s string) bool {
	return regexp.MustCompile("^Ubiquiti").MatchString(s)
}

func (m *Ubiquiti) Groups() []ModelType {
	return []ModelType{RadioModel}
}

func (m *Ubiquiti) Group(g ModelType) bool {
	switch g {
	case RadioModel:
		return true
	default:
		return false
	}
}

func (m *Ubiquiti) Identify(orig string, ip net.IP, timeout time.Duration, retries int) (*State, error) {

	community := env(m.Community, "UBIQUITY_COMMUNITY", "public")

	var snmp = &gosnmp.GoSNMP{
		Target:    ip.String(),
		Port:      161,
		Community: community,
		Version:   gosnmp.Version1,
		Timeout:   timeout,
		Retries:   retries,
	}
	if err := snmp.Connect(); err != nil {
		return nil, err
	}
	defer snmp.Conn.Close()

	oid, err := sysObjectID(snmp)
	if oid == nil || err != nil {
		return nil, err
	}

	// not an ubiquiti ...
	if (*oid) != ".1.3.6.1.4.1.10002.1" {
		return nil, nil
	}

	// optional state values ...
	r, err := snmp.Get([]string{
		".1.3.6.1.2.1.1.5.0",
		".1.3.6.1.2.1.1.6.0",
		".1.3.6.1.4.1.29956.3.1.1.1.1.12.1",
		".1.2.840.10036.3.1.2.1.3.5",
		".1.2.840.10036.3.1.2.1.4.5",
	})
	if err != nil {
		return nil, err
	}

	// state ....
	s := State{Values: make(map[string]interface{})}

	// default model
	s.Values["model"] = "Ubiquiti AirOS"

	// load in snmp derived details
	for _, v := range r.Variables {
		switch v.Type {
		case gosnmp.OctetString:
			switch v.Name {
			case ".1.2.840.10036.3.1.2.1.3.5":
				if a := (string)(v.Value.([]byte)); a != "" {
					s.Values["version"] = a
				}
			case ".1.2.840.10036.3.1.2.1.4.5":
				if a := (string)(v.Value.([]byte)); a != "" {
					s.Values["firmware"] = a
				}

			case ".1.3.6.1.2.1.1.5.0":
				if a := (string)(v.Value.([]byte)); a != "" {
					s.Values["name"] = a
				}

			case ".1.3.6.1.2.1.1.6.0":
				if a := (string)(v.Value.([]byte)); a != "" {
					s.Values["location"] = a
				}
			case ".1.3.6.1.4.1.29956.3.1.1.1.1.12.1":
				if a := (string)(v.Value.([]byte)); a != "" {
					s.Values["serial"] = a
				}
			}
		}
	}

	return &s, nil
}
