package dmc

import (
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/soniah/gosnmp"
)

type MikroTik struct {
	Community string
}

func (m *MikroTik) Name() string {
	return "MikroTik"
}

func (m *MikroTik) MatchString(s string) bool {
	return regexp.MustCompile("^MikroTik").MatchString(s)
}

func (m *MikroTik) Groups() []ModelType {
	return []ModelType{RadioModel, RouterModel}
}

func (m *MikroTik) Group(g ModelType) bool {
	switch g {
	case RadioModel:
		return true
	case RouterModel:
		return true
	default:
		return false
	}
}

func (m *MikroTik) Identify(orig string, ip net.IP, timeout time.Duration, retries int) (*State, error) {

	community := env(m.Community, "MIKROTIK_COMMUNITY", "public")

	var snmp = &gosnmp.GoSNMP{
		Target:    ip.String(),
		Port:      161,
		Community: community,
		Version:   gosnmp.Version2c,
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

	// not a mikrotik ...
	if (*oid) != ".1.3.6.1.4.1.14988.1" {
		return nil, nil
	}

	// default model ...
	label := "MikroTik RouterOS"

	// check interfaces ...
	w, err := snmp.WalkAll(".1.3.6.1.2.1.2.2.1.3")
	if err != nil {
		return nil, err
	}

	// check for wlan interfaces ...
	for _, v := range w {
		switch {
		case !(v.Type == gosnmp.Integer):
		case !(strings.HasPrefix(v.Name, ".1.3.6.1.2.1.2.2.1.3")):
		case v.Value.(int) == 71:
			label = "MikroTik Routerboard"
		}
	}

	// state ....
	s := make(State)

	s["model"] = label

	// status ...
	oids := []string{
		".1.3.6.1.2.1.1.1.0",
		".1.3.6.1.2.1.1.5.0",
		".1.3.6.1.2.1.1.6.0",
		".1.3.6.1.4.1.14988.1.1.4.1.0",
		".1.3.6.1.4.1.14988.1.1.4.4.0",
		".1.3.6.1.4.1.14988.1.1.7.3.0",
		".1.3.6.1.4.1.14988.1.1.7.4.0",
	}

	// check for optional status values
	if r, err := snmp.Get(oids); err == nil {
		for _, v := range r.Variables {
			switch v.Type {
			case gosnmp.OctetString:
				switch v.Name {
				case ".1.3.6.1.2.1.1.1.0":
					if f := strings.Fields((string)(v.Value.([]byte))); len(f) > 0 {
						if a := strings.Join(f[1:], " "); a != "" {
							s["version"] = a
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
				case ".1.3.6.1.4.1.14988.1.1.7.3.0":
					if a := (string)(v.Value.([]byte)); a != "" {
						s["serial"] = a
					}
				case ".1.3.6.1.4.1.14988.1.1.7.4.0":
					if a := (string)(v.Value.([]byte)); a != "" {
						s["firmware"] = a
					}
				case ".1.3.6.1.4.1.14988.1.1.4.4.0":
					if a := (string)(v.Value.([]byte)); a != "" {
						s["software"] = a
					}
				case ".1.3.6.1.4.1.14988.1.1.4.1.0":
					if a := (string)(v.Value.([]byte)); a != "" {
						s["license"] = a
					}
				}
			}
		}
	}

	return &s, nil
}
