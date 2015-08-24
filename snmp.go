package dmc

import (
	"fmt"

	"os"
	"strconv"
	"strings"

	"github.com/soniah/gosnmp"
)

func def(v, d string) string {
	switch v {
	case "":
		return d
	default:
		return v
	}
}

func env(v, e, d string) string {
	return def(v, def(os.Getenv(e), d))
}

func index(oid string) int {
	f := strings.Split(oid, ".")

	if !(len(f) > 0) {
		return 0
	}
	if n, err := strconv.Atoi(f[len(f)-1]); err == nil {
		return n
	}

	return 0
}

func sysObjectID(snmp *gosnmp.GoSNMP) (*string, error) {
	r, err := snmp.Get([]string{
		".1.3.6.1.2.1.1.2.0",
	})
	if r == nil || err != nil {
		return nil, err
	}

	var oid *string
	for _, v := range r.Variables {
		switch v.Type {
		case gosnmp.ObjectIdentifier:
			switch v.Name {
			case ".1.3.6.1.2.1.1.2.0":
				if a := v.Value.(string); a != "" {
					oid = &a
				}
			}
		}
	}

	return oid, nil
}

type sysInterface struct {
	Name string
	Type int

	RX *uint32
	TX *uint32
}

func sysInterfaces(snmp *gosnmp.GoSNMP) (map[int]sysInterface, error) {

	w, err := snmp.WalkAll(".1.3.6.1.2.1.2.2.1")
	if err != nil {
		return nil, err
	}

	ids := make(map[int]sysInterface)
	for _, v := range w {
		switch v.Type {
		case gosnmp.OctetString:
			switch {
			case strings.HasPrefix(v.Name, ".1.3.6.1.2.1.2.2.1.2"):
				ids[index(v.Name)] = sysInterface{Name: string(v.Value.([]byte))}
			}
		}
	}
	for _, v := range w {
		i, ok := ids[index(v.Name)]
		if !ok {
			continue
		}
		switch v.Type {
		case gosnmp.Integer:
			switch {
			case strings.HasPrefix(v.Name, ".1.3.6.1.2.1.2.2.1.3"):
				i.Type = v.Value.(int)
			}
		case gosnmp.Counter32:
			switch {
			case strings.HasPrefix(v.Name, ".1.3.6.1.2.1.2.2.1.10"):
				if n := (uint32)(v.Value.(uint)); true {
					i.RX = &n
				}
			case strings.HasPrefix(v.Name, ".1.3.6.1.2.1.2.2.1.16"):
				if n := (uint32)(v.Value.(uint)); true {
					i.TX = &n
				}
			}
		}
		ids[index(v.Name)] = i
	}

	return ids, nil
}

func interfaces(snmp *gosnmp.GoSNMP, iftype int) ([]interface{}, error) {
	ids, err := sysInterfaces(snmp)
	if err != nil {
		return nil, err
	}

	var inf = make([]interface{}, 0)
	for _, k := range ids {
		if k.Type != iftype {
			continue
		}

		t := make(map[string]interface{})

		t["name"] = k.Name
		t["rx"] = k.RX
		t["tx"] = k.TX

		inf = append(inf, t)
	}

	fmt.Println("->", inf)
	return inf, nil
}
