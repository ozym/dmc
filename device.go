package dmc

import (
	"fmt"
	"net"
	"time"
)

type Device struct {
	Name  string
	IP    net.IP
	Model string
}

func (d *Device) String() string {
	return fmt.Sprintf("%s [%s]: %s", d.Name, d.IP.String(), d.Model)
}

func (d *Device) Match(model Model) bool {
	return model.MatchString(d.Model)
}

func (d *Device) Identify(model Model, orig string, timeout time.Duration, retries int) (*State, error) {
	return model.Identify(orig, d.IP, timeout, retries)
}

func (d *Device) Discover(model Model, orig string, timeout time.Duration, retries int) *State {

	for _, g := range model.Groups() {
		for _, m := range ModelList {
			if !m.Group(g) {
				continue
			}
			if s, _ := d.Identify(m, orig, timeout, retries); s != nil {
				return s
			}
		}
	}

	return nil
}
